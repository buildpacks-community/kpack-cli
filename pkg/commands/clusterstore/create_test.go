// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"

	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/clusterstore/fakes"
	storecmds "github.com/pivotal/build-service-cli/pkg/commands/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterStoreCreateCommand(t *testing.T) {
	spec.Run(t, "TestClusterStoreCreateCommand", testClusterStoreCreateCommand)
}

func testClusterStoreCreateCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		buildpackage1 = "some/newbp"
		uploadedBp1   = "some-registry.io/some-repo/newbp@sha256:123newbp"
		buildpackage2 = "bpfromcnb.cnb"
		uploadedBp2   = "some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb"

		fakeBuildpackageUploader = fakes.FakeBuildpackageUploader{
			buildpackage1: uploadedBp1,
			buildpackage2: uploadedBp2,
		}

		factory = &clusterstore.Factory{
			Uploader: fakeBuildpackageUploader,
		}

		config = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{
				"canonical.repository":                "some-registry.io/some-repo",
				"canonical.repository.serviceaccount": "some-serviceaccount",
			},
		}

		expectedStore = &v1alpha1.ClusterStore{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterStoreKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-store",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-store","creationTimestamp":null},"spec":{"sources":[{"image":"some-registry.io/some-repo/newbp@sha256:123newbp"},{"image":"some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb"}]},"status":{}}`,
				},
			},
			Spec: v1alpha1.ClusterStoreSpec{
				Sources: []v1alpha1.StoreImage{
					{Image: uploadedBp1},
					{Image: uploadedBp2},
				},
			},
		}
	)

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return storecmds.NewCreateCommand(clientSetProvider, factory)
	}

	it("creates a cluster store", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			Args: []string{
				expectedStore.Name,
				"--buildpackage", buildpackage1,
				"-b", buildpackage2,
				"--registry-ca-cert-path", "some-cert-path",
				"--registry-verify-certs",
			},
			ExpectedOutput: "Creating Cluster Store...\n\"test-store\" created\n",
			ExpectCreates: []runtime.Object{
				expectedStore,
			},
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("fails when kp-config configmap is not found", func() {
		testhelpers.CommandTest{
			Args: []string{
				expectedStore.Name,
				"--buildpackage", buildpackage1,
				"-b", buildpackage2,
			},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical repository: configmaps "kp-config" not found
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("fails when canonical.repository key is not found in kp-config configmap", func() {
		badConfig := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{},
		}

		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				badConfig,
			},
			Args: []string{
				expectedStore.Name,
				"--buildpackage", buildpackage1,
				"-b", buildpackage2,
			},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical repository: key "canonical.repository" not found in configmap "kp-config"
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("fails when a buildpackage is not provided", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			Args: []string{
				expectedStore.Name,
			},
			ExpectErr:      true,
			ExpectedOutput: "Creating Cluster Store...\nError: At least one buildpackage must be provided\n",
		}.TestK8sAndKpack(t, cmdFunc)
	})
}

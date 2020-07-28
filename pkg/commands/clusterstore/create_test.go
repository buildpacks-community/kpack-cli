// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

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
		uploadedBp1   = "some/path/newbp@sha256:123newbp"
		buildpackage2 = "bpfromcnb.cnb"
		uploadedBp2   = "some/path/bpfromcnb@sha256:123imagefromcnb"

		fakeBuildpackageUploader = fakes.FakeBuildpackageUploader{
			buildpackage1: uploadedBp1,
			buildpackage2: uploadedBp2,
		}

		factory = &clusterstore.Factory{
			Uploader: fakeBuildpackageUploader,
		}

		expectedStore = &expv1alpha1.ClusterStore{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.ClusterStoreKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-store",
				Annotations: map[string]string{
					"buildservice.pivotal.io/defaultRepository":        "some-registry.io/some-repo",
					"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"ClusterStore","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"test-store","creationTimestamp":null,"annotations":{"buildservice.pivotal.io/defaultRepository":"some-registry.io/some-repo"}},"spec":{"sources":[{"image":"some/path/newbp@sha256:123newbp"},{"image":"some/path/bpfromcnb@sha256:123imagefromcnb"}]},"status":{}}`,
				},
			},
			Spec: expv1alpha1.ClusterStoreSpec{
				Sources: []expv1alpha1.StoreImage{
					{Image: uploadedBp1},
					{Image: uploadedBp2},
				},
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return storecmds.NewCreateCommand(clientSetProvider, factory)
	}

	it("creates a cluster store", func() {
		testhelpers.CommandTest{
			Args: []string{
				expectedStore.Name,
				buildpackage1,
				buildpackage2,
				"--default-repository", "some-registry.io/some-repo",
			},
			ExpectedOutput: "Uploading to 'some-registry.io/some-repo'...\n\"test-store\" created\n",
			ExpectCreates: []runtime.Object{
				expectedStore,
			},
		}.TestKpack(t, cmdFunc)
	})

	it("fails if a buildpackage is not provided", func() {
		testhelpers.CommandTest{
			Args: []string{
				expectedStore.Name,
				"--default-repository", "some-registry.io/some-repo",
			},
			ExpectErr:      true,
			ExpectedOutput: "Error: At least one buildpackage must be provided\n",
		}.TestKpack(t, cmdFunc)
	})

	it("validates the default repository", func() {
		testhelpers.CommandTest{
			Args: []string{
				expectedStore.Name,
				buildpackage1,
				"--default-repository", "bad-repo@",
			},
			ExpectErr:      true,
			ExpectedOutput: "Error: could not parse reference: bad-repo@\n",
		}.TestKpack(t, cmdFunc)
	})
}

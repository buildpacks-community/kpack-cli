// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack_test

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

	clusterstackfakes "github.com/pivotal/build-service-cli/pkg/clusterstack/fakes"
	clusterstackcmds "github.com/pivotal/build-service-cli/pkg/commands/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestCreateCommand(t *testing.T) {
	spec.Run(t, "TestCreateCommand", testCreateCommand)
}

func testCreateCommand(t *testing.T, when spec.G, it spec.S) {
	fakeUploader := &clusterstackfakes.FakeStackUploader{
		Images: map[string]string{
			"some-build-image": "some-uploaded-build-image",
			"some-run-image":   "some-uploaded-run-image",
		},
		StackID: "some-stack-id",
	}

	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kp-config",
			Namespace: "kpack",
		},
		Data: map[string]string{
			"canonical.repository":                "some-registry.io/some-repo",
			"canonical.repository.serviceaccount": "some-serviceaccount",
		},
	}

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return clusterstackcmds.NewCreateCommand(clientSetProvider, fakeUploader)
	}

	expectedStack := &v1alpha1.ClusterStack{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ClusterStackKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-stack",
		},
		Spec: v1alpha1.ClusterStackSpec{
			Id: "some-stack-id",
			BuildImage: v1alpha1.ClusterStackSpecImage{
				Image: "some-uploaded-build-image",
			},
			RunImage: v1alpha1.ClusterStackSpecImage{
				Image: "some-uploaded-run-image",
			},
		},
	}

	it("creates a stack", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			Args: []string{
				"some-stack",
				"--build-image", "some-build-image",
				"--run-image", "some-run-image",
				"--registry-ca-cert-path", "some-cert-path",
				"--registry-verify-certs",
			},
			ExpectedOutput: `Creating ClusterStack...
Uploading to 'some-registry.io/some-repo'...
ClusterStack "some-stack" created
`,
			ExpectCreates: []runtime.Object{
				expectedStack,
			},
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("fails when kp-config configmap is not found", func() {
		testhelpers.CommandTest{
			Args: []string{
				"some-stack",
				"--build-image", "some-build-image",
				"--run-image", "some-run-image",
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
				"some-stack",
				"--build-image", "some-build-image",
				"--run-image", "some-run-image",
			},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical repository: key "canonical.repository" not found in configmap "kp-config"
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	when("output flag is used", func() {
		it("can output in yaml format", func() {
			const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: some-stack
spec:
  buildImage:
    image: some-uploaded-build-image
  id: some-stack-id
  runImage:
    image: some-uploaded-run-image
status:
  buildImage: {}
  runImage: {}
`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-build-image",
					"--run-image", "some-run-image",
					"--output", "yaml",
				},
				ExpectedOutput: resourceYAML,
				ExpectedErrorOutput: `Creating ClusterStack...
Uploading to 'some-registry.io/some-repo'...
`,
				ExpectCreates: []runtime.Object{
					expectedStack,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("can output in json format", func() {
			const resourceJSON = `{
    "kind": "ClusterStack",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "some-stack",
        "creationTimestamp": null
    },
    "spec": {
        "id": "some-stack-id",
        "buildImage": {
            "image": "some-uploaded-build-image"
        },
        "runImage": {
            "image": "some-uploaded-run-image"
        }
    },
    "status": {
        "buildImage": {},
        "runImage": {}
    }
}
`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-build-image",
					"--run-image", "some-run-image",
					"--output", "json",
				},
				ExpectedOutput: resourceJSON,
				ExpectedErrorOutput: `Creating ClusterStack...
Uploading to 'some-registry.io/some-repo'...
`,
				ExpectCreates: []runtime.Object{
					expectedStack,
				},
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})

	when("dry-run flag is used", func() {
		it("does not create a clusterstack and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-build-image",
					"--run-image", "some-run-image",
					"--dry-run",
				},
				ExpectedOutput: `Creating ClusterStack... (dry run)
ClusterStack "some-stack" created (dry run)
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("does not create a clusterstack and prints the resource output", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: some-stack
spec:
  buildImage:
    image: some-uploaded-build-image
  id: some-stack-id
  runImage:
    image: some-uploaded-run-image
status:
  buildImage: {}
  runImage: {}
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					Args: []string{
						"some-stack",
						"--build-image", "some-build-image",
						"--run-image", "some-run-image",
						"--dry-run",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Creating ClusterStack... (dry run)
Uploading to 'some-registry.io/some-repo'...
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})
}

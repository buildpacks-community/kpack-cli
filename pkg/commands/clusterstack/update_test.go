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
	clientgotesting "k8s.io/client-go/testing"

	clusterstackfakes "github.com/pivotal/build-service-cli/pkg/clusterstack/fakes"
	clusterstackcmds "github.com/pivotal/build-service-cli/pkg/commands/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestUpdateCommand(t *testing.T) {
	spec.Run(t, "TestUpdateCommand", testUpdateCommand)
}

func testUpdateCommand(t *testing.T, when spec.G, it spec.S) {
	fakeUploader := &clusterstackfakes.FakeStackUploader{
		Images: map[string]string{
			"some-old-build-image": "some-old-build-image@some-old-digest",
			"some-old-run-image":   "some-old-run-image@some-old-digest",
			"some-new-build-image": "some-new-uploaded-build-image@some-digest",
			"some-new-run-image":   "some-new-uploaded-run-image@some-digest",
		},
		StackID: "some-new-id",
	}

	stack := &v1alpha1.ClusterStack{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-stack",
		},
		Spec: v1alpha1.ClusterStackSpec{
			Id: "some-old-id",
			BuildImage: v1alpha1.ClusterStackSpecImage{
				Image: "some-old-build-image",
			},
			RunImage: v1alpha1.ClusterStackSpecImage{
				Image: "some-old-run-image",
			},
		},
		Status: v1alpha1.ClusterStackStatus{
			ResolvedClusterStack: v1alpha1.ResolvedClusterStack{
				Id: "some-old-id",
				BuildImage: v1alpha1.ClusterStackStatusImage{
					LatestImage: "some-old-build-image@some-old-digest",
					Image:       "some-old-build-image@some-old-digest",
				},
				RunImage: v1alpha1.ClusterStackStatusImage{
					LatestImage: "some-old-run-image@some-old-digest",
					Image:       "some-old-run-image@some-old-digest",
				},
			},
		},
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
		return clusterstackcmds.NewUpdateCommand(clientSetProvider, fakeUploader)
	}

	it("updates the stack id, run image, and build image", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			KpackObjects: []runtime.Object{
				stack,
			},
			Args: []string{"some-stack",
				"--build-image", "some-new-build-image",
				"--run-image", "some-new-run-image",
				"--registry-ca-cert-path", "some-cert-path",
				"--registry-verify-certs",
			},
			ExpectErr: false,
			ExpectUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &v1alpha1.ClusterStack{
						ObjectMeta: stack.ObjectMeta,
						Spec: v1alpha1.ClusterStackSpec{
							Id: "some-new-id",
							BuildImage: v1alpha1.ClusterStackSpecImage{
								Image: "some-new-uploaded-build-image@some-digest",
							},
							RunImage: v1alpha1.ClusterStackSpecImage{
								Image: "some-new-uploaded-run-image@some-digest",
							},
						},
						Status: stack.Status,
					},
				},
			},
			ExpectedOutput: `Updating ClusterStack...
Uploading to 'some-registry.io/some-repo'...
ClusterStack "some-stack" updated
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("does not add stack images with the same digest", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			KpackObjects: []runtime.Object{
				stack,
			},
			Args:      []string{"some-stack", "--build-image", "some-old-build-image", "--run-image", "some-old-run-image"},
			ExpectErr: false,
			ExpectedOutput: `Updating ClusterStack...
Uploading to 'some-registry.io/some-repo'...
Build and Run images already exist in stack
ClusterStack "some-stack" updated (no change)
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("returns error when kp-config configmap is not found", func() {
		testhelpers.CommandTest{
			KpackObjects: []runtime.Object{
				stack,
			},
			Args:      []string{"some-stack", "--build-image", "some-new-build-image", "--run-image", "some-new-run-image"},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical repository: configmaps "kp-config" not found
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("returns error when canonical.repository key is not found in kp-config configmap", func() {
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
			KpackObjects: []runtime.Object{
				stack,
			},
			Args:      []string{"some-stack", "--build-image", "some-new-build-image", "--run-image", "some-new-run-image"},
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
    image: some-new-uploaded-build-image@some-digest
  id: some-new-id
  runImage:
    image: some-new-uploaded-run-image@some-digest
status:
  buildImage:
    image: some-old-build-image@some-old-digest
    latestImage: some-old-build-image@some-old-digest
  id: some-old-id
  runImage:
    image: some-old-run-image@some-old-digest
    latestImage: some-old-run-image@some-old-digest
`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				KpackObjects: []runtime.Object{
					stack,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-new-build-image",
					"--run-image", "some-new-run-image",
					"--output", "yaml",
				},
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: &v1alpha1.ClusterStack{
							ObjectMeta: stack.ObjectMeta,
							Spec: v1alpha1.ClusterStackSpec{
								Id: "some-new-id",
								BuildImage: v1alpha1.ClusterStackSpecImage{
									Image: "some-new-uploaded-build-image@some-digest",
								},
								RunImage: v1alpha1.ClusterStackSpecImage{
									Image: "some-new-uploaded-run-image@some-digest",
								},
							},
							Status: stack.Status,
						},
					},
				},
				ExpectedOutput: resourceYAML,
				ExpectedErrorOutput: `Updating ClusterStack...
Uploading to 'some-registry.io/some-repo'...
`,
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
        "id": "some-new-id",
        "buildImage": {
            "image": "some-new-uploaded-build-image@some-digest"
        },
        "runImage": {
            "image": "some-new-uploaded-run-image@some-digest"
        }
    },
    "status": {
        "id": "some-old-id",
        "buildImage": {
            "latestImage": "some-old-build-image@some-old-digest",
            "image": "some-old-build-image@some-old-digest"
        },
        "runImage": {
            "latestImage": "some-old-run-image@some-old-digest",
            "image": "some-old-run-image@some-old-digest"
        }
    }
}
`

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				KpackObjects: []runtime.Object{
					stack,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-new-build-image",
					"--run-image", "some-new-run-image",
					"--output", "json",
				},
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: &v1alpha1.ClusterStack{
							ObjectMeta: stack.ObjectMeta,
							Spec: v1alpha1.ClusterStackSpec{
								Id: "some-new-id",
								BuildImage: v1alpha1.ClusterStackSpecImage{
									Image: "some-new-uploaded-build-image@some-digest",
								},
								RunImage: v1alpha1.ClusterStackSpecImage{
									Image: "some-new-uploaded-run-image@some-digest",
								},
							},
							Status: stack.Status,
						},
					},
				},
				ExpectedOutput: resourceJSON,
				ExpectedErrorOutput: `Updating ClusterStack...
Uploading to 'some-registry.io/some-repo'...
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("there are no changes in the update", func() {
			it("can output original resource in requested format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: some-stack
spec:
  buildImage:
    image: some-old-build-image
  id: some-old-id
  runImage:
    image: some-old-run-image
status:
  buildImage:
    image: some-old-build-image@some-old-digest
    latestImage: some-old-build-image@some-old-digest
  id: some-old-id
  runImage:
    image: some-old-run-image@some-old-digest
    latestImage: some-old-run-image@some-old-digest
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						stack,
					},
					Args: []string{
						"some-stack",
						"--build-image", "some-old-build-image",
						"--run-image", "some-old-run-image",
						"--output", "yaml",
					},
					ExpectedErrorOutput: `Updating ClusterStack...
Uploading to 'some-registry.io/some-repo'...
Build and Run images already exist in stack
`,
					ExpectedOutput: resourceYAML,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})

	when("dry-run flag is used", func() {
		it("does not update the clusterstack and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				KpackObjects: []runtime.Object{
					stack,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-new-build-image",
					"--run-image", "some-new-run-image",
					"--dry-run",
				},
				ExpectedOutput: `Updating ClusterStack... (dry run)
ClusterStack "some-stack" updated (dry run)
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("there are no changes in the update", func() {
			it("does not create a clusterstack and informs of no change", func() {
				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						stack,
					},
					Args: []string{
						"some-stack",
						"--build-image", "some-old-build-image",
						"--run-image", "some-old-run-image",
						"--dry-run",
					},
					ExpectedOutput: `Updating ClusterStack... (dry run)
Build and Run images already exist in stack
ClusterStack "some-stack" updated (dry run)
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		when("output flag is used", func() {
			it("does not update the clusterstack and prints the resource output", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: some-stack
spec:
  buildImage:
    image: some-new-uploaded-build-image@some-digest
  id: some-new-id
  runImage:
    image: some-new-uploaded-run-image@some-digest
status:
  buildImage:
    image: some-old-build-image@some-old-digest
    latestImage: some-old-build-image@some-old-digest
  id: some-old-id
  runImage:
    image: some-old-run-image@some-old-digest
    latestImage: some-old-run-image@some-old-digest
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						stack,
					},
					Args: []string{
						"some-stack",
						"--build-image", "some-new-build-image",
						"--run-image", "some-new-run-image",
						"--dry-run",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Updating ClusterStack... (dry run)
Uploading to 'some-registry.io/some-repo'...
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})
	})
}

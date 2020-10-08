// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack_test

import (
	"testing"

	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"

	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	clusterstackcmds "github.com/pivotal/build-service-cli/pkg/commands/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/image/fakes"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestSaveCommand(t *testing.T) {
	spec.Run(t, "TestSaveCommand", testSaveCommand)
}

func testSaveCommand(t *testing.T, when spec.G, it spec.S) {
	fetcher := &fakes.Fetcher{}

	buildImage, buildImageId, runImage, runImageId := makeStackImages(t, "some-stack-id")
	fetcher.AddImage("some-build-image", buildImage)
	fetcher.AddImage("some-run-image", runImage)

	newBuildImage, newBuildImageId, newRunImage, newRunImageId := makeStackImages(t, "some-new-id")
	fetcher.AddImage("some-new-build-image", newBuildImage)
	fetcher.AddImage("some-new-run-image", newRunImage)

	relocator := &fakes.Relocator{}

	stackFactory := &clusterstack.Factory{
		Fetcher:   fetcher,
		Relocator: relocator,
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
		return clusterstackcmds.NewSaveCommand(clientSetProvider, stackFactory)
	}

	when("creating", func() {
		it("creates a stack when it does not exist", func() {
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
						Image: "some-registry.io/some-repo/build@" + buildImageId,
					},
					RunImage: v1alpha1.ClusterStackSpecImage{
						Image: "some-registry.io/some-repo/run@" + runImageId,
					},
				},
			}

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
"some-stack" created
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

		it("validates build stack ID is equal to run stack ID", func() {
			_, _, runImage, _ := makeStackImages(t, "some-other-stack-id")

			fetcher.AddImage("some-other-run-image", runImage)

			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-build-image",
					"--run-image", "some-other-run-image",
				},
				ExpectErr: true,
				ExpectedOutput: `Creating ClusterStack...
Error: build stack 'some-stack-id' does not match run stack 'some-other-stack-id'
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
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
						Image: "some-registry.io/some-repo/build@" + buildImageId,
					},
					RunImage: v1alpha1.ClusterStackSpecImage{
						Image: "some-registry.io/some-repo/run@" + runImageId,
					},
				},
			}

			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: some-stack
spec:
  buildImage:
    image: some-registry.io/some-repo/build@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
  id: some-stack-id
  runImage:
    image: some-registry.io/some-repo/run@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
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
            "image": "some-registry.io/some-repo/build@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d"
        },
        "runImage": {
            "image": "some-registry.io/some-repo/run@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d"
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
Uploading to 'some-registry.io/some-repo'...
"some-stack" created (dry run)
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
    image: some-registry.io/some-repo/build@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
  id: some-stack-id
  runImage:
    image: some-registry.io/some-repo/run@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
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
	})

	when("updating", func() {
		stack := &v1alpha1.ClusterStack{
			ObjectMeta: metav1.ObjectMeta{
				Name: "some-stack",
			},
			Spec: v1alpha1.ClusterStackSpec{
				Id: "some-stack-id",
				BuildImage: v1alpha1.ClusterStackSpecImage{
					Image: "some-build-image",
				},
				RunImage: v1alpha1.ClusterStackSpecImage{
					Image: "some-run-image",
				},
			},
			Status: v1alpha1.ClusterStackStatus{
				ResolvedClusterStack: v1alpha1.ResolvedClusterStack{
					Id: "some-old-id",
					BuildImage: v1alpha1.ClusterStackStatusImage{
						LatestImage: "some-registry.io/old-repo/build@" + buildImageId,
						Image:       "some-old-build-image",
					},
					RunImage: v1alpha1.ClusterStackStatusImage{
						LatestImage: "some-registry.io/old-repo/run@" + runImageId,
						Image:       "some-old-run-image",
					},
				},
			},
		}

		it("updates the stack id, run image, and build image when the clusterstack does exist", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				KpackObjects: []runtime.Object{
					stack,
				},
				Args:      []string{"some-stack", "--build-image", "some-new-build-image", "--run-image", "some-new-run-image"},
				ExpectErr: false,
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: &v1alpha1.ClusterStack{
							ObjectMeta: stack.ObjectMeta,
							Spec: v1alpha1.ClusterStackSpec{
								Id: "some-new-id",
								BuildImage: v1alpha1.ClusterStackSpecImage{
									Image: "some-registry.io/some-repo/build@" + newBuildImageId,
								},
								RunImage: v1alpha1.ClusterStackSpecImage{
									Image: "some-registry.io/some-repo/run@" + newRunImageId,
								},
							},
							Status: stack.Status,
						},
					},
				},
				ExpectedOutput: `Updating ClusterStack...
Uploading to 'some-registry.io/some-repo'...
"some-stack" updated
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("does not update clusterstack when the images already exist", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				KpackObjects: []runtime.Object{
					stack,
				},
				Args: []string{
					"some-stack",
					"--build-image", "some-build-image",
					"--run-image", "some-run-image",
				},
				ExpectedOutput: `Updating ClusterStack...
Uploading to 'some-registry.io/some-repo'...
Build and Run images already exist in stack
"some-stack" updated (no change)
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
    image: some-registry.io/some-repo/build@sha256:2dc4df5d1ff625346ecdb753ebc2aa2d18fc02027fd2459a1ff59b81bde904e7
  id: some-new-id
  runImage:
    image: some-registry.io/some-repo/run@sha256:2dc4df5d1ff625346ecdb753ebc2aa2d18fc02027fd2459a1ff59b81bde904e7
status:
  buildImage:
    image: some-old-build-image
    latestImage: some-registry.io/old-repo/build@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
  id: some-old-id
  runImage:
    image: some-old-run-image
    latestImage: some-registry.io/old-repo/run@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
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
										Image: "some-registry.io/some-repo/build@" + newBuildImageId,
									},
									RunImage: v1alpha1.ClusterStackSpecImage{
										Image: "some-registry.io/some-repo/run@" + newRunImageId,
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
            "image": "some-registry.io/some-repo/build@sha256:2dc4df5d1ff625346ecdb753ebc2aa2d18fc02027fd2459a1ff59b81bde904e7"
        },
        "runImage": {
            "image": "some-registry.io/some-repo/run@sha256:2dc4df5d1ff625346ecdb753ebc2aa2d18fc02027fd2459a1ff59b81bde904e7"
        }
    },
    "status": {
        "id": "some-old-id",
        "buildImage": {
            "latestImage": "some-registry.io/old-repo/build@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d",
            "image": "some-old-build-image"
        },
        "runImage": {
            "latestImage": "some-registry.io/old-repo/run@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d",
            "image": "some-old-run-image"
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
										Image: "some-registry.io/some-repo/build@" + newBuildImageId,
									},
									RunImage: v1alpha1.ClusterStackSpecImage{
										Image: "some-registry.io/some-repo/run@" + newRunImageId,
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
    image: some-build-image
  id: some-stack-id
  runImage:
    image: some-run-image
status:
  buildImage:
    image: some-old-build-image
    latestImage: some-registry.io/old-repo/build@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
  id: some-old-id
  runImage:
    image: some-old-run-image
    latestImage: some-registry.io/old-repo/run@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
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
							"--build-image", "some-build-image",
							"--run-image", "some-run-image",
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
Uploading to 'some-registry.io/some-repo'...
"some-stack" updated (dry run)
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
							"--build-image", "some-build-image",
							"--run-image", "some-run-image",
							"--dry-run",
						},
						ExpectedOutput: `Updating ClusterStack... (dry run)
Uploading to 'some-registry.io/some-repo'...
Build and Run images already exist in stack
"some-stack" updated (no change)
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
    image: some-registry.io/some-repo/build@sha256:2dc4df5d1ff625346ecdb753ebc2aa2d18fc02027fd2459a1ff59b81bde904e7
  id: some-new-id
  runImage:
    image: some-registry.io/some-repo/run@sha256:2dc4df5d1ff625346ecdb753ebc2aa2d18fc02027fd2459a1ff59b81bde904e7
status:
  buildImage:
    image: some-old-build-image
    latestImage: some-registry.io/old-repo/build@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
  id: some-old-id
  runImage:
    image: some-old-run-image
    latestImage: some-registry.io/old-repo/run@sha256:9dc5608d0f7f31ecd4cd26c00ec56629180dc29bba3423e26fc87317e1c2846d
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
	})
}

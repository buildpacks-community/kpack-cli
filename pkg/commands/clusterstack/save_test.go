// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	k8sfakes "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterstack"
	commandsfakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	registryfakes "github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestSaveCommand(t *testing.T) {
	spec.Run(t, "TestSaveCommand", testSaveCommand)
}

func testSaveCommand(t *testing.T, when spec.G, it spec.S) {
	fakeFetcher := &registryfakes.Fetcher{}
	fakeRelocator := &registryfakes.Relocator{}
	fakeRegistryUtilProvider := &registryfakes.UtilProvider{
		FakeFetcher: fakeFetcher,
	}

	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kp-config",
			Namespace: "kpack",
		},
		Data: map[string]string{
			"canonical.repository":                "canonical-registry.io/canonical-repo",
			"canonical.repository.serviceaccount": "some-serviceaccount",
		},
	}

	fakeWaiter := &commandsfakes.FakeWaiter{}

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return clusterstack.NewSaveCommand(clientSetProvider, fakeRegistryUtilProvider, func(dynamic.Interface) commands.ResourceWaiter {
			return fakeWaiter
		})
	}

	when("creating", func() {
		fakeFetcher.AddStackImages(registryfakes.StackInfo{
			StackID: "stack-id",
			BuildImg: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/some-build-image",
				Digest: "build-image-digest",
			},
			RunImg: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/some-run-image",
				Digest: "run-image-digest",
			},
		})

		expectedStack := &v1alpha2.ClusterStack{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.ClusterStackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "stack-name",
			},
			Spec: v1alpha2.ClusterStackSpec{
				Id: "stack-id",
				BuildImage: v1alpha2.ClusterStackSpecImage{
					Image: "canonical-registry.io/canonical-repo/build@sha256:build-image-digest",
				},
				RunImage: v1alpha2.ClusterStackSpecImage{
					Image: "canonical-registry.io/canonical-repo/run@sha256:run-image-digest",
				},
			},
		}

		it("creates a stack", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
				},
				Args: []string{
					"stack-name",
					"--build-image", "some-registry.io/repo/some-build-image",
					"--run-image", "some-registry.io/repo/some-run-image",
					"--registry-ca-cert-path", "some-cert-path",
					"--registry-verify-certs",
				},
				ExpectedOutput: `Creating ClusterStack...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
ClusterStack "stack-name" created
`,
				ExpectCreates: []runtime.Object{
					expectedStack,
				},
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
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
				Objects: []runtime.Object{
					badConfig,
				},
				Args: []string{
					"stack-name",
					"--build-image", "some-registry.io/repo/some-build-image",
					"--run-image", "some-registry.io/repo/some-run-image",
				},
				ExpectErr:           true,
				ExpectedOutput:      "Creating ClusterStack...\n",
				ExpectedErrorOutput: "Error: failed to get canonical repository: use \"kp config canonical-repository\" to set\n",
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
status:
  buildImage: {}
  runImage: {}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
					},
					Args: []string{
						"stack-name",
						"--build-image", "some-registry.io/repo/some-build-image",
						"--run-image", "some-registry.io/repo/some-run-image",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Creating ClusterStack...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
`,
					ExpectCreates: []runtime.Object{
						expectedStack,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				const resourceJSON = `{
    "kind": "ClusterStack",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "stack-name",
        "creationTimestamp": null
    },
    "spec": {
        "id": "stack-id",
        "buildImage": {
            "image": "canonical-registry.io/canonical-repo/build@sha256:build-image-digest"
        },
        "runImage": {
            "image": "canonical-registry.io/canonical-repo/run@sha256:run-image-digest"
        }
    },
    "status": {
        "buildImage": {},
        "runImage": {}
    }
}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
					},
					Args: []string{
						"stack-name",
						"--build-image", "some-registry.io/repo/some-build-image",
						"--run-image", "some-registry.io/repo/some-run-image",
						"--output", "json",
					},
					ExpectedOutput: resourceJSON,
					ExpectedErrorOutput: `Creating ClusterStack...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
`,
					ExpectCreates: []runtime.Object{
						expectedStack,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		when("dry-run flag is used", func() {
			fakeRelocator.SetSkip(true)

			it("does not create a clusterstack and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
					},
					Args: []string{
						"stack-name",
						"--build-image", "some-registry.io/repo/some-build-image",
						"--run-image", "some-registry.io/repo/some-run-image",
						"--dry-run",
					},
					ExpectedOutput: `Creating ClusterStack... (dry run)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Skipping 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
ClusterStack "stack-name" created (dry run)
`,
				}.TestK8sAndKpack(t, cmdFunc)
				require.Len(t, fakeWaiter.WaitCalls, 0)
			})

			when("output flag is used", func() {
				it("does not create a clusterstack and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
status:
  buildImage: {}
  runImage: {}
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							config,
						},
						Args: []string{
							"stack-name",
							"--build-image", "some-registry.io/repo/some-build-image",
							"--run-image", "some-registry.io/repo/some-run-image",
							"--dry-run",
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
						ExpectedErrorOutput: `Creating ClusterStack... (dry run)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Skipping 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})

		when("dry-run-with-image-upload flag is used", func() {
			it("does not create a clusterstack and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
					},
					Args: []string{
						"stack-name",
						"--build-image", "some-registry.io/repo/some-build-image",
						"--run-image", "some-registry.io/repo/some-run-image",
						"--dry-run-with-image-upload",
					},
					ExpectedOutput: `Creating ClusterStack... (dry run with image upload)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
ClusterStack "stack-name" created (dry run with image upload)
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("output flag is used", func() {
				it("does not create a clusterstack and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
status:
  buildImage: {}
  runImage: {}
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							config,
						},
						Args: []string{
							"stack-name",
							"--build-image", "some-registry.io/repo/some-build-image",
							"--run-image", "some-registry.io/repo/some-run-image",
							"--dry-run-with-image-upload",
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
						ExpectedErrorOutput: `Creating ClusterStack... (dry run with image upload)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})
	})

	when("updating", func() {
		fakeFetcher.AddStackImages(registryfakes.StackInfo{
			StackID: "stack-id",
			BuildImg: registryfakes.ImageInfo{
				Ref:    "canonical-registry.io/canonical-repo/build@sha256:build-image-digest",
				Digest: "build-image-digest",
			},
			RunImg: registryfakes.ImageInfo{
				Ref:    "canonical-registry.io/canonical-repo/run@sha256:run-image-digest",
				Digest: "run-image-digest",
			},
		})

		fakeFetcher.AddStackImages(registryfakes.StackInfo{
			StackID: "stack-id",
			BuildImg: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/new-build",
				Digest: "new-build-image-digest",
			},
			RunImg: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/new-run",
				Digest: "new-run-image-digest",
			},
		})

		stack := &v1alpha2.ClusterStack{
			ObjectMeta: metav1.ObjectMeta{
				Name: "stack-name",
			},
			Spec: v1alpha2.ClusterStackSpec{
				Id: "stack-id",
				BuildImage: v1alpha2.ClusterStackSpecImage{
					Image: "canonical-registry.io/canonical-repo/build@sha256:build-image-digest",
				},
				RunImage: v1alpha2.ClusterStackSpecImage{
					Image: "canonical-registry.io/canonical-repo/run@sha256:run-image-digest",
				},
			},
			Status: v1alpha2.ClusterStackStatus{
				ResolvedClusterStack: v1alpha2.ResolvedClusterStack{
					Id: "stack-id",
					BuildImage: v1alpha2.ClusterStackStatusImage{
						LatestImage: "canonical-registry.io/canonical-repo/build@sha256:build-image-digest",
						Image:       "canonical-registry.io/canonical-repo/build@sha256:build-image-digest",
					},
					RunImage: v1alpha2.ClusterStackStatusImage{
						LatestImage: "canonical-registry.io/canonical-repo/run@sha256:run-image-digest",
						Image:       "canonical-registry.io/canonical-repo/run@sha256:run-image-digest",
					},
				},
			},
		}

		cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
			return clusterstack.NewUpdateCommand(clientSetProvider, fakeRegistryUtilProvider, func(dynamic.Interface) commands.ResourceWaiter {
				return fakeWaiter
			})
		}

		it("updates the stack id, run image, and build image", func() {
			expectedStack := &v1alpha2.ClusterStack{
				ObjectMeta: stack.ObjectMeta,
				Spec: v1alpha2.ClusterStackSpec{
					Id: "stack-id",
					BuildImage: v1alpha2.ClusterStackSpecImage{
						Image: "canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest",
					},
					RunImage: v1alpha2.ClusterStackSpecImage{
						Image: "canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest",
					},
				},
				Status: stack.Status,
			}
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
					stack,
				},
				Args: []string{
					"stack-name",
					"--build-image", "some-registry.io/repo/new-build",
					"--run-image", "some-registry.io/repo/new-run",
					"--registry-ca-cert-path", "some-cert-path",
					"--registry-verify-certs",
				},
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: expectedStack,
					},
				},
				ExpectedOutput: `Updating ClusterStack...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest'
ClusterStack "stack-name" updated
`,
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		it("does not add stack images with the same digest", func() {
			fakeFetcher.AddStackImages(registryfakes.StackInfo{
				StackID: "stack-id",
				BuildImg: registryfakes.ImageInfo{
					Ref:    "some-registry.io/repo/new-build",
					Digest: "build-image-digest",
				},
				RunImg: registryfakes.ImageInfo{
					Ref:    "some-registry.io/repo/new-run",
					Digest: "run-image-digest",
				},
			})

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
					stack,
				},
				Args: []string{
					"stack-name",
					"--build-image", "some-registry.io/repo/new-build",
					"--run-image", "some-registry.io/repo/new-run",
				},
				ExpectedOutput: `Updating ClusterStack...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
Build and Run images already exist in stack
ClusterStack "stack-name" updated (no change)
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
				Objects: []runtime.Object{
					badConfig,
					stack,
				},
				Args: []string{
					"stack-name",
					"--build-image", "some-registry.io/repo/new-build",
					"--run-image", "some-registry.io/repo/new-run",
				},
				ExpectErr:           true,
				ExpectedOutput:      "Updating ClusterStack...\n",
				ExpectedErrorOutput: "Error: failed to get canonical repository: use \"kp config canonical-repository\" to set\n",
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest
status:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
    latestImage: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
    latestImage: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
						stack,
					},
					Args: []string{
						"stack-name",
						"--build-image", "some-registry.io/repo/new-build",
						"--run-image", "some-registry.io/repo/new-run",
						"--output", "yaml",
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: &v1alpha2.ClusterStack{
								ObjectMeta: stack.ObjectMeta,
								Spec: v1alpha2.ClusterStackSpec{
									Id: "stack-id",
									BuildImage: v1alpha2.ClusterStackSpecImage{
										Image: "canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest",
									},
									RunImage: v1alpha2.ClusterStackSpecImage{
										Image: "canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest",
									},
								},
								Status: stack.Status,
							},
						},
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Updating ClusterStack...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest'
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				const resourceJSON = `{
    "kind": "ClusterStack",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "stack-name",
        "creationTimestamp": null
    },
    "spec": {
        "id": "stack-id",
        "buildImage": {
            "image": "canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest"
        },
        "runImage": {
            "image": "canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest"
        }
    },
    "status": {
        "id": "stack-id",
        "buildImage": {
            "latestImage": "canonical-registry.io/canonical-repo/build@sha256:build-image-digest",
            "image": "canonical-registry.io/canonical-repo/build@sha256:build-image-digest"
        },
        "runImage": {
            "latestImage": "canonical-registry.io/canonical-repo/run@sha256:run-image-digest",
            "image": "canonical-registry.io/canonical-repo/run@sha256:run-image-digest"
        }
    }
}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
						stack,
					},
					Args: []string{
						"stack-name",
						"--build-image", "some-registry.io/repo/new-build",
						"--run-image", "some-registry.io/repo/new-run",
						"--output", "json",
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: &v1alpha2.ClusterStack{
								ObjectMeta: stack.ObjectMeta,
								Spec: v1alpha2.ClusterStackSpec{
									Id: "stack-id",
									BuildImage: v1alpha2.ClusterStackSpecImage{
										Image: "canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest",
									},
									RunImage: v1alpha2.ClusterStackSpecImage{
										Image: "canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest",
									},
								},
								Status: stack.Status,
							},
						},
					},
					ExpectedOutput: resourceJSON,
					ExpectedErrorOutput: `Updating ClusterStack...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest'
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("there are no changes in the update", func() {
				fakeFetcher.AddStackImages(registryfakes.StackInfo{
					StackID: "stack-id",
					BuildImg: registryfakes.ImageInfo{
						Ref:    "some-registry.io/repo/new-build",
						Digest: "build-image-digest",
					},
					RunImg: registryfakes.ImageInfo{
						Ref:    "some-registry.io/repo/new-run",
						Digest: "run-image-digest",
					},
				})

				it("can output original resource in requested format", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
status:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
    latestImage: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
    latestImage: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							config,
							stack,
						},
						Args: []string{
							"stack-name",
							"--build-image", "some-registry.io/repo/new-build",
							"--run-image", "some-registry.io/repo/new-run",
							"--output", "yaml",
						},
						ExpectedErrorOutput: `Updating ClusterStack...
Uploading to 'canonical-registry.io/canonical-repo'...
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:run-image-digest'
Build and Run images already exist in stack
`,
						ExpectedOutput: resourceYAML,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})

		when("dry-run flag is used", func() {
			fakeRelocator.SetSkip(true)

			it("does not update the clusterstack and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
						stack,
					},
					Args: []string{
						"stack-name",
						"--build-image", "some-registry.io/repo/new-build",
						"--run-image", "some-registry.io/repo/new-run",
						"--dry-run",
					},
					ExpectedOutput: `Updating ClusterStack... (dry run)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest'
	Skipping 'canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest'
ClusterStack "stack-name" updated (dry run)
`,
				}.TestK8sAndKpack(t, cmdFunc)
				require.Len(t, fakeWaiter.WaitCalls, 0)
			})

			when("output flag is used", func() {
				it("does not update the clusterstack and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest
status:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
    latestImage: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
    latestImage: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							config,
							stack,
						},
						Args: []string{
							"stack-name",
							"--build-image", "some-registry.io/repo/new-build",
							"--run-image", "some-registry.io/repo/new-run",
							"--dry-run",
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
						ExpectedErrorOutput: `Updating ClusterStack... (dry run)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest'
	Skipping 'canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest'
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})

		when("dry-run--with-image-upload flag is used", func() {
			it("does not update the clusterstack and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
						stack,
					},
					Args: []string{
						"stack-name",
						"--build-image", "some-registry.io/repo/new-build",
						"--run-image", "some-registry.io/repo/new-run",
						"--dry-run-with-image-upload",
					},
					ExpectedOutput: `Updating ClusterStack... (dry run with image upload)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest'
ClusterStack "stack-name" updated (dry run with image upload)
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("output flag is used", func() {
				it("does not update the clusterstack and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest
status:
  buildImage:
    image: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
    latestImage: canonical-registry.io/canonical-repo/build@sha256:build-image-digest
  id: stack-id
  runImage:
    image: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
    latestImage: canonical-registry.io/canonical-repo/run@sha256:run-image-digest
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							config,
							stack,
						},
						Args: []string{
							"stack-name",
							"--build-image", "some-registry.io/repo/new-build",
							"--run-image", "some-registry.io/repo/new-run",
							"--dry-run-with-image-upload",
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
						ExpectedErrorOutput: `Updating ClusterStack... (dry run with image upload)
Uploading to 'canonical-registry.io/canonical-repo'... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/build@sha256:new-build-image-digest'
	Uploading 'canonical-registry.io/canonical-repo/run@sha256:new-run-image-digest'
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})
	})
}

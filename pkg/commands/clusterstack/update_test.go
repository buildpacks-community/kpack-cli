// Copyright 2020-Present VMware, Inc.
//
//SPDX-License-Identifier: Apache-2.0

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
	clusterstackcmds "github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterstack"
	commandsfakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
	registryfakes "github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestUpdateCommand(t *testing.T) {
	spec.Run(t, "TestUpdateCommand", testUpdateCommand(clusterstackcmds.NewUpdateCommand))
}

func testUpdateCommand(imageCommand func(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command) func(t *testing.T, when spec.G, it spec.S) {
	return func(t *testing.T, when spec.G, it spec.S) {
		fakeFetcher := registryfakes.NewStackImagesFetcher(
			registryfakes.StackInfo{
				StackID: "stack-id",
				BuildImg: registryfakes.ImageInfo{
					Ref:    "some-registry.io/repo/new-build",
					Digest: "new-build-image-digest",
				},
				RunImg: registryfakes.ImageInfo{
					Ref:    "some-registry.io/repo/new-run",
					Digest: "new-run-image-digest",
				},
			},
		)
		fakeRegistryUtilProvider := &registryfakes.UtilProvider{
			FakeFetcher: fakeFetcher,
		}

		stack := &v1alpha2.ClusterStack{
			ObjectMeta: metav1.ObjectMeta{
				Name: "stack-name",
			},
			Spec: v1alpha2.ClusterStackSpec{
				Id: "stack-id",
				BuildImage: v1alpha2.ClusterStackSpecImage{
					Image: "default-registry.io/default-repo@sha256:build-image-digest",
				},
				RunImage: v1alpha2.ClusterStackSpecImage{
					Image: "default-registry.io/default-repo@sha256:run-image-digest",
				},
				ServiceAccountRef: &corev1.ObjectReference{
					Namespace: "some-namespace",
					Name:      "some-serviceaccount",
				},
			},
			Status: v1alpha2.ClusterStackStatus{
				ResolvedClusterStack: v1alpha2.ResolvedClusterStack{
					Id: "stack-id",
					BuildImage: v1alpha2.ClusterStackStatusImage{
						LatestImage: "default-registry.io/default-repo@sha256:build-image-digest",
						Image:       "default-registry.io/default-repo@sha256:build-image-digest",
					},
					RunImage: v1alpha2.ClusterStackStatusImage{
						LatestImage: "default-registry.io/default-repo@sha256:run-image-digest",
						Image:       "default-registry.io/default-repo@sha256:run-image-digest",
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
				"default.repository":                          "default-registry.io/default-repo",
				"default.repository.serviceaccount":           "some-serviceaccount",
				"default.repository.serviceaccount.namespace": "some-namespace",
			},
		}

		fakeWaiter := &commandsfakes.FakeWaiter{}

		cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
			return imageCommand(clientSetProvider, fakeRegistryUtilProvider, func(dynamic.Interface) commands.ResourceWaiter {
				return fakeWaiter
			})
		}

		it("updates the stack but does not update the service account", func() {
			config.Data["default.repository.serviceaccount"] = "some-new-serviceaccount"
			config.Data["default.repository.serviceaccount.namespace"] = "some-new-namespace"

			expectedStack := &v1alpha2.ClusterStack{
				ObjectMeta: stack.ObjectMeta,
				Spec: v1alpha2.ClusterStackSpec{
					Id: "stack-id",
					BuildImage: v1alpha2.ClusterStackSpecImage{
						Image: "default-registry.io/default-repo@sha256:new-build-image-digest",
					},
					RunImage: v1alpha2.ClusterStackSpecImage{
						Image: "default-registry.io/default-repo@sha256:new-run-image-digest",
					},
					ServiceAccountRef: &corev1.ObjectReference{
						Namespace: "some-namespace",
						Name:      "some-serviceaccount",
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
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:new-build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:new-run-image-digest'
ClusterStack "stack-name" updated
`,
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		it("returns error when default.repository key is not found in kp-config configmap", func() {
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
				ExpectedErrorOutput: "Error: failed to get default repository: use \"kp config default-repository\" to set\n",
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
    image: default-registry.io/default-repo@sha256:new-build-image-digest
  id: stack-id
  runImage:
    image: default-registry.io/default-repo@sha256:new-run-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status:
  buildImage:
    image: default-registry.io/default-repo@sha256:build-image-digest
    latestImage: default-registry.io/default-repo@sha256:build-image-digest
  id: stack-id
  runImage:
    image: default-registry.io/default-repo@sha256:run-image-digest
    latestImage: default-registry.io/default-repo@sha256:run-image-digest
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
										Image: "default-registry.io/default-repo@sha256:new-build-image-digest",
									},
									RunImage: v1alpha2.ClusterStackSpecImage{
										Image: "default-registry.io/default-repo@sha256:new-run-image-digest",
									},
									ServiceAccountRef: &corev1.ObjectReference{
										Namespace: "some-namespace",
										Name:      "some-serviceaccount",
									},
								},
								Status: stack.Status,
							},
						},
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Updating ClusterStack...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:new-build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:new-run-image-digest'
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
            "image": "default-registry.io/default-repo@sha256:new-build-image-digest"
        },
        "runImage": {
            "image": "default-registry.io/default-repo@sha256:new-run-image-digest"
        },
        "serviceAccountRef": {
            "namespace": "some-namespace",
            "name": "some-serviceaccount"
        }
    },
    "status": {
        "id": "stack-id",
        "buildImage": {
            "latestImage": "default-registry.io/default-repo@sha256:build-image-digest",
            "image": "default-registry.io/default-repo@sha256:build-image-digest"
        },
        "runImage": {
            "latestImage": "default-registry.io/default-repo@sha256:run-image-digest",
            "image": "default-registry.io/default-repo@sha256:run-image-digest"
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
										Image: "default-registry.io/default-repo@sha256:new-build-image-digest",
									},
									RunImage: v1alpha2.ClusterStackSpecImage{
										Image: "default-registry.io/default-repo@sha256:new-run-image-digest",
									},
									ServiceAccountRef: &corev1.ObjectReference{
										Namespace: "some-namespace",
										Name:      "some-serviceaccount",
									},
								},
								Status: stack.Status,
							},
						},
					},
					ExpectedOutput: resourceJSON,
					ExpectedErrorOutput: `Updating ClusterStack...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:new-build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:new-run-image-digest'
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

				it("reports no change", func() {
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
						ExpectErr: false,
						ExpectedOutput: `Updating ClusterStack...
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:run-image-digest'
ClusterStack "stack-name" updated (no change)
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})

				it("can output original resource in requested format", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStack
metadata:
  creationTimestamp: null
  name: stack-name
spec:
  buildImage:
    image: default-registry.io/default-repo@sha256:build-image-digest
  id: stack-id
  runImage:
    image: default-registry.io/default-repo@sha256:run-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status:
  buildImage:
    image: default-registry.io/default-repo@sha256:build-image-digest
    latestImage: default-registry.io/default-repo@sha256:build-image-digest
  id: stack-id
  runImage:
    image: default-registry.io/default-repo@sha256:run-image-digest
    latestImage: default-registry.io/default-repo@sha256:run-image-digest
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
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:run-image-digest'
`,
						ExpectedOutput: resourceYAML,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})

		when("dry-run flag is used", func() {
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
Uploading to 'default-registry.io/default-repo'... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:new-build-image-digest'
	Skipping 'default-registry.io/default-repo@sha256:new-run-image-digest'
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
    image: default-registry.io/default-repo@sha256:new-build-image-digest
  id: stack-id
  runImage:
    image: default-registry.io/default-repo@sha256:new-run-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status:
  buildImage:
    image: default-registry.io/default-repo@sha256:build-image-digest
    latestImage: default-registry.io/default-repo@sha256:build-image-digest
  id: stack-id
  runImage:
    image: default-registry.io/default-repo@sha256:run-image-digest
    latestImage: default-registry.io/default-repo@sha256:run-image-digest
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
Uploading to 'default-registry.io/default-repo'... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:new-build-image-digest'
	Skipping 'default-registry.io/default-repo@sha256:new-run-image-digest'
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
Uploading to 'default-registry.io/default-repo'... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:new-build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:new-run-image-digest'
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
    image: default-registry.io/default-repo@sha256:new-build-image-digest
  id: stack-id
  runImage:
    image: default-registry.io/default-repo@sha256:new-run-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
status:
  buildImage:
    image: default-registry.io/default-repo@sha256:build-image-digest
    latestImage: default-registry.io/default-repo@sha256:build-image-digest
  id: stack-id
  runImage:
    image: default-registry.io/default-repo@sha256:run-image-digest
    latestImage: default-registry.io/default-repo@sha256:run-image-digest
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
Uploading to 'default-registry.io/default-repo'... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:new-build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:new-run-image-digest'
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})
	}
}

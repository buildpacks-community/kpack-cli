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

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	clusterstackcmds "github.com/buildpacks-community/kpack-cli/pkg/commands/clusterstack"
	commandsfakes "github.com/buildpacks-community/kpack-cli/pkg/commands/fakes"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
	registryfakes "github.com/buildpacks-community/kpack-cli/pkg/registry/fakes"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
)

func TestCreateCommand(t *testing.T) {
	spec.Run(t, "TestCreateCommand", testCreateCommand(clusterstackcmds.NewCreateCommand))
}

func testCreateCommand(imageCommand func(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command) func(t *testing.T, when spec.G, it spec.S) {
	return func(t *testing.T, when spec.G, it spec.S) {
		stackInfo := registryfakes.StackInfo{
			StackID: "stack-id",
			BuildImg: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/some-build-image",
				Digest: "build-image-digest",
			},
			RunImg: registryfakes.ImageInfo{
				Ref:    "some-registry.io/repo/some-run-image",
				Digest: "run-image-digest",
			},
		}

		fakeRelocator := &registryfakes.Relocator{}
		fakeRegistryUtilProvider := &registryfakes.UtilProvider{
			FakeFetcher: registryfakes.NewStackImagesFetcher(stackInfo),
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

		expectedStack := &v1alpha2.ClusterStack{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.ClusterStackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        "stack-name",
				Annotations: nil,
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
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:run-image-digest'
ClusterStack "stack-name" created
`,
				ExpectCreates: []runtime.Object{
					expectedStack,
				},
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		it("fails when default.repository key is not found in kp-config configmap", func() {
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
    image: default-registry.io/default-repo@sha256:build-image-digest
  id: stack-id
  runImage:
    image: default-registry.io/default-repo@sha256:run-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
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
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:run-image-digest'
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
            "image": "default-registry.io/default-repo@sha256:build-image-digest"
        },
        "runImage": {
            "image": "default-registry.io/default-repo@sha256:run-image-digest"
        },
        "serviceAccountRef": {
            "namespace": "some-namespace",
            "name": "some-serviceaccount"
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
Uploading to 'default-registry.io/default-repo'...
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:run-image-digest'
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
Uploading to 'default-registry.io/default-repo'... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:build-image-digest'
	Skipping 'default-registry.io/default-repo@sha256:run-image-digest'
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
    image: default-registry.io/default-repo@sha256:build-image-digest
  id: stack-id
  runImage:
    image: default-registry.io/default-repo@sha256:run-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
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
Uploading to 'default-registry.io/default-repo'... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:build-image-digest'
	Skipping 'default-registry.io/default-repo@sha256:run-image-digest'
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
Uploading to 'default-registry.io/default-repo'... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:run-image-digest'
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
    image: default-registry.io/default-repo@sha256:build-image-digest
  id: stack-id
  runImage:
    image: default-registry.io/default-repo@sha256:run-image-digest
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
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
Uploading to 'default-registry.io/default-repo'... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:build-image-digest'
	Uploading 'default-registry.io/default-repo@sha256:run-image-digest'
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})
	}
}

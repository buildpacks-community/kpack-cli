// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	storecmds "github.com/buildpacks-community/kpack-cli/pkg/commands/clusterstore"
	commandsfakes "github.com/buildpacks-community/kpack-cli/pkg/commands/fakes"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
	registryfakes "github.com/buildpacks-community/kpack-cli/pkg/registry/fakes"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	k8sfakes "k8s.io/client-go/kubernetes/fake"
)

func TestClusterStoreAddCommand(t *testing.T) {
	spec.Run(t, "TestClusterStoreAddCommand", testAddCommand(storecmds.NewAddCommand))
	spec.Run(t, "TestClusterStoreAddCommandDNEError", testAddCommandDNEError)
}

func testAddCommand(clusterStackCommand func(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command) func(t *testing.T, when spec.G, it spec.S) {
	return func(t *testing.T, when spec.G, it spec.S) {
		fakeRegistryUtilProvider := &registryfakes.UtilProvider{
			FakeFetcher: registryfakes.NewBuildpackImagesFetcher(
				registryfakes.BuildpackImgInfo{
					Id: "old-buildpack-id",
					ImageInfo: registryfakes.ImageInfo{
						Ref:    "default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest",
						Digest: "old-buildpack-digest",
					},
				},
				registryfakes.BuildpackImgInfo{
					Id: "new-buildpack-id",
					ImageInfo: registryfakes.ImageInfo{
						Ref:    "some-registry.io/repo/new-buildpack",
						Digest: "new-buildpack-digest",
					},
				},
			),
		}

		fakeWaiter := &commandsfakes.FakeWaiter{}

		existingStore := &v1alpha2.ClusterStore{
			ObjectMeta: v1.ObjectMeta{
				Name: "store-name",
			},
			Spec: v1alpha2.ClusterStoreSpec{
				ServiceAccountRef: &corev1.ObjectReference{
					Namespace: "some-namespace",
					Name:      "some-serviceaccount",
				},
				Sources: []corev1alpha1.ImageSource{
					{Image: "default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest"},
				},
			},
		}

		config := &corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{
				"default.repository":                          "default-registry.io/default-repo",
				"default.repository.serviceaccount":           "some-serviceaccount",
				"default.repository.serviceaccount.namespace": "some-namespace",
			},
		}

		cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
			return clusterStackCommand(clientSetProvider, fakeRegistryUtilProvider, func(dynamic.Interface) commands.ResourceWaiter {
				return fakeWaiter
			})
		}

		it("adds a buildpackage to store but does not update the default service account", func() {
			config.Data["default.repository.serviceaccount"] = "some-new-serviceaccount"
			config.Data["default.repository.serviceaccount.namespace"] = "some-new-namespace"

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
					existingStore,
				},
				Args: []string{
					"store-name",
					"--buildpackage", "some-registry.io/repo/new-buildpack",
					"-b", localCNBPath,
					"--registry-ca-cert-path", "some-cert-path",
					"--registry-verify-certs",
				},
				ExpectPatches: []string{
					`{"spec":{"sources":[{"image":"default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest"},{"image":"default-registry.io/default-repo@sha256:new-buildpack-digest"},{"image":"default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"}]}}`,
				},
				ExpectedOutput: `Adding to ClusterStore...
	Uploading 'default-registry.io/default-repo@sha256:new-buildpack-digest'
	Added Buildpackage
	Uploading 'default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
ClusterStore "store-name" updated
`,
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		it("does not add buildpackage with the same digest", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					config,
					existingStore,
				},
				Args: []string{
					"store-name",
					"-b", "default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest",
				},
				ExpectedOutput: `Adding to ClusterStore...
	Uploading 'default-registry.io/default-repo@sha256:old-buildpack-digest'
	Buildpackage already exists in the store
ClusterStore "store-name" updated (no change)
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("errors when default.repository key is not found", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					existingStore,
				},
				Args: []string{
					"store-name",
					"-b", "some/someimage",
				},
				ExpectErr:           true,
				ExpectedOutput:      "Adding to ClusterStore...\n",
				ExpectedErrorOutput: "Error: failed to get default repository: use \"kp config default-repository\" to set\n",
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStore
metadata:
  creationTimestamp: null
  name: store-name
spec:
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
  sources:
  - image: default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest
  - image: default-registry.io/default-repo@sha256:new-buildpack-digest
  - image: default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf
status: {}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
						existingStore,
					},
					Args: []string{
						"store-name",
						"--buildpackage", "some-registry.io/repo/new-buildpack",
						"-b", localCNBPath,
						"--output", "yaml",
					},
					ExpectPatches: []string{
						`{"spec":{"sources":[{"image":"default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest"},{"image":"default-registry.io/default-repo@sha256:new-buildpack-digest"},{"image":"default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"}]}}`,
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Adding to ClusterStore...
	Uploading 'default-registry.io/default-repo@sha256:new-buildpack-digest'
	Added Buildpackage
	Uploading 'default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				const resourceJSON = `{
    "kind": "ClusterStore",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "store-name",
        "creationTimestamp": null
    },
    "spec": {
        "sources": [
            {
                "image": "default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest"
            },
            {
                "image": "default-registry.io/default-repo@sha256:new-buildpack-digest"
            },
            {
                "image": "default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"
            }
        ],
        "serviceAccountRef": {
            "namespace": "some-namespace",
            "name": "some-serviceaccount"
        }
    },
    "status": {}
}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
						existingStore,
					},
					Args: []string{
						"store-name",
						"--buildpackage", "some-registry.io/repo/new-buildpack",
						"-b", localCNBPath,
						"--output", "json",
					},
					ExpectPatches:  []string{`{"spec":{"sources":[{"image":"default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest"},{"image":"default-registry.io/default-repo@sha256:new-buildpack-digest"},{"image":"default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"}]}}`},
					ExpectedOutput: resourceJSON,
					ExpectedErrorOutput: `Adding to ClusterStore...
	Uploading 'default-registry.io/default-repo@sha256:new-buildpack-digest'
	Added Buildpackage
	Uploading 'default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("there are no changes in the update", func() {
				it("can output original resource in requested format", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStore
metadata:
  creationTimestamp: null
  name: store-name
spec:
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
  sources:
  - image: default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest
status: {}
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							config,
							existingStore,
						},
						Args: []string{
							"store-name",
							"-b", "default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest",
							"--output", "yaml",
						},
						ExpectedErrorOutput: `Adding to ClusterStore...
	Uploading 'default-registry.io/default-repo@sha256:old-buildpack-digest'
	Buildpackage already exists in the store
`,
						ExpectedOutput: resourceYAML,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})

		when("dry-run flag is used", func() {
			it("does not create a clusterstore and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
						existingStore,
					},
					Args: []string{
						"store-name",
						"--buildpackage", "some-registry.io/repo/new-buildpack",
						"-b", localCNBPath,
						"--dry-run",
					},
					ExpectedOutput: `Adding to ClusterStore... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:new-buildpack-digest'
	Added Buildpackage
	Skipping 'default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
ClusterStore "store-name" updated (dry run)
`,
				}.TestK8sAndKpack(t, cmdFunc)
				require.Len(t, fakeWaiter.WaitCalls, 0)
			})

			when("there are no changes in the update", func() {
				it("does not create a clusterstore and informs of no change", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							config,
							existingStore,
						},
						Args: []string{
							"store-name",
							"-b", "default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest",
							"--dry-run",
						},
						ExpectedOutput: `Adding to ClusterStore... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:old-buildpack-digest'
	Buildpackage already exists in the store
ClusterStore "store-name" updated (dry run)
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})

			when("output flag is used", func() {
				it("does not create a clusterstore and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStore
metadata:
  creationTimestamp: null
  name: store-name
spec:
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
  sources:
  - image: default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest
  - image: default-registry.io/default-repo@sha256:new-buildpack-digest
  - image: default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf
status: {}
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							config,
							existingStore,
						},
						Args: []string{
							"store-name",
							"--buildpackage", "some-registry.io/repo/new-buildpack",
							"-b", localCNBPath,
							"--dry-run",
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
						ExpectedErrorOutput: `Adding to ClusterStore... (dry run)
	Skipping 'default-registry.io/default-repo@sha256:new-buildpack-digest'
	Added Buildpackage
	Skipping 'default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})

		when("dry-run-with-image-upload flag is used", func() {
			it("does not create a clusterstore and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						config,
						existingStore,
					},
					Args: []string{
						"store-name",
						"--buildpackage", "some-registry.io/repo/new-buildpack",
						"-b", localCNBPath,
						"--dry-run-with-image-upload",
					},
					ExpectedOutput: `Adding to ClusterStore... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:new-buildpack-digest'
	Added Buildpackage
	Uploading 'default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
ClusterStore "store-name" updated (dry run with image upload)
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("there are no changes in the update", func() {
				it("does not create a clusterstore and informs of no change", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							config,
							existingStore,
						},
						Args: []string{
							"store-name",
							"-b", "default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest",
							"--dry-run-with-image-upload",
						},
						ExpectedOutput: `Adding to ClusterStore... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:old-buildpack-digest'
	Buildpackage already exists in the store
ClusterStore "store-name" updated (dry run with image upload)
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})

			when("output flag is used", func() {
				it("does not create a clusterstore and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: ClusterStore
metadata:
  creationTimestamp: null
  name: store-name
spec:
  serviceAccountRef:
    name: some-serviceaccount
    namespace: some-namespace
  sources:
  - image: default-registry.io/default-repo/old-buildpack-id@sha256:old-buildpack-digest
  - image: default-registry.io/default-repo@sha256:new-buildpack-digest
  - image: default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf
status: {}
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							config,
							existingStore,
						},
						Args: []string{
							"store-name",
							"--buildpackage", "some-registry.io/repo/new-buildpack",
							"-b", localCNBPath,
							"--dry-run-with-image-upload",
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
						ExpectedErrorOutput: `Adding to ClusterStore... (dry run with image upload)
	Uploading 'default-registry.io/default-repo@sha256:new-buildpack-digest'
	Added Buildpackage
	Uploading 'default-registry.io/default-repo@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})
	}
}

func testAddCommandDNEError(t *testing.T, when spec.G, it spec.S) {
	it("errors when the provided store does not exist", func() {
		cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
			return storecmds.NewAddCommand(clientSetProvider, nil, nil)
		}
		testhelpers.CommandTest{
			Objects: []runtime.Object{},
			Args: []string{
				"invalid-store",
				"-b", "some/image",
			},
			ExpectErr:           true,
			ExpectedErrorOutput: "Error: ClusterStore 'invalid-store' does not exist\n",
		}.TestK8sAndKpack(t, cmdFunc)
	})
}

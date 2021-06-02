// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
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
	storecmds "github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterstore"
	commandsfakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	registryfakes "github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestClusterStoreSaveCommand(t *testing.T) {
	spec.Run(t, "TestClusterStoreSaveCommand", testClusterStoreSaveCommand)
}

func testClusterStoreSaveCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		fakeFetcher              = &registryfakes.Fetcher{}
		fakeRegistryUtilProvider = &registryfakes.UtilProvider{
			FakeRelocator: &registryfakes.Relocator{},
			FakeFetcher:   fakeFetcher,
		}

		config = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{
				"canonical.repository":                "canonical-registry.io/canonical-repo",
				"canonical.repository.serviceaccount": "some-serviceaccount",
			},
		}
	)

	fakeWaiter := &commandsfakes.FakeWaiter{}

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return storecmds.NewSaveCommand(clientSetProvider, fakeRegistryUtilProvider, func(dynamic.Interface) commands.ResourceWaiter {
			return fakeWaiter
		})
	}

	when("creating", func() {
		newStore := &v1alpha1.ClusterStore{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterStoreKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "store-name",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"store-name","creationTimestamp":null},"spec":{"sources":[{"image":"canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest"},{"image":"canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"}]},"status":{}}`,
				},
			},
			Spec: v1alpha1.ClusterStoreSpec{
				Sources: []v1alpha1.StoreImage{
					{Image: "canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest"},
					{Image: "canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"},
				},
			},
		}

		fakeFetcher.AddBuildpackImages(
			registryfakes.BuildpackImgInfo{
				Id: "buildpack-id",
				ImageInfo: registryfakes.ImageInfo{
					Ref:    "some-registry.io/repo/buildpack",
					Digest: "buildpack-digest",
				},
			},
		)

		it("creates a cluster store", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"store-name",
					"--buildpackage", "some-registry.io/repo/buildpack",
					"-b", localCNBPath,
					"--registry-ca-cert-path", "some-cert-path",
					"--registry-verify-certs",
				},
				ExpectedOutput: `Creating ClusterStore...
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest'
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
ClusterStore "store-name" created
`,
				ExpectCreates: []runtime.Object{
					newStore,
				},
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		it("fails when kp-config configmap is not found", func() {
			testhelpers.CommandTest{
				Args: []string{
					"store-name",
					"--buildpackage", "some-registry.io/repo/buildpack",
					"-b", localCNBPath,
				},
				ExpectErr: true,
				ExpectedOutput: "Creating ClusterStore...\nError: configmaps \"kp-config\" not found\n",
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
					"store-name",
					"--buildpackage", "some-registry.io/repo/buildpack",
					"-b", localCNBPath,
				},
				ExpectErr: true,
				ExpectedOutput: "Creating ClusterStore...\nError: key \"canonical.repository\" not found in configmap \"kp-config\"\n",
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("fails when a buildpackage is not provided", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				Args: []string{
					"store-name",
				},
				ExpectErr:      true,
				ExpectedOutput: "Creating ClusterStore...\nError: At least one buildpackage must be provided\n",
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"store-name","creationTimestamp":null},"spec":{"sources":[{"image":"canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest"},{"image":"canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"}]},"status":{}}'
  creationTimestamp: null
  name: store-name
spec:
  sources:
  - image: canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest
  - image: canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf
status: {}
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					Args: []string{
						"store-name",
						"--buildpackage", "some-registry.io/repo/buildpack",
						"-b", localCNBPath,
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Creating ClusterStore...
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest'
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
`,
					ExpectCreates: []runtime.Object{
						newStore,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				const resourceJSON = `{
    "kind": "ClusterStore",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "store-name",
        "creationTimestamp": null,
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterStore\",\"apiVersion\":\"kpack.io/v1alpha1\",\"metadata\":{\"name\":\"store-name\",\"creationTimestamp\":null},\"spec\":{\"sources\":[{\"image\":\"canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest\"},{\"image\":\"canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf\"}]},\"status\":{}}"
        }
    },
    "spec": {
        "sources": [
            {
                "image": "canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest"
            },
            {
                "image": "canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"
            }
        ]
    },
    "status": {}
}
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					Args: []string{
						"store-name",
						"--buildpackage", "some-registry.io/repo/buildpack",
						"-b", localCNBPath,
						"--output", "json",
					},
					ExpectedOutput: resourceJSON,
					ExpectedErrorOutput: `Creating ClusterStore...
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest'
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
`,
					ExpectCreates: []runtime.Object{
						newStore,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		when("dry-run flag is used", func() {
			it("does not create a clusterstore and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					Args: []string{
						"store-name",
						"--buildpackage", "some-registry.io/repo/buildpack",
						"-b", localCNBPath,
						"--dry-run",
					},
					ExpectedOutput: `Creating ClusterStore... (dry run)
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest'
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
ClusterStore "store-name" created (dry run)
`,
				}.TestK8sAndKpack(t, cmdFunc)
				require.Len(t, fakeWaiter.WaitCalls, 0)
			})

			when("output flag is used", func() {
				it("does not create a clusterstore and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"store-name","creationTimestamp":null},"spec":{"sources":[{"image":"canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest"},{"image":"canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"}]},"status":{}}'
  creationTimestamp: null
  name: store-name
spec:
  sources:
  - image: canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest
  - image: canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf
status: {}
`

					testhelpers.CommandTest{
						K8sObjects: []runtime.Object{
							config,
						},
						Args: []string{
							"store-name",
							"--buildpackage", "some-registry.io/repo/buildpack",
							"-b", localCNBPath,
							"--output", "yaml",
							"--dry-run",
						},
						ExpectedOutput: resourceYAML,
						ExpectedErrorOutput: `Creating ClusterStore... (dry run)
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest'
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})

		when("dry-run-with-image-upload flag is used", func() {
			it("does not create a clusterstore and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					Args: []string{
						"store-name",
						"--buildpackage", "some-registry.io/repo/buildpack",
						"-b", localCNBPath,
						"--dry-run-with-image-upload",
					},
					ExpectedOutput: `Creating ClusterStore... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest'
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
ClusterStore "store-name" created (dry run with image upload)
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("output flag is used", func() {
				it("does not create a clusterstore and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"store-name","creationTimestamp":null},"spec":{"sources":[{"image":"canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest"},{"image":"canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"}]},"status":{}}'
  creationTimestamp: null
  name: store-name
spec:
  sources:
  - image: canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest
  - image: canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf
status: {}
`

					testhelpers.CommandTest{
						K8sObjects: []runtime.Object{
							config,
						},
						Args: []string{
							"store-name",
							"--buildpackage", "some-registry.io/repo/buildpack",
							"-b", localCNBPath,
							"--output", "yaml",
							"--dry-run-with-image-upload",
						},
						ExpectedOutput: resourceYAML,
						ExpectedErrorOutput: `Creating ClusterStore... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/buildpack-id@sha256:buildpack-digest'
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})
	})

	when("updating", func() {
		existingStore := &v1alpha1.ClusterStore{
			ObjectMeta: metav1.ObjectMeta{
				Name: "store-name",
			},
			Spec: v1alpha1.ClusterStoreSpec{
				Sources: []v1alpha1.StoreImage{
					{Image: "canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest"},
				},
			},
		}

		fakeFetcher.AddBuildpackImages(
			registryfakes.BuildpackImgInfo{
				Id: "old-buildpack-id",
				ImageInfo: registryfakes.ImageInfo{
					Ref:    "canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest",
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
		)

		it("adds a buildpackage to a store when it exists", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				KpackObjects: []runtime.Object{
					existingStore,
				},
				Args: []string{
					"store-name",
					"--buildpackage", "some-registry.io/repo/new-buildpack",
					"-b", localCNBPath,
					"--registry-ca-cert-path", "some-cert-path",
					"--registry-verify-certs",
				},
				ExpectErr: false,
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: &v1alpha1.ClusterStore{
							ObjectMeta: existingStore.ObjectMeta,
							Spec: v1alpha1.ClusterStoreSpec{
								Sources: []v1alpha1.StoreImage{
									{Image: "canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest"},
									{Image: "canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest"},
									{Image: "canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"},
								},
							},
						},
					},
				},
				ExpectedOutput: `Adding to ClusterStore...
	Uploading 'canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest'
	Added Buildpackage
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
ClusterStore "store-name" updated
`,
			}.TestK8sAndKpack(t, cmdFunc)
			require.Len(t, fakeWaiter.WaitCalls, 1)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  creationTimestamp: null
  name: store-name
spec:
  sources:
  - image: canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest
  - image: canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest
  - image: canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf
status: {}
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						existingStore,
					},
					Args: []string{
						"store-name",
						"--buildpackage", "some-registry.io/repo/new-buildpack",
						"-b", localCNBPath,
						"--output", "yaml",
					},
					ExpectErr: false,
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: &v1alpha1.ClusterStore{
								ObjectMeta: existingStore.ObjectMeta,
								Spec: v1alpha1.ClusterStoreSpec{
									Sources: []v1alpha1.StoreImage{
										{Image: "canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest"},
										{Image: "canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest"},
										{Image: "canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"},
									},
								},
							},
						},
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Adding to ClusterStore...
	Uploading 'canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest'
	Added Buildpackage
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				const resourceJSON = `{
    "kind": "ClusterStore",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "store-name",
        "creationTimestamp": null
    },
    "spec": {
        "sources": [
            {
                "image": "canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest"
            },
            {
                "image": "canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest"
            },
            {
                "image": "canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"
            }
        ]
    },
    "status": {}
}
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						existingStore,
					},
					Args: []string{
						"store-name",
						"--buildpackage", "some-registry.io/repo/new-buildpack",
						"-b", localCNBPath,
						"--output", "json",
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: &v1alpha1.ClusterStore{
								ObjectMeta: existingStore.ObjectMeta,
								Spec: v1alpha1.ClusterStoreSpec{
									Sources: []v1alpha1.StoreImage{
										{Image: "canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest"},
										{Image: "canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest"},
										{Image: "canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"},
									},
								},
							},
						},
					},
					ExpectedOutput: resourceJSON,
					ExpectedErrorOutput: `Adding to ClusterStore...
	Uploading 'canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest'
	Added Buildpackage
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("there are no changes in the update", func() {
				it("can output original resource in requested format", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  creationTimestamp: null
  name: store-name
spec:
  sources:
  - image: canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest
status: {}
`

					testhelpers.CommandTest{
						K8sObjects: []runtime.Object{
							config,
						},
						KpackObjects: []runtime.Object{
							existingStore,
						},
						Args: []string{
							"store-name",
							"-b", "canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest",
							"--output", "yaml",
						},
						ExpectErr: false,
						ExpectedErrorOutput: `Adding to ClusterStore...
	Uploading 'canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest'
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
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						existingStore,
					},
					Args: []string{
						"store-name",
						"--buildpackage", "some-registry.io/repo/new-buildpack",
						"-b", localCNBPath,
						"--dry-run",
					},
					ExpectedOutput: `Adding to ClusterStore... (dry run)
	Uploading 'canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest'
	Added Buildpackage
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
ClusterStore "store-name" updated (dry run)
`,
				}.TestK8sAndKpack(t, cmdFunc)
				require.Len(t, fakeWaiter.WaitCalls, 0)
			})

			when("there are no changes in the update", func() {
				it("does not create a clusterstore and informs of no change", func() {
					testhelpers.CommandTest{
						K8sObjects: []runtime.Object{
							config,
						},
						KpackObjects: []runtime.Object{
							existingStore,
						},
						Args: []string{
							"store-name",
							"-b", "canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest",
							"--dry-run",
						},
						ExpectErr: false,
						ExpectedOutput: `Adding to ClusterStore... (dry run)
	Uploading 'canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest'
	Buildpackage already exists in the store
ClusterStore "store-name" updated (dry run)
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})

			when("output flag is used", func() {
				it("does not create a clusterstore and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  creationTimestamp: null
  name: store-name
spec:
  sources:
  - image: canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest
  - image: canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest
  - image: canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf
status: {}
`

					testhelpers.CommandTest{
						K8sObjects: []runtime.Object{
							config,
						},
						KpackObjects: []runtime.Object{
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
	Uploading 'canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest'
	Added Buildpackage
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})

		when("dry-run-with-image-upload flag is used", func() {
			it("does not create a clusterstore and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						existingStore,
					},
					Args: []string{
						"store-name",
						"--buildpackage", "some-registry.io/repo/new-buildpack",
						"-b", localCNBPath,
						"--dry-run-with-image-upload",
					},
					ExpectedOutput: `Adding to ClusterStore... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest'
	Added Buildpackage
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
ClusterStore "store-name" updated (dry run with image upload)
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("there are no changes in the update", func() {
				it("does not create a clusterstore and informs of no change", func() {
					testhelpers.CommandTest{
						K8sObjects: []runtime.Object{
							config,
						},
						KpackObjects: []runtime.Object{
							existingStore,
						},
						Args: []string{
							"store-name",
							"-b", "canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest",
							"--dry-run-with-image-upload",
						},
						ExpectErr: false,
						ExpectedOutput: `Adding to ClusterStore... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest'
	Buildpackage already exists in the store
ClusterStore "store-name" updated (dry run with image upload)
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})

			when("output flag is used", func() {
				it("does not create a clusterstore and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  creationTimestamp: null
  name: store-name
spec:
  sources:
  - image: canonical-registry.io/canonical-repo/old-buildpack-id@sha256:old-buildpack-digest
  - image: canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest
  - image: canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf
status: {}
`

					testhelpers.CommandTest{
						K8sObjects: []runtime.Object{
							config,
						},
						KpackObjects: []runtime.Object{
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
	Uploading 'canonical-registry.io/canonical-repo/new-buildpack-id@sha256:new-buildpack-digest'
	Added Buildpackage
	Uploading 'canonical-registry.io/canonical-repo/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf'
	Added Buildpackage
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})
	})
}

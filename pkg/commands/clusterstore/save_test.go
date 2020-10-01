// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore_test

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

	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/clusterstore/fakes"
	storecmds "github.com/pivotal/build-service-cli/pkg/commands/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterStoreSaveCommand(t *testing.T) {
	spec.Run(t, "TestClusterStoreSaveCommand", testClusterStoreSaveCommand)
}

func testClusterStoreSaveCommand(t *testing.T, when spec.G, it spec.S) {
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
		return storecmds.NewSaveCommand(clientSetProvider, factory)
	}

	when("creating", func() {
		it("creates a cluster store when it does not exist", func() {
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

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-store","creationTimestamp":null},"spec":{"sources":[{"image":"some-registry.io/some-repo/newbp@sha256:123newbp"},{"image":"some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb"}]},"status":{}}'
  creationTimestamp: null
  name: test-store
spec:
  sources:
  - image: some-registry.io/some-repo/newbp@sha256:123newbp
  - image: some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb
status: {}
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					Args: []string{
						expectedStore.Name,
						"--buildpackage", buildpackage1,
						"-b", buildpackage2,
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Creating Cluster Store...
`,
					ExpectCreates: []runtime.Object{
						expectedStore,
					},
				}.TestK8sAndKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				const resourceJSON = `{
    "kind": "ClusterStore",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "test-store",
        "creationTimestamp": null,
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterStore\",\"apiVersion\":\"kpack.io/v1alpha1\",\"metadata\":{\"name\":\"test-store\",\"creationTimestamp\":null},\"spec\":{\"sources\":[{\"image\":\"some-registry.io/some-repo/newbp@sha256:123newbp\"},{\"image\":\"some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb\"}]},\"status\":{}}"
        }
    },
    "spec": {
        "sources": [
            {
                "image": "some-registry.io/some-repo/newbp@sha256:123newbp"
            },
            {
                "image": "some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb"
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
						expectedStore.Name,
						"--buildpackage", buildpackage1,
						"-b", buildpackage2,
						"--output", "json",
					},
					ExpectedOutput: resourceJSON,
					ExpectedErrorOutput: `Creating Cluster Store...
`,
					ExpectCreates: []runtime.Object{
						expectedStore,
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
						expectedStore.Name,
						"--buildpackage", buildpackage1,
						"-b", buildpackage2,
						"--dry-run",
					},
					ExpectedOutput: `Creating Cluster Store...
"test-store" created (dry run)
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("output flag is used", func() {
				it("does not create a clusterstore and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-store","creationTimestamp":null},"spec":{"sources":[{"image":"some-registry.io/some-repo/newbp@sha256:123newbp"},{"image":"some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb"}]},"status":{}}'
  creationTimestamp: null
  name: test-store
spec:
  sources:
  - image: some-registry.io/some-repo/newbp@sha256:123newbp
  - image: some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb
status: {}
`

					testhelpers.CommandTest{
						K8sObjects: []runtime.Object{
							config,
						},
						Args: []string{
							expectedStore.Name,
							"--buildpackage", buildpackage1,
							"-b", buildpackage2,
							"--output", "yaml",
							"--dry-run",
						},
						ExpectedOutput: resourceYAML,
						ExpectedErrorOutput: `Creating Cluster Store...
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})
	})

	when("updating", func() {
		fakeBuildpackageUploader["patch/bp"] = "some/path/patchbp@sha256:abc123"

		it("adds a buildpackage to a store when it exists", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				KpackObjects: []runtime.Object{
					expectedStore,
				},
				Args: []string{
					expectedStore.Name,
					"--buildpackage", "patch/bp",
				},
				ExpectErr: false,
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: &v1alpha1.ClusterStore{
							TypeMeta:   expectedStore.TypeMeta,
							ObjectMeta: expectedStore.ObjectMeta,
							Spec: v1alpha1.ClusterStoreSpec{
								Sources: []v1alpha1.StoreImage{
									{Image: uploadedBp1},
									{Image: uploadedBp2},
									{
										Image: "some/path/patchbp@sha256:abc123",
									},
								},
							},
						},
					},
				},
				ExpectedOutput: `Adding Buildpackages...
	Added Buildpackage
ClusterStore Updated
`,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("can output in yaml format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-store","creationTimestamp":null},"spec":{"sources":[{"image":"some-registry.io/some-repo/newbp@sha256:123newbp"},{"image":"some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb"}]},"status":{}}'
  creationTimestamp: null
  name: test-store
spec:
  sources:
  - image: some-registry.io/some-repo/newbp@sha256:123newbp
  - image: some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb
  - image: some/path/patchbp@sha256:abc123
status: {}
`

				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						expectedStore,
					},
					Args: []string{
						expectedStore.Name,
						"--buildpackage", "patch/bp",
						"--output", "yaml",
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: &v1alpha1.ClusterStore{
								TypeMeta:   expectedStore.TypeMeta,
								ObjectMeta: expectedStore.ObjectMeta,
								Spec: v1alpha1.ClusterStoreSpec{
									Sources: []v1alpha1.StoreImage{
										{Image: uploadedBp1},
										{Image: uploadedBp2},
										{
											Image: "some/path/patchbp@sha256:abc123",
										},
									},
								},
							},
						},
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Adding Buildpackages...
	Added Buildpackage
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			it("can output in json format", func() {
				const resourceJSON = `{
    "kind": "ClusterStore",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "test-store",
        "creationTimestamp": null,
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"ClusterStore\",\"apiVersion\":\"kpack.io/v1alpha1\",\"metadata\":{\"name\":\"test-store\",\"creationTimestamp\":null},\"spec\":{\"sources\":[{\"image\":\"some-registry.io/some-repo/newbp@sha256:123newbp\"},{\"image\":\"some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb\"}]},\"status\":{}}"
        }
    },
    "spec": {
        "sources": [
            {
                "image": "some-registry.io/some-repo/newbp@sha256:123newbp"
            },
            {
                "image": "some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb"
            },
            {
                "image": "some/path/patchbp@sha256:abc123"
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
						expectedStore,
					},
					Args: []string{
						expectedStore.Name,
						"--buildpackage", "patch/bp",
						"--output", "json",
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: &v1alpha1.ClusterStore{
								TypeMeta:   expectedStore.TypeMeta,
								ObjectMeta: expectedStore.ObjectMeta,
								Spec: v1alpha1.ClusterStoreSpec{
									Sources: []v1alpha1.StoreImage{
										{Image: uploadedBp1},
										{Image: uploadedBp2},
										{
											Image: "some/path/patchbp@sha256:abc123",
										},
									},
								},
							},
						},
					},
					ExpectedOutput: resourceJSON,
					ExpectedErrorOutput: `Adding Buildpackages...
	Added Buildpackage
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		when("dry-run flag is used", func() {
			it("does not create a clusterstore and prints result with dry run indicated", func() {
				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						expectedStore,
					},
					Args: []string{
						expectedStore.Name,
						"--buildpackage", "patch/bp",
						"--dry-run",
					},
					ExpectedOutput: `Adding Buildpackages...
	Added Buildpackage
ClusterStore Updated (dry run)
`,
				}.TestK8sAndKpack(t, cmdFunc)
			})

			when("output flag is used", func() {
				it("does not create a clusterstore and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"ClusterStore","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"test-store","creationTimestamp":null},"spec":{"sources":[{"image":"some-registry.io/some-repo/newbp@sha256:123newbp"},{"image":"some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb"}]},"status":{}}'
  creationTimestamp: null
  name: test-store
spec:
  sources:
  - image: some-registry.io/some-repo/newbp@sha256:123newbp
  - image: some-registry.io/some-repo/bpfromcnb@sha256:123imagefromcnb
  - image: some/path/patchbp@sha256:abc123
status: {}
`

					testhelpers.CommandTest{
						K8sObjects: []runtime.Object{
							config,
						},
						KpackObjects: []runtime.Object{
							expectedStore,
						},
						Args: []string{
							expectedStore.Name,
							"--buildpackage", "patch/bp",
							"--dry-run",
							"--output", "yaml",
						},
						ExpectedOutput: resourceYAML,
						ExpectedErrorOutput: `Adding Buildpackages...
	Added Buildpackage
`,
					}.TestK8sAndKpack(t, cmdFunc)
				})
			})
		})
	})
}

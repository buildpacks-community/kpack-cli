// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/clusterstore/fakes"
	storecmds "github.com/pivotal/build-service-cli/pkg/commands/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterStoreAddCommand(t *testing.T) {
	spec.Run(t, "TestClusterStoreAddCommand", testClusterStoreAddCommand)
}

func testClusterStoreAddCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		imageAlreadyInStore = "some/imageinStore@sha256:123alreadyInStore"
		storeName           = "some-store-name"
	)

	fakeBuildpackageUploader := fakes.FakeBuildpackageUploader{
		"some/newbp":    "some/path/newbp@sha256:123newbp",
		"bpfromcnb.cnb": "some/path/bpfromcnb@sha256:123imagefromcnb",

		"some/imageAlreadyInStore": "some/path/imageInStoreDifferentPath@sha256:123alreadyInStore",
	}

	factory := &clusterstore.Factory{
		Uploader: fakeBuildpackageUploader,
	}

	config := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kp-config",
			Namespace: "kpack",
		},
		Data: map[string]string{
			"canonical.repository":                "some-registry.io/some-repo",
			"canonical.repository.serviceaccount": "some-serviceaccount",
		},
	}

	store := &v1alpha1.ClusterStore{
		ObjectMeta: v1.ObjectMeta{
			Name: storeName,
		},
		Spec: v1alpha1.ClusterStoreSpec{
			Sources: []v1alpha1.StoreImage{
				{
					Image: imageAlreadyInStore,
				},
			},
		},
	}

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return storecmds.NewAddCommand(clientSetProvider, factory)
	}

	it("adds a buildpackage to store", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			KpackObjects: []runtime.Object{
				store,
			},
			Args: []string{
				storeName,
				"--buildpackage", "some/newbp",
				"-b", "bpfromcnb.cnb",
			},
			ExpectErr: false,
			ExpectUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &v1alpha1.ClusterStore{
						ObjectMeta: store.ObjectMeta,
						Spec: v1alpha1.ClusterStoreSpec{
							Sources: []v1alpha1.StoreImage{
								{
									Image: imageAlreadyInStore,
								},
								{
									Image: "some/path/newbp@sha256:123newbp",
								},
								{
									Image: "some/path/bpfromcnb@sha256:123imagefromcnb",
								},
							},
						},
					},
				},
			},
			ExpectedOutput: "Adding Buildpackages...\n\tAdded Buildpackage\n\tAdded Buildpackage\nClusterStore Updated\n",
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("does not add buildpackage with the same digest", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			KpackObjects: []runtime.Object{
				store,
			},
			Args: []string{
				storeName,
				"-b", "some/imageAlreadyInStore",
			},
			ExpectErr:      false,
			ExpectedOutput: "Adding Buildpackages...\n\tBuildpackage already exists in the store\nClusterStore Unchanged\n",
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("errors when the provided store does not exist", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			KpackObjects: []runtime.Object{
				store,
			},
			Args: []string{
				"invalid-store",
				"-b", "some/image",
			},
			ExpectErr:      true,
			ExpectedOutput: "Error: ClusterStore 'invalid-store' does not exist\n",
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("errors when kp-config configmap is not found", func() {
		testhelpers.CommandTest{
			KpackObjects: []runtime.Object{
				store,
			},
			Args: []string{
				storeName,
				"-b", "some/someimage",
			},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical repository: configmaps "kp-config" not found
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("errors when canonical.repository key is not found in kp-config configmap", func() {
		badConfig := &corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
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
				store,
			},
			Args: []string{
				storeName,
				"-b", "some/someimage",
			},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical repository: key "canonical.repository" not found in configmap "kp-config"
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	when("dry run is specified", func() {
		const resourceYAML = `metadata:
  creationTimestamp: null
  name: some-store-name
spec:
  sources:
  - image: some/imageinStore@sha256:123alreadyInStore
  - image: some/path/newbp@sha256:123newbp
  - image: some/path/bpfromcnb@sha256:123imagefromcnb
status: {}
`
		const resourceJSON = `{
    "metadata": {
        "name": "some-store-name",
        "creationTimestamp": null
    },
    "spec": {
        "sources": [
            {
                "image": "some/imageinStore@sha256:123alreadyInStore"
            },
            {
                "image": "some/path/newbp@sha256:123newbp"
            },
            {
                "image": "some/path/bpfromcnb@sha256:123imagefromcnb"
            }
        ]
    },
    "status": {}
}
`

		when("without an output format", func() {
			it("does not create a clusterstore and defaults resource output to yaml format", func() {
				testhelpers.CommandTest{
					K8sObjects: []runtime.Object{
						config,
					},
					KpackObjects: []runtime.Object{
						store,
					},
					Args: []string{
						storeName,
						"--buildpackage", "some/newbp",
						"-b", "bpfromcnb.cnb",
						"--dry-run",
					},
					ExpectedOutput: resourceYAML,
				}.TestK8sAndKpack(t, cmdFunc)
			})
		})

		it("does not add buildpackage to store and outputs the resource in yaml format", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				KpackObjects: []runtime.Object{
					store,
				},
				Args: []string{
					storeName,
					"--buildpackage", "some/newbp",
					"-b", "bpfromcnb.cnb",
					"--dry-run", "-o", "yaml",
				},
				ExpectedOutput: resourceYAML,
			}.TestK8sAndKpack(t, cmdFunc)
		})

		it("does not add buildpackage to store and outputs the resource in json format", func() {
			testhelpers.CommandTest{
				K8sObjects: []runtime.Object{
					config,
				},
				KpackObjects: []runtime.Object{
					store,
				},
				Args: []string{
					storeName,
					"--buildpackage", "some/newbp",
					"-b", "bpfromcnb.cnb",
					"--dry-run", "-o", "json",
				},
				ExpectedOutput: resourceJSON,
			}.TestK8sAndKpack(t, cmdFunc)
		})
	})
}

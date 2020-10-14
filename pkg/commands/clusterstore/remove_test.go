// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/commands/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterStoreRemoveCommand(t *testing.T) {
	spec.Run(t, "TestClusterStoreRemoveCommand", testClusterStoreRemoveCommand)
}

func testClusterStoreRemoveCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		storeName     = "some-store"
		image1InStore = "some/imageinStore1@sha256:1231alreadyInStore"
		image2InStore = "some/imageinStore2@sha256:1232alreadyInStore"
	)

	cmdFunc := func(clientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterstore.NewRemoveCommand(clientSetProvider)
	}

	store := &v1alpha1.ClusterStore{
		ObjectMeta: v1.ObjectMeta{
			Name: storeName,
		},
		Spec: v1alpha1.ClusterStoreSpec{
			Sources: []v1alpha1.StoreImage{
				{
					Image: image1InStore,
				},
				{
					Image: image2InStore,
				},
			},
		},
	}

	it("removes a single buildpackage from the store", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args: []string{
				storeName,
				"--buildpackage", "some/imageinStore1@sha256:1231alreadyInStore",
			},
			ExpectErr: false,
			ExpectUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &v1alpha1.ClusterStore{
						ObjectMeta: store.ObjectMeta,
						Spec: v1alpha1.ClusterStoreSpec{
							Sources: []v1alpha1.StoreImage{
								{
									Image: image2InStore,
								},
							},
						},
					},
				},
			},
			ExpectedOutput: `Removing Buildpackages...
Removing buildpackage some/imageinStore1@sha256:1231alreadyInStore
ClusterStore "some-store" updated
`,
		}.TestKpack(t, cmdFunc)
	})

	it("removes multiple buildpackages from the store", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args: []string{
				storeName,
				"-b", "some/imageinStore1@sha256:1231alreadyInStore",
				"-b", "some/imageinStore2@sha256:1232alreadyInStore",
			},
			ExpectErr: false,
			ExpectUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &v1alpha1.ClusterStore{
						ObjectMeta: store.ObjectMeta,
						Spec: v1alpha1.ClusterStoreSpec{
							Sources: []v1alpha1.StoreImage{},
						},
					},
				},
			},
			ExpectedOutput: `Removing Buildpackages...
Removing buildpackage some/imageinStore1@sha256:1231alreadyInStore
Removing buildpackage some/imageinStore2@sha256:1232alreadyInStore
ClusterStore "some-store" updated
`,
		}.TestKpack(t, cmdFunc)
	})

	it("fails if the provided store does not exist", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args: []string{
				"invalid-store",
				"-b", "some/imageinStore1@sha256:1231alreadyInStore",
				"-b", "some/imageNotinStore@sha256:1232notInStore",
			},
			ExpectErr:      true,
			ExpectedOutput: "Error: ClusterStore 'invalid-store' does not exist\n",
		}.TestKpack(t, cmdFunc)
	})

	it("fails if even one buildpackage is not in the store but the rest are", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args: []string{
				storeName,
				"-b", "some/imageinStore1@sha256:1231alreadyInStore",
				"-b", "some/imageNotinStore@sha256:1232notInStore",
			},
			ExpectErr:      true,
			ExpectedOutput: "Error: Buildpackage 'some/imageNotinStore@sha256:1232notInStore' does not exist in the ClusterStore\n",
		}.TestKpack(t, cmdFunc)
	})

	it("returns error if buildpackage does not exist in store", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args: []string{
				storeName,
				"-b",
				"some/imageNotinStore@sha256:1233alreadyInStore",
			},
			ExpectErr:      true,
			ExpectedOutput: "Error: Buildpackage 'some/imageNotinStore@sha256:1233alreadyInStore' does not exist in the ClusterStore\n",
		}.TestKpack(t, cmdFunc)
	})

	when("output flag is used", func() {
		it("can output in yaml format", func() {
			const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  creationTimestamp: null
  name: some-store
spec:
  sources:
  - image: some/imageinStore2@sha256:1232alreadyInStore
status: {}
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					store,
				},
				Args: []string{
					storeName,
					"--buildpackage", "some/imageinStore1@sha256:1231alreadyInStore",
					"--output", "yaml",
				},
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: &v1alpha1.ClusterStore{
							ObjectMeta: store.ObjectMeta,
							Spec: v1alpha1.ClusterStoreSpec{
								Sources: []v1alpha1.StoreImage{
									{
										Image: image2InStore,
									},
								},
							},
						},
					},
				},
				ExpectedOutput: resourceYAML,
				ExpectedErrorOutput: `Removing Buildpackages...
Removing buildpackage some/imageinStore1@sha256:1231alreadyInStore
`,
			}.TestKpack(t, cmdFunc)
		})

		it("can output in json format", func() {
			const resourceJSON = `{
    "kind": "ClusterStore",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "some-store",
        "creationTimestamp": null
    },
    "spec": {
        "sources": [
            {
                "image": "some/imageinStore2@sha256:1232alreadyInStore"
            }
        ]
    },
    "status": {}
}
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					store,
				},
				Args: []string{
					storeName,
					"--buildpackage", "some/imageinStore1@sha256:1231alreadyInStore",
					"--output", "json",
				},
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: &v1alpha1.ClusterStore{
							ObjectMeta: store.ObjectMeta,
							Spec: v1alpha1.ClusterStoreSpec{
								Sources: []v1alpha1.StoreImage{
									{
										Image: image2InStore,
									},
								},
							},
						},
					},
				},
				ExpectedOutput: resourceJSON,
				ExpectedErrorOutput: `Removing Buildpackages...
Removing buildpackage some/imageinStore1@sha256:1231alreadyInStore
`,
			}.TestKpack(t, cmdFunc)
		})

		it("errors when the provided store does not exist", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					store,
				},
				Args: []string{
					"invalid-store",
					"-b", "some/imageinStore1@sha256:1231alreadyInStore",
					"-b", "some/imageNotinStore@sha256:1232notInStore",
					"--output", "yaml",
				},
				ExpectErr:      true,
				ExpectedOutput: "Error: ClusterStore 'invalid-store' does not exist\n",
			}.TestKpack(t, cmdFunc)
		})
	})

	when("dry-run flag is used", func() {
		it("does not remove a buildpackage and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					store,
				},
				Args: []string{
					storeName,
					"--buildpackage", "some/imageinStore1@sha256:1231alreadyInStore",
					"--dry-run",
				},
				ExpectedOutput: `Removing Buildpackages... (dry run)
Removing buildpackage some/imageinStore1@sha256:1231alreadyInStore
ClusterStore "some-store" updated (dry run)
`,
			}.TestKpack(t, cmdFunc)
		})

		it("errors when the provided store does not exist", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					store,
				},
				Args: []string{
					"invalid-store",
					"-b", "some/imageinStore1@sha256:1231alreadyInStore",
					"-b", "some/imageNotinStore@sha256:1232notInStore",
					"--dry-run",
				},
				ExpectErr:      true,
				ExpectedOutput: "Error: ClusterStore 'invalid-store' does not exist\n",
			}.TestKpack(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("does not update a clusterstore and prints the resource output in requested format", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  creationTimestamp: null
  name: some-store
spec:
  sources:
  - image: some/imageinStore2@sha256:1232alreadyInStore
status: {}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						store,
					},
					Args: []string{
						storeName,
						"--buildpackage", "some/imageinStore1@sha256:1231alreadyInStore",
						"--dry-run",
						"--output", "yaml",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Removing Buildpackages... (dry run)
Removing buildpackage some/imageinStore1@sha256:1231alreadyInStore
`,
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}

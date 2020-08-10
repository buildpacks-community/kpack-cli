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
			ExpectedOutput: "Removing buildpackage some/imageinStore1@sha256:1231alreadyInStore\nClusterStore Updated\n",
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
			ExpectedOutput: "Removing buildpackage some/imageinStore1@sha256:1231alreadyInStore\nRemoving buildpackage some/imageinStore2@sha256:1232alreadyInStore\nClusterStore Updated\n",
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
			ExpectedOutput: "Error: Buildpackage 'some/imageNotinStore@sha256:1232notInStore' does not exist in the clusterstore\n",
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
			ExpectedOutput: "Error: Buildpackage 'some/imageNotinStore@sha256:1233alreadyInStore' does not exist in the clusterstore\n",
		}.TestKpack(t, cmdFunc)
	})
}

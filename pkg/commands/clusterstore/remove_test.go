// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
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

	st := &expv1alpha1.ClusterStore{
		ObjectMeta: v1.ObjectMeta{
			Name: storeName,
			Annotations: map[string]string{
				"buildservice.pivotal.io/defaultRepository": "some/path",
			},
		},
		Spec: expv1alpha1.ClusterStoreSpec{
			Sources: []expv1alpha1.StoreImage{
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
				st,
			},
			Args:      []string{storeName, "some/imageinStore1@sha256:1231alreadyInStore"},
			ExpectErr: false,
			ExpectUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &expv1alpha1.ClusterStore{
						ObjectMeta: st.ObjectMeta,
						Spec: expv1alpha1.ClusterStoreSpec{
							Sources: []expv1alpha1.StoreImage{
								{
									Image: image2InStore,
								},
							},
						},
					},
				},
			},
			ExpectedOutput: "Removing buildpackage some/imageinStore1@sha256:1231alreadyInStore\nStore Updated\n",
		}.TestKpack(t, cmdFunc)
	})

	it("removes multiple buildpackages from the store", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				st,
			},
			Args:      []string{storeName, "some/imageinStore1@sha256:1231alreadyInStore", "some/imageinStore2@sha256:1232alreadyInStore"},
			ExpectErr: false,
			ExpectUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &expv1alpha1.ClusterStore{
						ObjectMeta: st.ObjectMeta,
						Spec: expv1alpha1.ClusterStoreSpec{
							Sources: []expv1alpha1.StoreImage{},
						},
					},
				},
			},
			ExpectedOutput: "Removing buildpackage some/imageinStore1@sha256:1231alreadyInStore\nRemoving buildpackage some/imageinStore2@sha256:1232alreadyInStore\nStore Updated\n",
		}.TestKpack(t, cmdFunc)
	})

	it("fails if the provided store does not exist", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				st,
			},
			Args:           []string{"invalid-store", "some/imageinStore1@sha256:1231alreadyInStore", "some/imageNotinStore@sha256:1232notInStore"},
			ExpectErr:      true,
			ExpectedOutput: "Error: Store 'invalid-store' does not exist\n",
		}.TestKpack(t, cmdFunc)
	})

	it("fails if even one buildpackage is not in the store but the rest are", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				st,
			},
			Args:           []string{storeName, "some/imageinStore1@sha256:1231alreadyInStore", "some/imageNotinStore@sha256:1232notInStore"},
			ExpectErr:      true,
			ExpectedOutput: "Error: Buildpackage 'some/imageNotinStore@sha256:1232notInStore' does not exist in the store\n",
		}.TestKpack(t, cmdFunc)
	})

	it("returns error if buildpackage does not exist in store", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				st,
			},
			Args:           []string{storeName, "some/imageNotinStore@sha256:1233alreadyInStore"},
			ExpectErr:      true,
			ExpectedOutput: "Error: Buildpackage 'some/imageNotinStore@sha256:1233alreadyInStore' does not exist in the store\n",
		}.TestKpack(t, cmdFunc)
	})
}

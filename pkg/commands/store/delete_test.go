package store_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/commands/store"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestStoreDeleteCommand(t *testing.T) {
	spec.Run(t, "TestStoreDeleteCommand", testStoreDeleteCommand)
}

func testStoreDeleteCommand(t *testing.T, when spec.G, it spec.S) {
	const image1InStore = "some/imageinStore1@sha256:1231alreadyInStore"
	const image2InStore = "some/imageinStore2@sha256:1232alreadyInStore"

	cmdFunc := func(clientSet *kpackfakes.Clientset) *cobra.Command {
		cmdContext := testhelpers.NewFakeKpackClusterContext(clientSet)
		return store.NewDeleteCommand(cmdContext)
	}

	store := &expv1alpha1.Store{
		ObjectMeta: v1.ObjectMeta{
			Name: store.DefaultStoreName,
			Annotations: map[string]string{
				"buildservice.pivotal.io/defaultRepository": "some/path",
			},
		},
		Spec: expv1alpha1.StoreSpec{
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

	it("removes single buildpackages from the store", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args:      []string{"some/imageinStore1@sha256:1231alreadyInStore"},
			ExpectErr: false,
			ExpectUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &expv1alpha1.Store{
						ObjectMeta: store.ObjectMeta,
						Spec: expv1alpha1.StoreSpec{
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
				store,
			},
			Args:      []string{"some/imageinStore1@sha256:1231alreadyInStore", "some/imageinStore2@sha256:1232alreadyInStore"},
			ExpectErr: false,
			ExpectUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &expv1alpha1.Store{
						ObjectMeta: store.ObjectMeta,
						Spec: expv1alpha1.StoreSpec{
							Sources: []expv1alpha1.StoreImage{},
						},
					},
				},
			},
			ExpectedOutput: "Removing buildpackage some/imageinStore1@sha256:1231alreadyInStore\nRemoving buildpackage some/imageinStore2@sha256:1232alreadyInStore\nStore Updated\n",
		}.TestKpack(t, cmdFunc)
	})

	it("fails if even one buildpackage is not in the store but the rest are", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args:           []string{"some/imageinStore1@sha256:1231alreadyInStore", "some/imageNotinStore@sha256:1232notInStore"},
			ExpectErr:      true,
			ExpectedOutput: "Error: Buildpackage 'some/imageNotinStore@sha256:1232notInStore' does not exist in the store\n",
		}.TestKpack(t, cmdFunc)
	})

	it("returns error if buildpackage does not exist in store", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args:           []string{"some/imageNotinStore@sha256:1233alreadyInStore"},
			ExpectErr:      true,
			ExpectedOutput: "Error: Buildpackage 'some/imageNotinStore@sha256:1233alreadyInStore' does not exist in the store\n",
		}.TestKpack(t, cmdFunc)
	})
}

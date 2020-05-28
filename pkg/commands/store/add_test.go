package store_test

import (
	"fmt"
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	storecmds "github.com/pivotal/build-service-cli/pkg/commands/store"
	"github.com/pivotal/build-service-cli/pkg/store"
	"github.com/pivotal/build-service-cli/pkg/store/fakes"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestStoreAddCommand(t *testing.T) {
	spec.Run(t, "TestStoreAddCommand", testStoreAddCommand)
}

func testStoreAddCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		imageAlreadyInStore = "some/imageinStore@sha256:123alreadyInStore"
		storeName           = "some-store-name"
	)

	fakeBuildpackageUploader := fakes.FakeBuildpackageUploader{
		"some/newbp":    "some/path/newbp@sha256:123newbp",
		"bpfromcnb.cnb": "some/path/bpfromcnb@sha256:123imagefromcnb",

		"some/imageAlreadyInStore": "some/path/imageInStoreDifferentPath@sha256:123alreadyInStore",
	}

	factory := &store.Factory{
		Uploader: fakeBuildpackageUploader,
	}

	cmdFunc := func(clientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return storecmds.NewAddCommand(clientSetProvider, factory)
	}

	store := &expv1alpha1.Store{
		ObjectMeta: v1.ObjectMeta{
			Name: storeName,
			Annotations: map[string]string{
				"buildservice.pivotal.io/defaultRepository": "some/path",
			},
		},
		Spec: expv1alpha1.StoreSpec{
			Sources: []expv1alpha1.StoreImage{
				{
					Image: imageAlreadyInStore,
				},
			},
		},
	}

	it("adds a buildpackage to store", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args:      []string{storeName, "some/newbp", "bpfromcnb.cnb"},
			ExpectErr: false,
			ExpectUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &expv1alpha1.Store{
						ObjectMeta: store.ObjectMeta,
						Spec: expv1alpha1.StoreSpec{
							Sources: []expv1alpha1.StoreImage{
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
			ExpectedOutput: "Uploading to 'some/path'...\nAdded Buildpackage 'some/path/newbp@sha256:123newbp'\nAdded Buildpackage 'some/path/bpfromcnb@sha256:123imagefromcnb'\nStore Updated\n",
		}.TestKpack(t, cmdFunc)
	})

	it("does not add buildpackage with the same digest", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args:           []string{storeName, "some/imageAlreadyInStore"},
			ExpectErr:      false,
			ExpectedOutput: "Uploading to 'some/path'...\nBuildpackage 'some/path/imageInStoreDifferentPath@sha256:123alreadyInStore' already exists in the store\nStore Unchanged\n",
		}.TestKpack(t, cmdFunc)
	})

	it("errors if the provided store does not exist", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args:           []string{"invalid-store", "some/image"},
			ExpectErr:      true,
			ExpectedOutput: "Error: Store 'invalid-store' does not exist\n",
		}.TestKpack(t, cmdFunc)
	})

	it("errors on invalid registry annotation", func() {
		store.Annotations["buildservice.pivotal.io/defaultRepository"] = ""

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args:           []string{storeName, "some/someimage"},
			ExpectErr:      true,
			ExpectedOutput: fmt.Sprintf("Error: Unable to find default registry for store: %s\n", storeName),
		}.TestKpack(t, cmdFunc)
	})
}

package store_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/pkg/errors"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/commands/store"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestStoreAddCommand(t *testing.T) {
	spec.Run(t, "TestStoreAddCommand", testStoreAddCommand)
}

func testStoreAddCommand(t *testing.T, when spec.G, it spec.S) {
	const imageAlreadyInStore = "some/imageinStore@sha256:123alreadyInStore"

	fakeBuildpackageUploader := FakeBuildpackageUploader{
		"some/newbp":    "some/path/newbp@sha256:123newbp",
		"bpfromcnb.cnb": "some/path/bpfromcnb@sha256:123imagefromcnb",

		"some/imageAlreadyInStore": "some/path/imageInStoreDifferentPath@sha256:123alreadyInStore",
	}

	cmdFunc := func(clientSet *kpackfakes.Clientset) *cobra.Command {
		return store.NewAddCommand(clientSet, fakeBuildpackageUploader)
	}

	store := &expv1alpha1.Store{
		ObjectMeta: v1.ObjectMeta{
			Name: "build-service-store",
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
			Args:      []string{"some/newbp", "bpfromcnb.cnb"},
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
			Args:           []string{"some/imageAlreadyInStore"},
			ExpectErr:      false,
			ExpectedOutput: "Uploading to 'some/path'...\nBuildpackage 'some/path/imageInStoreDifferentPath@sha256:123alreadyInStore' already exists in the store\nStore Unchanged\n",
		}.TestKpack(t, cmdFunc)
	})

	it("returns error on invalid registry annotation", func() {
		store.Annotations["buildservice.pivotal.io/defaultRepository"] = ""

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				store,
			},
			Args:           []string{"some/someimage"},
			ExpectErr:      true,
			ExpectedOutput: "Error: Unable to find default registry for store: build-service-store\n",
		}.TestKpack(t, cmdFunc)
	})
}

type FakeBuildpackageUploader map[string]string

func (f FakeBuildpackageUploader) Upload(defaultRepository string, buildpackage string) (string, error) {
	const expectedRepository = "some/path"
	if defaultRepository != expectedRepository {
		return "", errors.Errorf("unexpected repository %s expected %s", defaultRepository, expectedRepository)
	}

	uploadedImage, ok := f[buildpackage]
	if !ok {
		return "", errors.Errorf("could not upload %s", buildpackage)
	}
	return uploadedImage, nil
}

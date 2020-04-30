package stack_test

import (
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/pkg/errors"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/commands/stack"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

const expectedRepository = "some-registry.com/some-repo"

func TestUpdateCommand(t *testing.T) {
	spec.Run(t, "TestUpdateCommand", testUpdateCommand)
}

func testUpdateCommand(t *testing.T, when spec.G, it spec.S) {
	oldBuildImage, oldRunImage := makeStackImages(t, "some-old-id")
	newBuildImage, newRunImage := makeStackImages(t, "some-new-id")

	imageUploader := fakeImageUploader{
		"some-old-build-image": fakeImageTuple{
			ref:   "some-registry.com/my-repo/build@sha256:xyz",
			image: oldBuildImage,
		},
		"some-old-run-image": fakeImageTuple{
			ref:   "some-registry.com/my-repo/run@sha256:xyz",
			image: oldRunImage,
		},
		"some-new-build-image": fakeImageTuple{
			ref:   "some-registry.com/my-repo/build@sha256:abc",
			image: newBuildImage,
		},
		"some-new-run-image": fakeImageTuple{
			ref:   "some-registry.com/my-repo/run@sha256:abc",
			image: newRunImage,
		},
	}

	cmdFunc := func(clientSet *kpackfakes.Clientset) *cobra.Command {
		return stack.NewUpdateCommand(clientSet, imageUploader)
	}

	stck := &expv1alpha1.Stack{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-stack",
			Annotations: map[string]string{
				stack.DefaultRepositoryAnnotation: expectedRepository,
			},
		},
		Spec: expv1alpha1.StackSpec{
			Id: "some-old-id",
			BuildImage: expv1alpha1.StackSpecImage{
				Image: "some-old-build-image",
			},
			RunImage: expv1alpha1.StackSpecImage{
				Image: "some-old-run-image",
			},
		},
		Status: expv1alpha1.StackStatus{
			ResolvedStack: expv1alpha1.ResolvedStack{
				Id: "some-old-id",
				BuildImage: expv1alpha1.StackStatusImage{
					LatestImage: "some-registry.com/my-repo/build@sha256:xyz",
					Image:       "some-old-build-image",
				},
				RunImage: expv1alpha1.StackStatusImage{
					LatestImage: "some-registry.com/my-repo/run@sha256:xyz",
					Image:       "some-old-run-image",
				},
			},
		},
	}

	it("updates the stack id, run image, and build image", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				stck,
			},
			Args:      []string{"some-stack", "--build-image", "some-new-build-image", "--run-image", "some-new-run-image"},
			ExpectErr: false,
			ExpectUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: &expv1alpha1.Stack{
						ObjectMeta: stck.ObjectMeta,
						Spec: expv1alpha1.StackSpec{
							Id: "some-new-id",
							BuildImage: expv1alpha1.StackSpecImage{
								Image: "some-registry.com/my-repo/build@sha256:abc",
							},
							RunImage: expv1alpha1.StackSpecImage{
								Image: "some-registry.com/my-repo/run@sha256:abc",
							},
						},
						Status: stck.Status,
					},
				},
			},
			ExpectedOutput: "Uploading to 'some-registry.com/some-repo'...\nStack Updated\n",
		}.TestKpack(t, cmdFunc)
	})

	it("does not add stack images with the same digest", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				stck,
			},
			Args:           []string{"some-stack", "--build-image", "some-old-build-image", "--run-image", "some-old-run-image"},
			ExpectErr:      false,
			ExpectedOutput: "Uploading to 'some-registry.com/some-repo'...\nBuild and Run images already exist in stack\nStack Unchanged\n",
		}.TestKpack(t, cmdFunc)
	})

	it("returns error on invalid registry annotation", func() {
		stck.Annotations[stack.DefaultRepositoryAnnotation] = ""

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				stck,
			},
			Args:           []string{"some-stack", "--build-image", "some-new-build-image", "--run-image", "some-new-run-image"},
			ExpectErr:      true,
			ExpectedOutput: "Error: Unable to find default registry for stack: some-stack\n",
		}.TestKpack(t, cmdFunc)
	})

	it("returns error when build image and run image have different stack Ids", func() {
		_, runImage := makeStackImages(t, "other-stack-id")

		imageUploader["some-new-run-image"] = fakeImageTuple{
			ref:   "some-registry.com/my-repo/run@sha256:abc",
			image: runImage,
		}

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				stck,
			},
			Args:           []string{"some-stack", "--build-image", "some-new-build-image", "--run-image", "some-new-run-image"},
			ExpectErr:      true,
			ExpectedOutput: "Uploading to 'some-registry.com/some-repo'...\nError: build stack 'some-new-id' does not match run stack 'other-stack-id'\n",
		}.TestKpack(t, cmdFunc)
	})
}

func makeStackImages(t *testing.T, stackId string) (v1.Image, v1.Image) {
	buildImage, err := random.Image(0, 0)
	if err != nil {
		t.Fatal(err)
	}

	buildImage, err = imagehelpers.SetStringLabel(buildImage, stack.StackIdLabel, stackId)
	if err != nil {
		t.Fatal(err)
	}

	runImage, err := random.Image(0, 0)
	if err != nil {
		t.Fatal(err)
	}

	runImage, err = imagehelpers.SetStringLabel(runImage, stack.StackIdLabel, stackId)
	if err != nil {
		t.Fatal(err)
	}

	return buildImage, runImage
}

type fakeImageTuple struct {
	ref   string
	image v1.Image
}

type fakeImageUploader map[string]fakeImageTuple

func (f fakeImageUploader) Upload(repository, name, image string) (string, v1.Image, error) {
	if repository != expectedRepository {
		return "", nil, errors.Errorf("unexpected repository %s expected %s", repository, expectedRepository)
	}
	tuple, ok := f[image]
	if !ok {
		return "", nil, errors.Errorf("could not upload %s", image)
	}
	return tuple.ref, tuple.image, nil
}

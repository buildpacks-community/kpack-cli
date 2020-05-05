package stack_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	stackcmds "github.com/pivotal/build-service-cli/pkg/commands/stack"
	"github.com/pivotal/build-service-cli/pkg/image/fakes"
	"github.com/pivotal/build-service-cli/pkg/stack"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestCreateCommand(t *testing.T) {
	spec.Run(t, "TestCreateCommand", testCreateCommand)
}

func testCreateCommand(t *testing.T, when spec.G, it spec.S) {
	buildImage, buildImageId, runImage, runImageId := makeStackImages(t, "some-stack-id")

	fetcher := &fakes.Fetcher{}
	fetcher.AddImage("some-build-image", buildImage)
	fetcher.AddImage("some-run-image", runImage)

	relocator := &fakes.Relocator{}

	stackFactory := &stack.Factory{
		Fetcher:   fetcher,
		Relocator: relocator,
	}

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return stackcmds.NewCreateCommand(clientSet, stackFactory)
	}

	it("creates a stack", func() {
		expectedStack := &expv1alpha1.Stack{
			ObjectMeta: metav1.ObjectMeta{
				Name: "some-stack",
				Annotations: map[string]string{
					stack.DefaultRepositoryAnnotation: "some-registry.io/some-repo",
				},
			},
			Spec: expv1alpha1.StackSpec{
				Id: "some-stack-id",
				BuildImage: expv1alpha1.StackSpecImage{
					Image: "some-registry.io/some-repo/build@" + buildImageId,
				},
				RunImage: expv1alpha1.StackSpecImage{
					Image: "some-registry.io/some-repo/run@" + runImageId,
				},
			},
		}

		testhelpers.CommandTest{
			Args: []string{
				"some-stack",
				"--default-repository", "some-registry.io/some-repo",
				"--build-image", "some-build-image",
				"--run-image", "some-run-image",
			},
			ExpectedOutput: "\"some-stack\" created\n",
			ExpectCreates: []runtime.Object{
				expectedStack,
			},
		}.TestKpack(t, cmdFunc)
	})

	it("validates build stack ID is equal to run stack ID", func() {
		_, _, runImage, _ := makeStackImages(t, "some-other-stack-id")

		fetcher.AddImage("some-other-run-image", runImage)

		testhelpers.CommandTest{
			Args: []string{
				"some-stack",
				"--default-repository", "a-bad-repo",
				"--build-image", "some-build-image",
				"--run-image", "some-other-run-image",
			},
			ExpectErr:      true,
			ExpectedOutput: "Error: build stack 'some-stack-id' does not match run stack 'some-other-stack-id'\n",
		}.TestKpack(t, cmdFunc)
	})
}

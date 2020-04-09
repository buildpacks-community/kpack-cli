package image_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/image"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestImageDeleteCommand(t *testing.T) {
	spec.Run(t, "TestImageDeleteCommand", testImageDeleteCommand)
}

func testImageDeleteCommand(t *testing.T, when spec.G, it spec.S) {

	const defaultNamespace = "some-default-namespace"

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return image.NewDeleteCommand(clientSet, defaultNamespace)
	}

	when("a namespace is provided", func() {
		when("an image is available", func() {
			it("deletes the image", func() {
				image := &v1alpha1.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "some-image",
						Namespace: "some-namespace",
					},
				}
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						image,
					},
					Args: []string{"some-image", "-n", "some-namespace"},
					ExpectDeletes: []string{
						image.Name,
					},
					ExpectedOutput: `"some-image" deleted
`,
				}.Test(t, cmdFunc)
			})
		})
	})

	when("a namespace is not provided", func() {
		when("an image is available", func() {
			it("deletes the image", func() {
				image := &v1alpha1.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "some-image",
						Namespace: defaultNamespace,
					},
				}
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						image,
					},
					ExpectDeletes: []string{
						image.Name,
					},
					Args: []string{"some-image"},
					ExpectedOutput: `"some-image" deleted
`,
				}.Test(t, cmdFunc)
			})
		})

		when("an image is not available", func() {
			it("returns an error", func() {
				testhelpers.CommandTest{
					Objects: nil,
					Args:    []string{"some-image", "-n", "some-namespace"},
					ExpectedOutput: `Error: image "some-image" not found
`,
					ExpectDeletes: []string{
						"some-image",
					},
					ExpectErr: true,
				}.Test(t, cmdFunc)
			})
		})
	})
}

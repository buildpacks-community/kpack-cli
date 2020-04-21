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

func TestImageGetCommand(t *testing.T) {
	spec.Run(t, "TestImageGetCommand", testImageGetCommand)
}

func testImageGetCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return image.NewGetCommand(clientSet, defaultNamespace)
	}

	when("a namespace is provided", func() {
		when("an image is available", func() {
			it("returns the image", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						&v1alpha1.Image{
							ObjectMeta: v1.ObjectMeta{
								Name:      "some-image",
								Namespace: "some-namespace",
							},
						},
					},
					Args: []string{"some-image", "-n", "some-namespace"},
					ExpectedOutput: `metadata:
  creationTimestamp: null
  name: some-image
  namespace: some-namespace
spec:
  builder: {}
  source: {}
  tag: ""
status: {}
`,
				}.TestKpack(t, cmdFunc)
			})
		})

		when("an image is not available", func() {
			it("returns an error", func() {
				testhelpers.CommandTest{
					Objects:        nil,
					Args:           []string{"some-image", "-n", "some-namespace"},
					ExpectedOutput: "Error: images.build.pivotal.io \"some-image\" not found\n",
					ExpectErr:      true,
				}.TestKpack(t, cmdFunc)
			})
		})
	})

	when("a namespace is not provided", func() {
		when("an image is available", func() {
			it("returns the image", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						&v1alpha1.Image{
							ObjectMeta: v1.ObjectMeta{
								Name:      "some-image",
								Namespace: defaultNamespace,
							},
						},
					},
					Args: []string{"some-image"},
					ExpectedOutput: `metadata:
  creationTimestamp: null
  name: some-image
  namespace: some-default-namespace
spec:
  builder: {}
  source: {}
  tag: ""
status: {}
`,
				}.TestKpack(t, cmdFunc)
			})
		})

		when("an image is not available", func() {
			it("returns an error", func() {
				testhelpers.CommandTest{
					Objects:        nil,
					Args:           []string{"some-image"},
					ExpectedOutput: "Error: images.build.pivotal.io \"some-image\" not found\n",
					ExpectErr:      true,
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}

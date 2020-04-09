package image_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/commands/image"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestImageApplyCommand(t *testing.T) {
	spec.Run(t, "TestImageApplyCommand", testImageApplyCommand)
}

func testImageApplyCommand(t *testing.T, when spec.G, it spec.S) {

	const defaultNamespace = "some-default-namespace"

	var (
		expectedImage = &v1alpha1.Image{
			TypeMeta: v1.TypeMeta{
				Kind:       "Image",
				APIVersion: "build.pivotal.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-image",
				Namespace: "test-namespace",
			},
			Spec: v1alpha1.ImageSpec{
				Tag: "sample/image-from-git",
				Builder: corev1.ObjectReference{
					Kind: "ClusterBuilder",
					Name: "cluster-sample-builder",
				},
				Source: v1alpha1.SourceConfig{
					Git: &v1alpha1.Git{
						URL:      "https://github.com/buildpack/sample-java-app.git",
						Revision: "master",
					},
				},
				ServiceAccount: "service-account",
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return image.NewApplyCommand(clientSet, defaultNamespace)
	}

	when("a valid image config exists", func() {
		it("returns a success message from the image applier", func() {

			testhelpers.CommandTest{
				Args: []string{"-f", "./testdata/image.yaml"},
				ExpectedOutput: `"test-image" applied
`,
				ExpectCreates: []runtime.Object{
					expectedImage,
				},
			}.Test(t, cmdFunc)
		})

		when("a valid image config with no namespace exists", func() {
			it("uses the default namespace", func() {
				expectedImage.Namespace = defaultNamespace

				testhelpers.CommandTest{
					Args: []string{"-f", "./testdata/image-without-namespace.yaml"},
					ExpectedOutput: `"test-image" applied
`,
					ExpectCreates: []runtime.Object{
						expectedImage,
					},
				}.Test(t, cmdFunc)
			})
		})

		when("a valid image config is applied for an existing image", func() {
			it("updates the image", func() {
				existingImage := expectedImage.DeepCopy()
				existingImage.Spec.Source.Git.Revision = "old-git-revision"

				testhelpers.CommandTest{
					Args: []string{"-f", "./testdata/image.yaml"},
					Objects: []runtime.Object{
						existingImage,
					},
					ExpectedOutput: `"test-image" applied
`,
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedImage,
						},
					},
				}.Test(t, cmdFunc)
			})
		})
	})
}

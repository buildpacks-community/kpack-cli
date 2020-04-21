package image_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/image"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestImageStatusCommand(t *testing.T) {
	spec.Run(t, "TestImageStatusCommand", testImageStatusCommand)
}

func testImageStatusCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		defaultNamespace = "some-default-namespace"
		namespace        = "test-namespace"
		imageName        = "test-image"
	)

	testBuilds := testhelpers.MakeTestBuilds(imageName, defaultNamespace)
	testNamespacedBuilds := testhelpers.MakeTestBuilds(imageName, namespace)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return image.NewStatusCommand(clientSet, defaultNamespace)
	}

	when("a namespace is provided", func() {
		when("the namespaces has images", func() {
			it("returns a table of image details", func() {
				image := &v1alpha1.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      imageName,
						Namespace: namespace,
					},
					Status: v1alpha1.ImageStatus{
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionFalse,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-1@sha256:abcdef123",
					},
				}

				const expectedOutput = `Status:         NOT READY
Message:        N/A
LatestImage:    test-registry.io/test-image-1@sha256:abcdef123

Last Successful Build
Id:        1
Reason:    CONFIG

Last Failed Build
Id:        2
Reason:    COMMIT,BUILDPACK

`

				testhelpers.CommandTest{
					Objects:        append([]runtime.Object{image}, testNamespacedBuilds...),
					Args:           []string{imageName, "-n", namespace},
					ExpectedOutput: expectedOutput,
				}.TestKpack(t, cmdFunc)
			})

			when("the namespace has no images", func() {
				it("returns a message that the namespace has no images", func() {
					testhelpers.CommandTest{
						Args:           []string{imageName, "-n", namespace},
						ExpectErr:      true,
						ExpectedOutput: "Error: images.build.pivotal.io \"test-image\" not found\n",
					}.TestKpack(t, cmdFunc)

				})
			})
		})
	})

	when("a namespace is not provided", func() {
		when("the namespaces has images", func() {
			it("returns a table of image details", func() {
				image := &v1alpha1.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      imageName,
						Namespace: defaultNamespace,
					},
					Status: v1alpha1.ImageStatus{
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionFalse,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-1@sha256:abcdef123",
					},
				}

				const expectedOutput = `Status:         NOT READY
Message:        N/A
LatestImage:    test-registry.io/test-image-1@sha256:abcdef123

Last Successful Build
Id:        1
Reason:    CONFIG

Last Failed Build
Id:        2
Reason:    COMMIT,BUILDPACK

`

				testhelpers.CommandTest{
					Objects:        append([]runtime.Object{image}, testBuilds...),
					Args:           []string{imageName},
					ExpectedOutput: expectedOutput,
				}.TestKpack(t, cmdFunc)
			})

			when("the namespace has no images", func() {
				it("returns a message that the namespace has no images", func() {
					testhelpers.CommandTest{
						Args:           []string{imageName},
						ExpectErr:      true,
						ExpectedOutput: "Error: images.build.pivotal.io \"test-image\" not found\n",
					}.TestKpack(t, cmdFunc)

				})
			})
		})
	})
}

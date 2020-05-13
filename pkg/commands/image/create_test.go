package image_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	imgcmds "github.com/pivotal/build-service-cli/pkg/commands/image"
	"github.com/pivotal/build-service-cli/pkg/image"
	srcfakes "github.com/pivotal/build-service-cli/pkg/source/fakes"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestImageCreateCommand(t *testing.T) {
	spec.Run(t, "TestImageCreateCommand", testImageCreateCommand)
}

func testImageCreateCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	sourceUploader := &srcfakes.SourceUploader{
		ImageRef: "some-registry.io/some-repo-source:source-id",
	}

	imageFactory := &image.Factory{
		SourceUploader: sourceUploader,
	}

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return imgcmds.NewCreateCommand(clientSet, imageFactory, defaultNamespace)
	}

	when("a namespace is provided", func() {
		const namespace = "some-namespace"

		when("the image config is valid", func() {
			it("creates the image", func() {
				expectedImage := &v1alpha1.Image{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-image",
						Namespace: namespace,
					},
					Spec: v1alpha1.ImageSpec{
						Tag: "some-registry.io/some-repo",
						Builder: corev1.ObjectReference{
							Kind: expv1alpha1.CustomClusterBuilderKind,
							Name: "default",
						},
						ServiceAccount: "default",
						Source: v1alpha1.SourceConfig{
							Git: &v1alpha1.Git{
								URL:      "some-git-url",
								Revision: "some-git-rev",
							},
							SubPath: "some-sub-path",
						},
						Build: &v1alpha1.ImageBuild{
							Env: []corev1.EnvVar{
								{
									Name:  "some-key",
									Value: "some-val",
								},
							},
						},
					},
				}

				testhelpers.CommandTest{
					Args: []string{
						"some-image",
						"some-registry.io/some-repo",
						"--git", "some-git-url",
						"--git-revision", "some-git-rev",
						"--sub-path", "some-sub-path",
						"--env", "some-key=some-val",
						"-n", namespace,
					},
					ExpectedOutput: "\"some-image\" created\n",
					ExpectCreates: []runtime.Object{
						expectedImage,
					},
				}.TestKpack(t, cmdFunc)
			})
		})

		when("the image config is invalid", func() {
			it("returns an error", func() {
				testhelpers.CommandTest{
					Args: []string{
						"some-image",
						"some-registry.io/some-repo",
						"--blob", "some-blob",
						"--git", "some-git-url",
						"-n", namespace,
					},
					ExpectErr:      true,
					ExpectedOutput: "Error: image source must be one of git, blob, or local-path\n",
				}.TestKpack(t, cmdFunc)
			})
		})
	})

	when("a namespace is not provided", func() {
		when("the image config is valid", func() {
			it("creates the image", func() {
				expectedImage := &v1alpha1.Image{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-image",
						Namespace: defaultNamespace,
					},
					Spec: v1alpha1.ImageSpec{
						Tag: "some-registry.io/some-repo",
						Builder: corev1.ObjectReference{
							Kind: expv1alpha1.CustomClusterBuilderKind,
							Name: "default",
						},
						ServiceAccount: "default",
						Source: v1alpha1.SourceConfig{
							Git: &v1alpha1.Git{
								URL:      "some-git-url",
								Revision: "some-git-rev",
							},
							SubPath: "some-sub-path",
						},
						Build: &v1alpha1.ImageBuild{
							Env: []corev1.EnvVar{
								{
									Name:  "some-key",
									Value: "some-val",
								},
							},
						},
					},
				}

				testhelpers.CommandTest{
					Args: []string{
						"some-image",
						"some-registry.io/some-repo",
						"--git", "some-git-url",
						"--git-revision", "some-git-rev",
						"--sub-path", "some-sub-path",
						"--env", "some-key=some-val",
					},
					ExpectedOutput: "\"some-image\" created\n",
					ExpectCreates: []runtime.Object{
						expectedImage,
					},
				}.TestKpack(t, cmdFunc)
			})
		})

		when("the image config is invalid", func() {
			it("returns an error", func() {
				testhelpers.CommandTest{
					Args: []string{
						"some-image",
						"some-registry.io/some-repo",
						"--blob", "some-blob",
						"--git", "some-git-url",
					},
					ExpectErr:      true,
					ExpectedOutput: "Error: image source must be one of git, blob, or local-path\n",
				}.TestKpack(t, cmdFunc)
			})
		})
	})

	when("the image uses local source code", func() {
		it("uploads the source image and creates the image config", func() {
			expectedImage := &v1alpha1.Image{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-image",
					Namespace: defaultNamespace,
				},
				Spec: v1alpha1.ImageSpec{
					Tag: "some-registry.io/some-repo",
					Builder: corev1.ObjectReference{
						Kind: expv1alpha1.CustomClusterBuilderKind,
						Name: "default",
					},
					ServiceAccount: "default",
					Source: v1alpha1.SourceConfig{
						Registry: &v1alpha1.Registry{
							Image: "some-registry.io/some-repo-source:source-id",
						},
						SubPath: "some-sub-path",
					},
					Build: &v1alpha1.ImageBuild{
						Env: []corev1.EnvVar{
							{
								Name:  "some-key",
								Value: "some-val",
							},
						},
					},
				},
			}

			testhelpers.CommandTest{
				Args: []string{
					"some-image",
					"some-registry.io/some-repo",
					"--local-path", "some-local-path",
					"--sub-path", "some-sub-path",
					"--env", "some-key=some-val",
				},
				ExpectedOutput: "\"some-image\" created\n",
				ExpectCreates: []runtime.Object{
					expectedImage,
				},
			}.TestKpack(t, cmdFunc)
		})
	})

	when("the image uses a non-default builder", func() {
		it("uploads the source image and creates the image config", func() {
			expectedImage := &v1alpha1.Image{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-image",
					Namespace: defaultNamespace,
				},
				Spec: v1alpha1.ImageSpec{
					Tag: "some-registry.io/some-repo",
					Builder: corev1.ObjectReference{
						Kind:      expv1alpha1.CustomBuilderKind,
						Namespace: defaultNamespace,
						Name:      "some-builder",
					},
					ServiceAccount: "default",
					Source: v1alpha1.SourceConfig{
						Blob: &v1alpha1.Blob{
							URL: "some-blob",
						},
					},
					Build: &v1alpha1.ImageBuild{},
				},
			}

			testhelpers.CommandTest{
				Args: []string{
					"some-image",
					"some-registry.io/some-repo",
					"--blob", "some-blob",
					"--builder", "some-builder",
				},
				ExpectedOutput: "\"some-image\" created\n",
				ExpectCreates: []runtime.Object{
					expectedImage,
				},
			}.TestKpack(t, cmdFunc)
		})
	})

	when("the image uses a non-default cluster builder", func() {
		it("uploads the source image and creates the image config", func() {
			expectedImage := &v1alpha1.Image{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-image",
					Namespace: defaultNamespace,
				},
				Spec: v1alpha1.ImageSpec{
					Tag: "some-registry.io/some-repo",
					Builder: corev1.ObjectReference{
						Kind: expv1alpha1.CustomClusterBuilderKind,
						Name: "some-builder",
					},
					ServiceAccount: "default",
					Source: v1alpha1.SourceConfig{
						Blob: &v1alpha1.Blob{
							URL: "some-blob",
						},
					},
					Build: &v1alpha1.ImageBuild{},
				},
			}

			testhelpers.CommandTest{
				Args: []string{
					"some-image",
					"some-registry.io/some-repo",
					"--blob", "some-blob",
					"--cluster-builder", "some-builder",
				},
				ExpectedOutput: "\"some-image\" created\n",
				ExpectCreates: []runtime.Object{
					expectedImage,
				},
			}.TestKpack(t, cmdFunc)
		})
	})
}

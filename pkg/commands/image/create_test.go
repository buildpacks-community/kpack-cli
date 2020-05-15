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
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return imgcmds.NewCreateCommand(clientSetProvider, imageFactory)
	}

	when("a namespace is provided", func() {
		const namespace = "some-namespace"

		when("the image config is valid", func() {
			it("creates the image", func() {
				expectedImage := &v1alpha1.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "build.pivotal.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-image",
						Namespace: namespace,
						Annotations: map[string]string{
							"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Image","apiVersion":"build.pivotal.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"CustomClusterBuilder","name":"default"},"serviceAccount":"default","source":{"git":{"url":"some-git-url","revision":"some-git-rev"},"subPath":"some-sub-path"},"build":{"env":[{"name":"some-key","value":"some-val"}],"resources":{}}},"status":{}}`,
						},
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
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "build.pivotal.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-image",
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Image","apiVersion":"build.pivotal.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"CustomClusterBuilder","name":"default"},"serviceAccount":"default","source":{"git":{"url":"some-git-url","revision":"some-git-rev"},"subPath":"some-sub-path"},"build":{"env":[{"name":"some-key","value":"some-val"}],"resources":{}}},"status":{}}`,
						},
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
				TypeMeta: metav1.TypeMeta{
					Kind:       "Image",
					APIVersion: "build.pivotal.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-image",
					Namespace: defaultNamespace,
					Annotations: map[string]string{
						"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Image","apiVersion":"build.pivotal.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"CustomClusterBuilder","name":"default"},"serviceAccount":"default","source":{"registry":{"image":"some-registry.io/some-repo-source:source-id"},"subPath":"some-sub-path"},"build":{"env":[{"name":"some-key","value":"some-val"}],"resources":{}}},"status":{}}`,
					},
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
				TypeMeta: metav1.TypeMeta{
					Kind:       "Image",
					APIVersion: "build.pivotal.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-image",
					Namespace: defaultNamespace,
					Annotations: map[string]string{
						"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Image","apiVersion":"build.pivotal.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"CustomBuilder","namespace":"some-default-namespace","name":"some-builder"},"serviceAccount":"default","source":{"blob":{"url":"some-blob"}},"build":{"resources":{}}},"status":{}}`,
					},
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
				TypeMeta: metav1.TypeMeta{
					Kind:       "Image",
					APIVersion: "build.pivotal.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-image",
					Namespace: defaultNamespace,
					Annotations: map[string]string{
						"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Image","apiVersion":"build.pivotal.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"CustomClusterBuilder","name":"some-builder"},"serviceAccount":"default","source":{"blob":{"url":"some-blob"}},"build":{"resources":{}}},"status":{}}`,
					},
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

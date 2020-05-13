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

func TestImagePatchCommand(t *testing.T) {
	spec.Run(t, "TestImagePatchCommand", testImagePatchCommand)
}

func testImagePatchCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	sourceUploader := &srcfakes.SourceUploader{
		ImageRef: "",
	}

	patchFactory := &image.PatchFactory{
		SourceUploader: sourceUploader,
	}

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return imgcmds.NewPatchCommand(clientSet, patchFactory, defaultNamespace)
	}

	img := &v1alpha1.Image{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-image",
			Namespace: defaultNamespace,
		},
		Spec: v1alpha1.ImageSpec{
			Tag: "some-tag",
			Builder: corev1.ObjectReference{
				Kind: expv1alpha1.CustomClusterBuilderKind,
				Name: "some-ccb",
			},
			Source: v1alpha1.SourceConfig{
				Git: &v1alpha1.Git{
					URL:      "some-git-url",
					Revision: "some-revision",
				},
				SubPath: "some-path",
			},
			Build: &v1alpha1.ImageBuild{
				Env: []corev1.EnvVar{
					{
						Name:  "key1",
						Value: "value1",
					},
					{
						Name:  "key2",
						Value: "value2",
					},
				},
			},
		},
	}

	when("no parameters are provided", func() {
		it("does not create a patch", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					img,
				},
				Args: []string{
					"some-image",
				},
				ExpectedOutput: "nothing to patch\n",
			}.TestKpack(t, cmdFunc)
		})
	})

	when("patching source", func() {
		when("patching the sub path", func() {
			it("can patch it with an empty string", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						img,
					},
					Args: []string{
						"some-image",
						"--sub-path", "",
					},
					ExpectedOutput: "\"some-image\" patched\n",
					ExpectPatches: []string{
						`{"op":"remove","path":"/spec/source/subPath"}`,
					},
				}.TestKpack(t, cmdFunc)
			})

			it("can patch it with a non-empty string", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						img,
					},
					Args: []string{
						"some-image",
						"--sub-path", "a-new-path",
					},
					ExpectedOutput: "\"some-image\" patched\n",
					ExpectPatches: []string{
						`{"op":"replace","path":"/spec/source/subPath","value":"a-new-path"}`,
					},
				}.TestKpack(t, cmdFunc)
			})
		})

		it("can change source types", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					img,
				},
				Args: []string{
					"some-image",
					"--blob", "some-blob",
				},
				ExpectedOutput: "\"some-image\" patched\n",
				ExpectPatches: []string{
					`{"op":"add","path":"/spec/source/blob","value":{"url":"some-blob"}}`,
					`{"op":"remove","path":"/spec/source/git"}`,
				},
			}.TestKpack(t, cmdFunc)
		})

		it("can change git revision if existing source is git", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					img,
				},
				Args: []string{
					"some-image",
					"--git-revision", "some-new-revision",
				},
				ExpectedOutput: "\"some-image\" patched\n",
				ExpectPatches: []string{
					`{"op":"replace","path":"/spec/source/git/revision","value":"some-new-revision"}`,
				},
			}.TestKpack(t, cmdFunc)
		})

		it("git revision defaults to master if not provided with git", func() {
			img.Spec.Source = v1alpha1.SourceConfig{
				Blob: &v1alpha1.Blob{
					URL: "some-blob",
				},
			}

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					img,
				},
				Args: []string{
					"some-image",
					"--git", "some-new-git-url",
				},
				ExpectedOutput: "\"some-image\" patched\n",
				ExpectPatches: []string{
					`{"op":"add","path":"/spec/source/git","value":{"revision":"master","url":"some-new-git-url"}}`,
					`{"op":"remove","path":"/spec/source/blob"}`,
				},
			}.TestKpack(t, cmdFunc)
		})
	})

	when("patching the builder", func() {
		it("can patch the builder", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					img,
				},
				Args: []string{
					"some-image",
					"--builder", "some-builder",
				},
				ExpectedOutput: "\"some-image\" patched\n",
				ExpectPatches: []string{
					`{"op":"replace","path":"/spec/builder/kind","value":"CustomBuilder"}`,
					`{"op":"replace","path":"/spec/builder/name","value":"some-builder"}`,
					`{"op":"add","path":"/spec/builder/namespace","value":"some-default-namespace"}`,
				},
			}.TestKpack(t, cmdFunc)
		})
	})

	when("patching env vars", func() {
		it("can delete env vars", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					img,
				},
				Args: []string{
					"some-image",
					"-d", "key2",
				},
				ExpectedOutput: "\"some-image\" patched\n",
				ExpectPatches: []string{
					`{"op":"remove","path":"/spec/build/env/1"}`,
				},
			}.TestKpack(t, cmdFunc)
		})

		it("can update existing env vars", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					img,
				},
				Args: []string{
					"some-image",
					"-e", "key1=some-other-value",
				},
				ExpectedOutput: "\"some-image\" patched\n",
				ExpectPatches: []string{
					`{"op":"replace","path":"/spec/build/env/0/value","value":"some-other-value"}`,
				},
			}.TestKpack(t, cmdFunc)
		})

		it("can add new env vars", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					img,
				},
				Args: []string{
					"some-image",
					"-e", "key3=value3",
				},
				ExpectedOutput: "\"some-image\" patched\n",
				ExpectPatches: []string{
					`{"op":"add","path":"/spec/build/env/2","value":{"name":"key3","value":"value3"}}`,
				},
			}.TestKpack(t, cmdFunc)
		})
	})
}

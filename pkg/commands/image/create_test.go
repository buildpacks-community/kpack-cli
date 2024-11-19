// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	registryfakes "github.com/buildpacks-community/kpack-cli/pkg/registry/fakes"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	cmdFakes "github.com/buildpacks-community/kpack-cli/pkg/commands/fakes"
	imgcmds "github.com/buildpacks-community/kpack-cli/pkg/commands/image"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
)

func TestImageCreateCommand(t *testing.T) {
	spec.Run(t, "TestImageCreateCommand", testCreateCommand(imgcmds.NewCreateCommand))
}

func setLastAppliedAnnotation(i *v1alpha2.Image) error {
	lastApplied, err := json.Marshal(i)
	if err != nil {
		return err
	}
	i.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = string(lastApplied)
	return nil
}

func testCreateCommand(imageCommand func(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider, newImageWaiter func(k8s.ClientSet) imgcmds.ImageWaiter) *cobra.Command) func(t *testing.T, when spec.G, it spec.S) {
	return func(t *testing.T, when spec.G, it spec.S) {
		const defaultNamespace = "some-default-namespace"

		registryUtilProvider := registryfakes.UtilProvider{}
		fakeImageWaiter := &cmdFakes.FakeImageWaiter{}

		cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
			return imageCommand(clientSetProvider, registryUtilProvider, func(set k8s.ClientSet) imgcmds.ImageWaiter {
				return fakeImageWaiter
			})
		}

		when("a namespace is provided", func() {
			const namespace = "some-namespace"

			when("the image config is valid", func() {
				cacheSize := resource.MustParse("2G")
				expectedImage := &v1alpha2.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "kpack.io/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:        "some-image",
						Namespace:   namespace,
						Annotations: map[string]string{},
					},
					Spec: v1alpha2.ImageSpec{
						Tag: "some-registry.io/some-repo",
						AdditionalTags: []string{
							"some-registry.io/some-tag",
							"some-registry.io/some-other-tag",
						},
						Builder: corev1.ObjectReference{
							Kind: v1alpha2.ClusterBuilderKind,
							Name: "default",
						},
						ServiceAccountName: "default",
						Source: corev1alpha1.SourceConfig{
							Git: &corev1alpha1.Git{
								URL:      "some-git-url",
								Revision: "some-git-rev",
							},
							SubPath: "some-sub-path",
						},
						Build: &v1alpha2.ImageBuild{
							Env: []corev1.EnvVar{
								{
									Name:  "some-key",
									Value: "some-val",
								},
							},
							Services: v1alpha2.Services{
								{
									APIVersion: "v1",
									Kind:       "SomeResource",
									Name:       "some-binding",
								},
							},
						},
						Cache: &v1alpha2.ImageCacheConfig{
							Volume: &v1alpha2.ImagePersistentVolumeCache{
								Size: &cacheSize,
							},
						},
					},
				}

				it("creates the image and wait on the image", func() {
					require.NoError(t, setLastAppliedAnnotation(expectedImage))
					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--additional-tag", "some-registry.io/some-tag",
							"--additional-tag", "some-registry.io/some-other-tag",
							"--git", "some-git-url",
							"--git-revision", "some-git-rev",
							"--sub-path", "some-sub-path",
							"--env", "some-key=some-val",
							"--service-binding", "SomeResource:v1:some-binding",
							"--cache-size", "2G",
							"-n", namespace,
							"--registry-ca-cert-path", "some-cert-path",
							"--registry-verify-certs",
							"--wait",
						},
						ExpectedOutput: `Creating Image Resource...
Image Resource "some-image" created
`,
						ExpectCreates: []runtime.Object{
							expectedImage,
						},
					}.TestKpack(t, cmdFunc)

					assert.Len(t, fakeImageWaiter.Calls, 1)
					assert.Equal(t, fakeImageWaiter.Calls[0], expectedImage)
				})

				it("defaults the git revision to main", func() {
					expectedImage.Spec.Source.Git.Revision = "main"
					require.NoError(t, setLastAppliedAnnotation(expectedImage))

					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--additional-tag", "some-registry.io/some-tag",
							"--additional-tag", "some-registry.io/some-other-tag",
							"--git", "some-git-url",
							"--sub-path", "some-sub-path",
							"--env", "some-key=some-val",
							"--service-binding", "SomeResource:v1:some-binding",
							"--cache-size", "2G",
							"-n", namespace,
						},
						ExpectedOutput: `Creating Image Resource...
Image Resource "some-image" created
`,
						ExpectCreates: []runtime.Object{
							expectedImage,
						},
					}.TestKpack(t, cmdFunc)
				})

				it("defaults the service binding kind to Secret", func() {
					expectedImage.Spec.Build.Services = v1alpha2.Services{
						{
							Kind: "Secret",
							Name: "some-secret",
						},
					}
					require.NoError(t, setLastAppliedAnnotation(expectedImage))

					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--additional-tag", "some-registry.io/some-tag",
							"--additional-tag", "some-registry.io/some-other-tag",
							"--git", "some-git-url",
							"--git-revision", "some-git-rev",
							"--sub-path", "some-sub-path",
							"--env", "some-key=some-val",
							"--service-binding", "some-secret",
							"--cache-size", "2G",
							"-n", namespace,
						},
						ExpectedOutput: `Creating Image Resource...
Image Resource "some-image" created
`,
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
							"--tag", "some-registry.io/some-repo",
							"--blob", "some-blob",
							"--git", "some-git-url",
							"-n", namespace,
						},
						ExpectErr:           true,
						ExpectedOutput:      "Creating Image Resource...\n",
						ExpectedErrorOutput: "Error: image source must be one of git, blob, or local-path\n",
					}.TestKpack(t, cmdFunc)

					assert.Len(t, fakeImageWaiter.Calls, 0)
				})
			})
		})

		when("a namespace is not provided", func() {
			when("the image config is valid", func() {
				it("creates the image", func() {
					expectedImage := &v1alpha2.Image{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Image",
							APIVersion: "kpack.io/v1alpha2",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:        "some-image",
							Namespace:   defaultNamespace,
							Annotations: map[string]string{},
						},
						Spec: v1alpha2.ImageSpec{
							Tag: "some-registry.io/some-repo",
							Builder: corev1.ObjectReference{
								Kind: v1alpha2.ClusterBuilderKind,
								Name: "default",
							},
							ServiceAccountName: "default",
							Source: corev1alpha1.SourceConfig{
								Git: &corev1alpha1.Git{
									URL:      "some-git-url",
									Revision: "some-git-rev",
								},
								SubPath: "some-sub-path",
							},
							Build: &v1alpha2.ImageBuild{
								Env: []corev1.EnvVar{
									{
										Name:  "some-key",
										Value: "some-val",
									},
								},
								Services: v1alpha2.Services{
									{
										APIVersion: "v1",
										Kind:       "SomeResource",
										Name:       "some-binding",
									},
								},
							},
						},
					}

					require.NoError(t, setLastAppliedAnnotation(expectedImage))
					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--git", "some-git-url",
							"--git-revision", "some-git-rev",
							"--sub-path", "some-sub-path",
							"--env", "some-key=some-val",
							"--service-binding", "SomeResource:v1:some-binding",
						},
						ExpectedOutput: `Creating Image Resource...
Image Resource "some-image" created
`,
						ExpectCreates: []runtime.Object{
							expectedImage,
						},
					}.TestKpack(t, cmdFunc)

					assert.Len(t, fakeImageWaiter.Calls, 0)
				})
			})

			when("the image config is invalid", func() {
				it("returns an error", func() {
					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--blob", "some-blob",
							"--git", "some-git-url",
						},
						ExpectErr:           true,
						ExpectedOutput:      "Creating Image Resource...\n",
						ExpectedErrorOutput: "Error: image source must be one of git, blob, or local-path\n",
					}.TestKpack(t, cmdFunc)

					assert.Len(t, fakeImageWaiter.Calls, 0)
				})
			})
		})

		when("build history limits are provided", func() {
			when("the image config is valid", func() {
				it("creates the image with build history limits", func() {
					buildHistoryLimit := int64(5)
					defaultBuildHistoryLimit := &buildHistoryLimit

					expectedImage := &v1alpha2.Image{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Image",
							APIVersion: "kpack.io/v1alpha2",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:        "some-image",
							Namespace:   defaultNamespace,
							Annotations: map[string]string{},
						},
						Spec: v1alpha2.ImageSpec{
							Tag: "some-registry.io/some-repo",
							Builder: corev1.ObjectReference{
								Kind: v1alpha2.ClusterBuilderKind,
								Name: "default",
							},
							ServiceAccountName: "default",
							Source: corev1alpha1.SourceConfig{
								Git: &corev1alpha1.Git{
									URL:      "some-git-url",
									Revision: "some-git-rev",
								},
								SubPath: "some-sub-path",
							},
							SuccessBuildHistoryLimit: defaultBuildHistoryLimit,
							FailedBuildHistoryLimit:  defaultBuildHistoryLimit,
							Build: &v1alpha2.ImageBuild{
								Env: []corev1.EnvVar{
									{
										Name:  "some-key",
										Value: "some-val",
									},
								},
								Services: v1alpha2.Services{
									{
										APIVersion: "v1",
										Kind:       "SomeResource",
										Name:       "some-binding",
									},
								},
							},
						},
					}

					require.NoError(t, setLastAppliedAnnotation(expectedImage))
					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--git", "some-git-url",
							"--git-revision", "some-git-rev",
							"--sub-path", "some-sub-path",
							"--env", "some-key=some-val",
							"--success-build-history-limit", "5",
							"--failed-build-history-limit", "5",
							"--service-binding", "SomeResource:v1:some-binding",
						},
						ExpectedOutput: `Creating Image Resource...
Image Resource "some-image" created
`,
						ExpectCreates: []runtime.Object{
							expectedImage,
						},
					}.TestKpack(t, cmdFunc)

					assert.Len(t, fakeImageWaiter.Calls, 0)
				})
			})

			when("the image config is invalid", func() {
				it("returns an error", func() {
					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--blob", "some-blob",
							"--git", "some-git-url",
						},
						ExpectErr:           true,
						ExpectedOutput:      "Creating Image Resource...\n",
						ExpectedErrorOutput: "Error: image source must be one of git, blob, or local-path\n",
					}.TestKpack(t, cmdFunc)

					assert.Len(t, fakeImageWaiter.Calls, 0)
				})
			})
		})

		when("the image uses local source code", func() {
			it("uploads the source image and creates the image config", func() {
				expectedImage := &v1alpha2.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "kpack.io/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:        "some-image",
						Namespace:   defaultNamespace,
						Annotations: map[string]string{},
					},
					Spec: v1alpha2.ImageSpec{
						Tag: "some-registry.io/some-repo",
						Builder: corev1.ObjectReference{
							Kind: v1alpha2.ClusterBuilderKind,
							Name: "default",
						},
						ServiceAccountName: "default",
						Source: corev1alpha1.SourceConfig{
							Registry: &corev1alpha1.Registry{
								Image: "some-registry.io/some-repo-source:source-id",
							},
							SubPath: "some-sub-path",
						},
						Build: &v1alpha2.ImageBuild{
							Env: []corev1.EnvVar{
								{
									Name:  "some-key",
									Value: "some-val",
								},
							},
							Services: v1alpha2.Services{
								{
									APIVersion: "v1",
									Kind:       "SomeResource",
									Name:       "some-binding",
								},
							},
						},
					},
				}
				require.NoError(t, setLastAppliedAnnotation(expectedImage))

				testhelpers.CommandTest{
					Args: []string{
						"some-image",
						"--tag", "some-registry.io/some-repo",
						"--local-path", "some-local-path",
						"--sub-path", "some-sub-path",
						"--env", "some-key=some-val",
						"--service-binding", "SomeResource:v1:some-binding",
					},
					ExpectedOutput: `Creating Image Resource...
Uploading to 'some-registry.io/some-repo-source'...
	Uploading 'some-registry.io/some-repo-source:source-id'
Image Resource "some-image" created
`,
					ExpectCreates: []runtime.Object{
						expectedImage,
					},
				}.TestKpack(t, cmdFunc)

				assert.Len(t, fakeImageWaiter.Calls, 0)
			})
			it("sets the repository source path on the image", func() {
				expectedImage := &v1alpha2.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "kpack.io/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:        "some-image",
						Namespace:   defaultNamespace,
						Annotations: map[string]string{},
					},
					Spec: v1alpha2.ImageSpec{
						Tag: "some-registry.io/some-repo",
						Builder: corev1.ObjectReference{
							Kind: v1alpha2.ClusterBuilderKind,
							Name: "default",
						},
						ServiceAccountName: "default",
						Source: corev1alpha1.SourceConfig{
							Registry: &corev1alpha1.Registry{
								Image: "some-registry.io/some-repo-testing:source-id",
							},
							SubPath: "some-sub-path",
						},
						Build: &v1alpha2.ImageBuild{
							Env: []corev1.EnvVar{
								{
									Name:  "some-key",
									Value: "some-val",
								},
							},
							Services: v1alpha2.Services{
								{
									APIVersion: "v1",
									Kind:       "SomeResource",
									Name:       "some-binding",
								},
							},
						},
					},
				}
				require.NoError(t, setLastAppliedAnnotation(expectedImage))

				testhelpers.CommandTest{
					Args: []string{
						"some-image",
						"--tag", "some-registry.io/some-repo",
						"--local-path", "some-local-path",
						"--local-path-destination-image", "some-registry.io/some-repo-testing",
						"--sub-path", "some-sub-path",
						"--env", "some-key=some-val",
						"--service-binding", "SomeResource:v1:some-binding",
					},
					ExpectedOutput: `Creating Image Resource...
Uploading to 'some-registry.io/some-repo-testing'...
	Uploading 'some-registry.io/some-repo-testing:source-id'
Image Resource "some-image" created
`,
					ExpectCreates: []runtime.Object{
						expectedImage,
					},
				}.TestKpack(t, cmdFunc)

				assert.Len(t, fakeImageWaiter.Calls, 0)
			})
		})

		when("the image uses a non-default builder", func() {
			it("uploads the source image and creates the image config", func() {
				buildHistoryLimit := int64(10)
				defaultBuildHistoryLimit := &buildHistoryLimit

				expectedImage := &v1alpha2.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "kpack.io/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:        "some-image",
						Namespace:   defaultNamespace,
						Annotations: map[string]string{},
					},
					Spec: v1alpha2.ImageSpec{
						Tag: "some-registry.io/some-repo",
						Builder: corev1.ObjectReference{
							Kind:      v1alpha2.BuilderKind,
							Namespace: defaultNamespace,
							Name:      "some-builder",
						},
						ServiceAccountName: "default",
						Source: corev1alpha1.SourceConfig{
							Blob: &corev1alpha1.Blob{
								URL: "some-blob",
							},
						},
						Build:                    &v1alpha2.ImageBuild{},
						SuccessBuildHistoryLimit: defaultBuildHistoryLimit,
						FailedBuildHistoryLimit:  defaultBuildHistoryLimit,
					},
				}
				require.NoError(t, setLastAppliedAnnotation(expectedImage))

				testhelpers.CommandTest{
					Args: []string{
						"some-image",
						"--tag", "some-registry.io/some-repo",
						"--blob", "some-blob",
						"--builder", "some-builder",
						"--success-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
						"--failed-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
					},
					ExpectedOutput: `Creating Image Resource...
Image Resource "some-image" created
`,
					ExpectCreates: []runtime.Object{
						expectedImage,
					},
				}.TestKpack(t, cmdFunc)

				assert.Len(t, fakeImageWaiter.Calls, 0)
			})
		})

		when("the image uses a non-default cluster builder", func() {
			it("uploads the source image and creates the image config", func() {
				buildHistoryLimit := int64(5)
				defaultBuildHistoryLimit := &buildHistoryLimit

				expectedImage := &v1alpha2.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "kpack.io/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:        "some-image",
						Namespace:   defaultNamespace,
						Annotations: map[string]string{},
					},
					Spec: v1alpha2.ImageSpec{
						Tag: "some-registry.io/some-repo",
						Builder: corev1.ObjectReference{
							Kind: v1alpha2.ClusterBuilderKind,
							Name: "some-builder",
						},
						ServiceAccountName: "default",
						Source: corev1alpha1.SourceConfig{
							Blob: &corev1alpha1.Blob{
								URL: "some-blob",
							},
						},
						Build:                    &v1alpha2.ImageBuild{},
						SuccessBuildHistoryLimit: defaultBuildHistoryLimit,
						FailedBuildHistoryLimit:  defaultBuildHistoryLimit,
					},
				}
				require.NoError(t, setLastAppliedAnnotation(expectedImage))

				testhelpers.CommandTest{
					Args: []string{
						"some-image",
						"--tag", "some-registry.io/some-repo",
						"--blob", "some-blob",
						"--cluster-builder", "some-builder",
						"--success-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
						"--failed-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
					},
					ExpectedOutput: `Creating Image Resource...
Image Resource "some-image" created
`,
					ExpectCreates: []runtime.Object{
						expectedImage,
					},
				}.TestKpack(t, cmdFunc)

				assert.Len(t, fakeImageWaiter.Calls, 0)
			})
		})

		it("can use a non-default service account", func() {
			expectedImage := &v1alpha2.Image{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Image",
					APIVersion: "kpack.io/v1alpha2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "some-image",
					Namespace:   defaultNamespace,
					Annotations: map[string]string{},
				},
				Spec: v1alpha2.ImageSpec{
					AdditionalTags: []string{},
					Tag:            "some-registry.io/some-repo",
					Builder: corev1.ObjectReference{
						Kind: v1alpha2.ClusterBuilderKind,
						Name: "default",
					},
					ServiceAccountName: "some-sa",
					Source: corev1alpha1.SourceConfig{
						Git: &corev1alpha1.Git{
							URL:      "some-git-url",
							Revision: "some-git-rev",
						},
					},
					Build: &v1alpha2.ImageBuild{},
				},
			}
			require.NoError(t, setLastAppliedAnnotation(expectedImage))

			testhelpers.CommandTest{
				Args: []string{
					"some-image",
					"--tag", "some-registry.io/some-repo",
					"--git", "some-git-url",
					"--git-revision", "some-git-rev",
					"--service-account", "some-sa",
					"--wait",
				},
				ExpectedOutput: `Creating Image Resource...
Image Resource "some-image" created
`,
				ExpectCreates: []runtime.Object{
					expectedImage,
				},
			}.TestKpack(t, cmdFunc)

			assert.Len(t, fakeImageWaiter.Calls, 1)
			assert.Equal(t, fakeImageWaiter.Calls[0], expectedImage)
		})

		when("output flag is used", func() {
			when("the image config is invalid", func() {
				it("returns an error", func() {
					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--blob", "some-blob",
							"--git", "some-git-url",
						},
						ExpectErr:           true,
						ExpectedOutput:      "Creating Image Resource...\n",
						ExpectedErrorOutput: "Error: image source must be one of git, blob, or local-path\n",
					}.TestKpack(t, cmdFunc)
					assert.Len(t, fakeImageWaiter.Calls, 0)
				})
			})

			when("the image config is valid", func() {
				buildHistoryLimit := int64(10)
				defaultBuildHistoryLimit := &buildHistoryLimit

				expectedImage := &v1alpha2.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "kpack.io/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:        "some-image",
						Namespace:   defaultNamespace,
						Annotations: map[string]string{},
					},
					Spec: v1alpha2.ImageSpec{
						Tag: "some-registry.io/some-repo",
						Builder: corev1.ObjectReference{
							Kind: v1alpha2.ClusterBuilderKind,
							Name: "default",
						},
						ServiceAccountName: "default",
						Source: corev1alpha1.SourceConfig{
							Git: &corev1alpha1.Git{
								URL:      "some-git-url",
								Revision: "some-git-rev",
							},
							SubPath: "some-sub-path",
						},
						FailedBuildHistoryLimit:  defaultBuildHistoryLimit,
						SuccessBuildHistoryLimit: defaultBuildHistoryLimit,
						Build: &v1alpha2.ImageBuild{
							Env: []corev1.EnvVar{
								{
									Name:  "some-key",
									Value: "some-val",
								},
							},
							Services: v1alpha2.Services{
								{
									APIVersion: "v1",
									Kind:       "SomeResource",
									Name:       "some-binding",
								},
							},
						},
					},
				}

				it("can output in yaml format and does not wait", func() {
					require.NoError(t, setLastAppliedAnnotation(expectedImage))
					resourceYAML, err := getTestFile("./testdata/test-image.yaml")
					if err != nil {
						t.Fatalf("unable to convert test file to string: %v", err)
					}

					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--git", "some-git-url",
							"--git-revision", "some-git-rev",
							"--sub-path", "some-sub-path",
							"--env", "some-key=some-val",
							"--service-binding", "SomeResource:v1:some-binding",
							"--success-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
							"--failed-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
							"--output", "yaml",
							"--wait",
						},
						ExpectedOutput: resourceYAML,
						ExpectedErrorOutput: `Creating Image Resource...
`,
						ExpectCreates: []runtime.Object{
							expectedImage,
						},
					}.TestKpack(t, cmdFunc)
					assert.Len(t, fakeImageWaiter.Calls, 0)
				})

				it("can output in json format and does not wait", func() {
					buildHistoryLimit := int64(10)

					require.NoError(t, setLastAppliedAnnotation(expectedImage))
					resourceJSON, err := getTestFile("./testdata/test-image.json")
					if err != nil {
						t.Fatalf("unable to convert test file to string: %v", err)
					}

					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--git", "some-git-url",
							"--git-revision", "some-git-rev",
							"--sub-path", "some-sub-path",
							"--env", "some-key=some-val",
							"--service-binding", "SomeResource:v1:some-binding",
							"--success-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
							"--failed-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
							"--output", "json",
							"--wait",
						},
						ExpectedOutput: resourceJSON,
						ExpectedErrorOutput: `Creating Image Resource...
`,
						ExpectCreates: []runtime.Object{
							expectedImage,
						},
					}.TestKpack(t, cmdFunc)
					assert.Len(t, fakeImageWaiter.Calls, 0)
				})
			})
		})

		when("dry-run flag is used", func() {
			when("the image config is invalid", func() {
				it("returns an error", func() {
					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--blob", "some-blob",
							"--git", "some-git-url",
						},
						ExpectErr:           true,
						ExpectedOutput:      "Creating Image Resource...\n",
						ExpectedErrorOutput: "Error: image source must be one of git, blob, or local-path\n",
					}.TestKpack(t, cmdFunc)
				})
			})

			when("the image config is valid", func() {
				it("does not creates an image and prints result message with dry run indicated", func() {
					buildHistoryLimit := int64(10)
					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--git", "some-git-url",
							"--git-revision", "some-git-rev",
							"--sub-path", "some-sub-path",
							"--service-binding", "SomeResource:v1:some-binding",
							"--env", "some-key=some-val",
							"--success-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
							"--failed-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
							"--dry-run",
							"--wait",
						},
						ExpectedOutput: `Creating Image Resource... (dry run)
Image Resource "some-image" created (dry run)
`,
					}.TestKpack(t, cmdFunc)
					assert.Len(t, fakeImageWaiter.Calls, 0)
				})

				when("output flag is used", func() {
					it("does not create an image and prints the resource output", func() {
						resourceYAML, err := getTestFile("./testdata/test-image.yaml")
						if err != nil {
							t.Fatalf("unable to convert test file to string: %v", err)
						}
						buildHistoryLimit := int64(10)

						testhelpers.CommandTest{
							Args: []string{
								"some-image",
								"--tag", "some-registry.io/some-repo",
								"--git", "some-git-url",
								"--git-revision", "some-git-rev",
								"--sub-path", "some-sub-path",
								"--env", "some-key=some-val",
								"--service-binding", "SomeResource:v1:some-binding",
								"--success-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
								"--failed-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
								"--output", "yaml",
								"--dry-run",
								"--wait",
							},
							ExpectedOutput: resourceYAML,
							ExpectedErrorOutput: `Creating Image Resource... (dry run)
`,
						}.TestKpack(t, cmdFunc)
						assert.Len(t, fakeImageWaiter.Calls, 0)
					})
				})
			})
		})

		when("dry-run-with-image-upload flag is used", func() {
			when("the image config is invalid", func() {
				it("returns an error", func() {
					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--blob", "some-blob",
							"--git", "some-git-url",
							"--dry-run-with-image-upload",
						},
						ExpectErr:           true,
						ExpectedOutput:      "Creating Image Resource... (dry run with image upload)\n",
						ExpectedErrorOutput: "Error: image source must be one of git, blob, or local-path\n",
					}.TestKpack(t, cmdFunc)
				})
			})

			when("the image config is valid", func() {
				it("does not creates an image and prints result message with dry run indicated", func() {
					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--local-path", "some-local-path",
							"--sub-path", "some-sub-path",
							"--env", "some-key=some-val",
							"--service-binding", "SomeResource:v1:some-binding",
							"--dry-run-with-image-upload",
							"--wait",
						},
						ExpectedOutput: `Creating Image Resource... (dry run with image upload)
Uploading to 'some-registry.io/some-repo-source'... (dry run with image upload)
	Uploading 'some-registry.io/some-repo-source:source-id'
Image Resource "some-image" created (dry run with image upload)
`,
					}.TestKpack(t, cmdFunc)
					assert.Len(t, fakeImageWaiter.Calls, 0)
				})

				when("output flag is used", func() {
					it("does not create an image and prints the resource output", func() {
						resourceYAML, err := getTestFile("./testdata/test-image.yaml")
						if err != nil {
							t.Fatalf("unable to convert test file to string: %v", err)
						}
						buildHistoryLimit := int64(10)

						testhelpers.CommandTest{
							Args: []string{
								"some-image",
								"--tag", "some-registry.io/some-repo",
								"--git", "some-git-url",
								"--git-revision", "some-git-rev",
								"--sub-path", "some-sub-path",
								"--env", "some-key=some-val",
								"--service-binding", "SomeResource:v1:some-binding",
								"--success-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
								"--failed-build-history-limit", strconv.FormatInt(buildHistoryLimit, 10),
								"--output", "yaml",
								"--dry-run-with-image-upload",
								"--wait",
							},
							ExpectedOutput: resourceYAML,
							ExpectedErrorOutput: `Creating Image Resource... (dry run with image upload)
`,
						}.TestKpack(t, cmdFunc)
						assert.Len(t, fakeImageWaiter.Calls, 0)
					})
				})
			})
		})

		when("cache size is not provided", func() {
			it("does not set an empty field on the image", func() {
				// note: this is to allow for defaults to be set by kpack webhook

				expectedImage := &v1alpha2.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "kpack.io/v1alpha2",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:        "some-image",
						Namespace:   defaultNamespace,
						Annotations: map[string]string{},
					},
					Spec: v1alpha2.ImageSpec{
						Tag: "some-registry.io/some-repo",
						Builder: corev1.ObjectReference{
							Kind: v1alpha2.ClusterBuilderKind,
							Name: "default",
						},
						ServiceAccountName: "default",
						Source: corev1alpha1.SourceConfig{
							Git: &corev1alpha1.Git{
								URL:      "some-git-url",
								Revision: "some-git-rev",
							},
							SubPath: "some-sub-path",
						},
						Build: &v1alpha2.ImageBuild{
							Env: []corev1.EnvVar{
								{
									Name:  "some-key",
									Value: "some-val",
								},
							},
							Services: v1alpha2.Services{
								{
									APIVersion: "v1",
									Kind:       "SomeResource",
									Name:       "some-binding",
								},
							},
						},
					},
				}

				require.NoError(t, setLastAppliedAnnotation(expectedImage))
				testhelpers.CommandTest{
					Args: []string{
						"some-image",
						"--tag", "some-registry.io/some-repo",
						"--git", "some-git-url",
						"--git-revision", "some-git-rev",
						"--sub-path", "some-sub-path",
						"--env", "some-key=some-val",
						"--service-binding", "SomeResource:v1:some-binding",
					},
					ExpectedOutput: `Creating Image Resource...
Image Resource "some-image" created
`,
					ExpectCreates: []runtime.Object{
						expectedImage,
					},
				}.TestKpack(t, cmdFunc)

				assert.Len(t, fakeImageWaiter.Calls, 0)
			})
		})
	}
}

func getTestFile(testfile string) (string, error) {
	b, err := os.ReadFile(testfile)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

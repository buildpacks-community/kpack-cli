// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"encoding/json"
	"testing"

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

	cmdFakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	imgcmds "github.com/vmware-tanzu/kpack-cli/pkg/commands/image"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
	registryfakes "github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
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
							Cache: &v1alpha2.ImageCacheConfig{
								Volume: &v1alpha2.ImagePersistentVolumeCache{},
							},
							Build: &v1alpha2.ImageBuild{
								Env: []corev1.EnvVar{
									{
										Name:  "some-key",
										Value: "some-val",
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
						Cache: &v1alpha2.ImageCacheConfig{
							Volume: &v1alpha2.ImagePersistentVolumeCache{},
						},
						Build: &v1alpha2.ImageBuild{
							Env: []corev1.EnvVar{
								{
									Name:  "some-key",
									Value: "some-val",
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
		})

		when("the image uses a non-default builder", func() {
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
						Cache: &v1alpha2.ImageCacheConfig{
							Volume: &v1alpha2.ImagePersistentVolumeCache{},
						},
						Build: &v1alpha2.ImageBuild{},
					},
				}
				require.NoError(t, setLastAppliedAnnotation(expectedImage))

				testhelpers.CommandTest{
					Args: []string{
						"some-image",
						"--tag", "some-registry.io/some-repo",
						"--blob", "some-blob",
						"--builder", "some-builder",
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
						Cache: &v1alpha2.ImageCacheConfig{
							Volume: &v1alpha2.ImagePersistentVolumeCache{},
						},
						Build: &v1alpha2.ImageBuild{},
					},
				}
				require.NoError(t, setLastAppliedAnnotation(expectedImage))

				testhelpers.CommandTest{
					Args: []string{
						"some-image",
						"--tag", "some-registry.io/some-repo",
						"--blob", "some-blob",
						"--cluster-builder", "some-builder",
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
					Cache: &v1alpha2.ImageCacheConfig{
						Volume: &v1alpha2.ImagePersistentVolumeCache{},
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
						Cache: &v1alpha2.ImageCacheConfig{
							Volume: &v1alpha2.ImagePersistentVolumeCache{},
						},
						Build: &v1alpha2.ImageBuild{
							Env: []corev1.EnvVar{
								{
									Name:  "some-key",
									Value: "some-val",
								},
							},
						},
					},
				}

				it("can output in yaml format and does not wait", func() {
					require.NoError(t, setLastAppliedAnnotation(expectedImage))
					const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Image
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"Image","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"ClusterBuilder","name":"default"},"serviceAccountName":"default","source":{"git":{"url":"some-git-url","revision":"some-git-rev"},"subPath":"some-sub-path"},"cache":{"volume":{}},"build":{"env":[{"name":"some-key","value":"some-val"}],"resources":{}}},"status":{}}'
  creationTimestamp: null
  name: some-image
  namespace: some-default-namespace
spec:
  build:
    env:
    - name: some-key
      value: some-val
    resources: {}
  builder:
    kind: ClusterBuilder
    name: default
  cache:
    volume: {}
  serviceAccountName: default
  source:
    git:
      revision: some-git-rev
      url: some-git-url
    subPath: some-sub-path
  tag: some-registry.io/some-repo
status: {}
`

					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--git", "some-git-url",
							"--git-revision", "some-git-rev",
							"--sub-path", "some-sub-path",
							"--env", "some-key=some-val",
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
					require.NoError(t, setLastAppliedAnnotation(expectedImage))
					const resourceJSON = `{
    "kind": "Image",
    "apiVersion": "kpack.io/v1alpha2",
    "metadata": {
        "name": "some-image",
        "namespace": "some-default-namespace",
        "creationTimestamp": null,
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"Image\",\"apiVersion\":\"kpack.io/v1alpha2\",\"metadata\":{\"name\":\"some-image\",\"namespace\":\"some-default-namespace\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"some-registry.io/some-repo\",\"builder\":{\"kind\":\"ClusterBuilder\",\"name\":\"default\"},\"serviceAccountName\":\"default\",\"source\":{\"git\":{\"url\":\"some-git-url\",\"revision\":\"some-git-rev\"},\"subPath\":\"some-sub-path\"},\"cache\":{\"volume\":{}},\"build\":{\"env\":[{\"name\":\"some-key\",\"value\":\"some-val\"}],\"resources\":{}}},\"status\":{}}"
        }
    },
    "spec": {
        "tag": "some-registry.io/some-repo",
        "builder": {
            "kind": "ClusterBuilder",
            "name": "default"
        },
        "serviceAccountName": "default",
        "source": {
            "git": {
                "url": "some-git-url",
                "revision": "some-git-rev"
            },
            "subPath": "some-sub-path"
        },
        "cache": {
            "volume": {}
        },
        "build": {
            "env": [
                {
                    "name": "some-key",
                    "value": "some-val"
                }
            ],
            "resources": {}
        }
    },
    "status": {}
}
`

					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--git", "some-git-url",
							"--git-revision", "some-git-rev",
							"--sub-path", "some-sub-path",
							"--env", "some-key=some-val",
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
					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--git", "some-git-url",
							"--git-revision", "some-git-rev",
							"--sub-path", "some-sub-path",
							"--env", "some-key=some-val",
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
						const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Image
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"Image","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"ClusterBuilder","name":"default"},"serviceAccountName":"default","source":{"git":{"url":"some-git-url","revision":"some-git-rev"},"subPath":"some-sub-path"},"cache":{"volume":{}},"build":{"env":[{"name":"some-key","value":"some-val"}],"resources":{}}},"status":{}}'
  creationTimestamp: null
  name: some-image
  namespace: some-default-namespace
spec:
  build:
    env:
    - name: some-key
      value: some-val
    resources: {}
  builder:
    kind: ClusterBuilder
    name: default
  cache:
    volume: {}
  serviceAccountName: default
  source:
    git:
      revision: some-git-rev
      url: some-git-url
    subPath: some-sub-path
  tag: some-registry.io/some-repo
status: {}
`

						testhelpers.CommandTest{
							Args: []string{
								"some-image",
								"--tag", "some-registry.io/some-repo",
								"--git", "some-git-url",
								"--git-revision", "some-git-rev",
								"--sub-path", "some-sub-path",
								"--env", "some-key=some-val",
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
						const resourceYAML = `apiVersion: kpack.io/v1alpha2
kind: Image
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"Image","apiVersion":"kpack.io/v1alpha2","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"ClusterBuilder","name":"default"},"serviceAccountName":"default","source":{"git":{"url":"some-git-url","revision":"some-git-rev"},"subPath":"some-sub-path"},"cache":{"volume":{}},"build":{"env":[{"name":"some-key","value":"some-val"}],"resources":{}}},"status":{}}'
  creationTimestamp: null
  name: some-image
  namespace: some-default-namespace
spec:
  build:
    env:
    - name: some-key
      value: some-val
    resources: {}
  builder:
    kind: ClusterBuilder
    name: default
  cache:
    volume: {}
  serviceAccountName: default
  source:
    git:
      revision: some-git-rev
      url: some-git-url
    subPath: some-sub-path
  tag: some-registry.io/some-repo
status: {}
`

						testhelpers.CommandTest{
							Args: []string{
								"some-image",
								"--tag", "some-registry.io/some-repo",
								"--git", "some-git-url",
								"--git-revision", "some-git-rev",
								"--sub-path", "some-sub-path",
								"--env", "some-key=some-val",
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
	}
}

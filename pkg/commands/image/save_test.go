// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	imgcmds "github.com/pivotal/build-service-cli/pkg/commands/image"
	"github.com/pivotal/build-service-cli/pkg/image"
	"github.com/pivotal/build-service-cli/pkg/image/fakes"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	srcfakes "github.com/pivotal/build-service-cli/pkg/source/fakes"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestImageSaveCommand(t *testing.T) {
	spec.Run(t, "TestImageSaveCommand", testImageSaveCommand)
}

func testImageSaveCommand(t *testing.T, when spec.G, it spec.S) {
	when("creating", func() {
		const defaultNamespace = "some-default-namespace"

		sourceUploader := &srcfakes.SourceUploader{
			ImageRef: "some-registry.io/some-repo-source:source-id",
		}

		imageFactory := &image.Factory{
			SourceUploader: sourceUploader,
		}
		fakeImageWaiter := &fakes.FakeImageWaiter{}

		cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
			return imgcmds.NewSaveCommand(clientSetProvider, imageFactory, func(set k8s.ClientSet) imgcmds.ImageWaiter {
				return fakeImageWaiter
			})
		}

		when("a namespace is provided", func() {
			const namespace = "some-namespace"

			when("the image config is valid", func() {
				cacheSize := resource.MustParse("2G")
				expectedImage := &v1alpha1.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "kpack.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-image",
						Namespace: namespace,
						Annotations: map[string]string{
							"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Image","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"ClusterBuilder","name":"default"},"serviceAccount":"default","source":{"git":{"url":"some-git-url","revision":"some-git-rev"},"subPath":"some-sub-path"},"cacheSize":"2G","build":{"env":[{"name":"some-key","value":"some-val"}],"resources":{}}},"status":{}}`,
						},
					},
					Spec: v1alpha1.ImageSpec{
						Tag: "some-registry.io/some-repo",
						Builder: corev1.ObjectReference{
							Kind: v1alpha1.ClusterBuilderKind,
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
						CacheSize: &cacheSize,
					},
				}

				it("creates the image and wait on the image", func() {
					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
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
						ExpectedOutput: "\"some-image\" created\n",
						ExpectCreates: []runtime.Object{
							expectedImage,
						},
					}.TestKpack(t, cmdFunc)

					assert.Len(t, fakeImageWaiter.Calls, 1)
					assert.Equal(t, fakeImageWaiter.Calls[0], expectedImage)
				})

				it("defaults the git revision to master", func() {
					expectedImage.Spec.Source.Git.Revision = "master"
					expectedImage.ObjectMeta.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"Image","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"ClusterBuilder","name":"default"},"serviceAccount":"default","source":{"git":{"url":"some-git-url","revision":"master"},"subPath":"some-sub-path"},"cacheSize":"2G","build":{"env":[{"name":"some-key","value":"some-val"}],"resources":{}}},"status":{}}`

					testhelpers.CommandTest{
						Args: []string{
							"some-image",
							"--tag", "some-registry.io/some-repo",
							"--git", "some-git-url",
							"--sub-path", "some-sub-path",
							"--env", "some-key=some-val",
							"--cache-size", "2G",
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
							"--tag", "some-registry.io/some-repo",
							"--blob", "some-blob",
							"--git", "some-git-url",
							"-n", namespace,
						},
						ExpectErr:      true,
						ExpectedOutput: "Error: image source must be one of git, blob, or local-path\n",
					}.TestKpack(t, cmdFunc)

					assert.Len(t, fakeImageWaiter.Calls, 0)
				})
			})
		})

		when("a namespace is not provided", func() {
			when("the image config is valid", func() {
				it("creates the image", func() {
					expectedImage := &v1alpha1.Image{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Image",
							APIVersion: "kpack.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "some-image",
							Namespace: defaultNamespace,
							Annotations: map[string]string{
								"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Image","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"ClusterBuilder","name":"default"},"serviceAccount":"default","source":{"git":{"url":"some-git-url","revision":"some-git-rev"},"subPath":"some-sub-path"},"build":{"env":[{"name":"some-key","value":"some-val"}],"resources":{}}},"status":{}}`,
							},
						},
						Spec: v1alpha1.ImageSpec{
							Tag: "some-registry.io/some-repo",
							Builder: corev1.ObjectReference{
								Kind: v1alpha1.ClusterBuilderKind,
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
							"--tag", "some-registry.io/some-repo",
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
						ExpectErr:      true,
						ExpectedOutput: "Error: image source must be one of git, blob, or local-path\n",
					}.TestKpack(t, cmdFunc)

					assert.Len(t, fakeImageWaiter.Calls, 0)
				})
			})
		})

		when("the image uses local source code", func() {
			it("uploads the source image and creates the image config", func() {
				expectedImage := &v1alpha1.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "kpack.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-image",
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Image","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"ClusterBuilder","name":"default"},"serviceAccount":"default","source":{"registry":{"image":"some-registry.io/some-repo-source:source-id"},"subPath":"some-sub-path"},"build":{"env":[{"name":"some-key","value":"some-val"}],"resources":{}}},"status":{}}`,
						},
					},
					Spec: v1alpha1.ImageSpec{
						Tag: "some-registry.io/some-repo",
						Builder: corev1.ObjectReference{
							Kind: v1alpha1.ClusterBuilderKind,
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
						"--tag", "some-registry.io/some-repo",
						"--local-path", "some-local-path",
						"--sub-path", "some-sub-path",
						"--env", "some-key=some-val",
					},
					ExpectedOutput: "\"some-image\" created\n",
					ExpectCreates: []runtime.Object{
						expectedImage,
					},
				}.TestKpack(t, cmdFunc)

				assert.Len(t, fakeImageWaiter.Calls, 0)
			})
		})

		when("the image uses a non-default builder", func() {
			it("uploads the source image and creates the image config", func() {
				expectedImage := &v1alpha1.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "kpack.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-image",
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Image","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"Builder","namespace":"some-default-namespace","name":"some-builder"},"serviceAccount":"default","source":{"blob":{"url":"some-blob"}},"build":{"resources":{}}},"status":{}}`,
						},
					},
					Spec: v1alpha1.ImageSpec{
						Tag: "some-registry.io/some-repo",
						Builder: corev1.ObjectReference{
							Kind:      v1alpha1.BuilderKind,
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
						"--tag", "some-registry.io/some-repo",
						"--blob", "some-blob",
						"--builder", "some-builder",
					},
					ExpectedOutput: "\"some-image\" created\n",
					ExpectCreates: []runtime.Object{
						expectedImage,
					},
				}.TestKpack(t, cmdFunc)

				assert.Len(t, fakeImageWaiter.Calls, 0)
			})
		})

		when("the image uses a non-default cluster builder", func() {
			it("uploads the source image and creates the image config", func() {
				expectedImage := &v1alpha1.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "kpack.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-image",
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Image","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"ClusterBuilder","name":"some-builder"},"serviceAccount":"default","source":{"blob":{"url":"some-blob"}},"build":{"resources":{}}},"status":{}}`,
						},
					},
					Spec: v1alpha1.ImageSpec{
						Tag: "some-registry.io/some-repo",
						Builder: corev1.ObjectReference{
							Kind: v1alpha1.ClusterBuilderKind,
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
						"--tag", "some-registry.io/some-repo",
						"--blob", "some-blob",
						"--cluster-builder", "some-builder",
					},
					ExpectedOutput: "\"some-image\" created\n",
					ExpectCreates: []runtime.Object{
						expectedImage,
					},
				}.TestKpack(t, cmdFunc)

				assert.Len(t, fakeImageWaiter.Calls, 0)
			})
		})

		it("errors when tag is not provided", func() {
			testhelpers.CommandTest{
				Args: []string{
					"some-image",
				},
				ExpectErr:      true,
				ExpectedOutput: "Error: --tag is required to create the resource\n",
			}.TestKpack(t, cmdFunc)
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
						ExpectErr:      true,
						ExpectedOutput: "Error: image source must be one of git, blob, or local-path\n",
					}.TestKpack(t, cmdFunc)
					assert.Len(t, fakeImageWaiter.Calls, 0)
				})
			})

			when("the image config is valid", func() {
				expectedImage := &v1alpha1.Image{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Image",
						APIVersion: "kpack.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-image",
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"Image","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"ClusterBuilder","name":"default"},"serviceAccount":"default","source":{"git":{"url":"some-git-url","revision":"some-git-rev"},"subPath":"some-sub-path"},"build":{"env":[{"name":"some-key","value":"some-val"}],"resources":{}}},"status":{}}`,
						},
					},
					Spec: v1alpha1.ImageSpec{
						Tag: "some-registry.io/some-repo",
						Builder: corev1.ObjectReference{
							Kind: v1alpha1.ClusterBuilderKind,
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

				it("can output in yaml format and does not wait", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: Image
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"Image","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"ClusterBuilder","name":"default"},"serviceAccount":"default","source":{"git":{"url":"some-git-url","revision":"some-git-rev"},"subPath":"some-sub-path"},"build":{"env":[{"name":"some-key","value":"some-val"}],"resources":{}}},"status":{}}'
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
  serviceAccount: default
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
						ExpectCreates: []runtime.Object{
							expectedImage,
						},
					}.TestKpack(t, cmdFunc)
					assert.Len(t, fakeImageWaiter.Calls, 0)
				})

				it("can output in json format and does not wait", func() {
					const resourceJSON = `{
    "kind": "Image",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "some-image",
        "namespace": "some-default-namespace",
        "creationTimestamp": null,
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"Image\",\"apiVersion\":\"kpack.io/v1alpha1\",\"metadata\":{\"name\":\"some-image\",\"namespace\":\"some-default-namespace\",\"creationTimestamp\":null},\"spec\":{\"tag\":\"some-registry.io/some-repo\",\"builder\":{\"kind\":\"ClusterBuilder\",\"name\":\"default\"},\"serviceAccount\":\"default\",\"source\":{\"git\":{\"url\":\"some-git-url\",\"revision\":\"some-git-rev\"},\"subPath\":\"some-sub-path\"},\"build\":{\"env\":[{\"name\":\"some-key\",\"value\":\"some-val\"}],\"resources\":{}}},\"status\":{}}"
        }
    },
    "spec": {
        "tag": "some-registry.io/some-repo",
        "builder": {
            "kind": "ClusterBuilder",
            "name": "default"
        },
        "serviceAccount": "default",
        "source": {
            "git": {
                "url": "some-git-url",
                "revision": "some-git-rev"
            },
            "subPath": "some-sub-path"
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
						ExpectErr:      true,
						ExpectedOutput: "Error: image source must be one of git, blob, or local-path\n",
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
						ExpectedOutput: `"some-image" created (dry run)
`,
					}.TestKpack(t, cmdFunc)
					assert.Len(t, fakeImageWaiter.Calls, 0)
				})

				when("output flag is used", func() {
					it("does not create an image and prints the resource output", func() {
						const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: Image
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"kind":"Image","apiVersion":"kpack.io/v1alpha1","metadata":{"name":"some-image","namespace":"some-default-namespace","creationTimestamp":null},"spec":{"tag":"some-registry.io/some-repo","builder":{"kind":"ClusterBuilder","name":"default"},"serviceAccount":"default","source":{"git":{"url":"some-git-url","revision":"some-git-rev"},"subPath":"some-sub-path"},"build":{"env":[{"name":"some-key","value":"some-val"}],"resources":{}}},"status":{}}'
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
  serviceAccount: default
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
						}.TestKpack(t, cmdFunc)
						assert.Len(t, fakeImageWaiter.Calls, 0)
					})
				})
			})
		})
	})

	when("patching", func() {
		const defaultNamespace = "some-default-namespace"

		sourceUploader := &srcfakes.SourceUploader{
			ImageRef: "",
		}

		patchFactory := &image.Factory{
			SourceUploader: sourceUploader,
		}
		fakeImageWaiter := &fakes.FakeImageWaiter{}

		cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
			clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
			return imgcmds.NewPatchCommand(clientSetProvider, patchFactory, func(set k8s.ClientSet) imgcmds.ImageWaiter {
				return fakeImageWaiter
			})
		}

		img := &v1alpha1.Image{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-image",
				Namespace: defaultNamespace,
			},
			Spec: v1alpha1.ImageSpec{
				Tag: "some-tag",
				Builder: corev1.ObjectReference{
					Kind: v1alpha1.ClusterBuilderKind,
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
				assert.Len(t, fakeImageWaiter.Calls, 0)
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
							`{"spec":{"source":{"subPath":null}}}`,
						},
					}.TestKpack(t, cmdFunc)
					assert.Len(t, fakeImageWaiter.Calls, 0)
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
							`{"spec":{"source":{"subPath":"a-new-path"}}}`,
						},
					}.TestKpack(t, cmdFunc)
					assert.Len(t, fakeImageWaiter.Calls, 0)
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
						`{"spec":{"source":{"blob":{"url":"some-blob"},"git":null}}}`,
					},
				}.TestKpack(t, cmdFunc)
				assert.Len(t, fakeImageWaiter.Calls, 0)
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
						`{"spec":{"source":{"git":{"revision":"some-new-revision"}}}}`,
					},
				}.TestKpack(t, cmdFunc)
				assert.Len(t, fakeImageWaiter.Calls, 0)
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
						`{"spec":{"source":{"blob":null,"git":{"revision":"master","url":"some-new-git-url"}}}}`,
					},
				}.TestKpack(t, cmdFunc)
				assert.Len(t, fakeImageWaiter.Calls, 0)
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
						`{"spec":{"builder":{"kind":"Builder","name":"some-builder","namespace":"some-default-namespace"}}}`,
					},
				}.TestKpack(t, cmdFunc)
				assert.Len(t, fakeImageWaiter.Calls, 0)
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
						`{"spec":{"build":{"env":[{"name":"key1","value":"value1"}]}}}`,
					},
				}.TestKpack(t, cmdFunc)
				assert.Len(t, fakeImageWaiter.Calls, 0)
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
						`{"spec":{"build":{"env":[{"name":"key1","value":"some-other-value"},{"name":"key2","value":"value2"}]}}}`,
					},
				}.TestKpack(t, cmdFunc)
				assert.Len(t, fakeImageWaiter.Calls, 0)
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
						`{"spec":{"build":{"env":[{"name":"key1","value":"value1"},{"name":"key2","value":"value2"},{"name":"key3","value":"value3"}]}}}`,
					},
				}.TestKpack(t, cmdFunc)

				assert.Len(t, fakeImageWaiter.Calls, 0)
			})
		})

		it("can patch cache size", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					img,
				},
				Args: []string{
					"some-image",
					"--cache-size", "3G",
				},
				ExpectedOutput: "\"some-image\" patched\n",
				ExpectPatches: []string{
					`{"spec":{"cacheSize":"3G"}}`,
				},
			}.TestKpack(t, cmdFunc)
			assert.Len(t, fakeImageWaiter.Calls, 0)
		})

		it("will wait on the image update if requested", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					img,
				},
				Args: []string{
					"some-image",
					"--git-revision", "some-new-revision",
					"--wait",
				},
				ExpectedOutput: "\"some-image\" patched\n",
				ExpectPatches: []string{
					`{"spec":{"source":{"git":{"revision":"some-new-revision"}}}}`,
				},
			}.TestKpack(t, cmdFunc)

			expectedWaitImage := img.DeepCopy()
			expectedWaitImage.Spec.Source.Git.Revision = "some-new-revision"

			assert.Len(t, fakeImageWaiter.Calls, 1)
			assert.Equal(t, fakeImageWaiter.Calls[0], expectedWaitImage)
		})

		when("output flag is used", func() {
			when("no parameters are provided", func() {
				it("does not give output and informs user of nothing to patch", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							img,
						},
						Args: []string{
							"some-image",
							"--output", "yaml",
						},
						ExpectedErrorOutput: "nothing to patch\n",
					}.TestKpack(t, cmdFunc)
					assert.Len(t, fakeImageWaiter.Calls, 0)
				})
			})

			it("can output resources in yaml and does not wait", func() {
				const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: Image
metadata:
  creationTimestamp: null
  name: some-image
  namespace: some-default-namespace
spec:
  build:
    env:
    - name: key1
      value: value1
    - name: key2
      value: value2
    resources: {}
  builder:
    kind: ClusterBuilder
    name: some-ccb
  source:
    blob:
      url: some-blob
    subPath: some-path
  tag: some-tag
status: {}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						img,
					},
					Args: []string{
						"some-image",
						"--blob", "some-blob",
						"--output", "yaml",
						"--wait",
					},
					ExpectedOutput: resourceYAML,
					ExpectPatches: []string{
						`{"spec":{"source":{"blob":{"url":"some-blob"},"git":null}}}`,
					},
				}.TestKpack(t, cmdFunc)
				assert.Len(t, fakeImageWaiter.Calls, 0)
			})

			it("can output resources in json and does not wait", func() {
				const resourceJSON = `{
    "kind": "Image",
    "apiVersion": "kpack.io/v1alpha1",
    "metadata": {
        "name": "some-image",
        "namespace": "some-default-namespace",
        "creationTimestamp": null
    },
    "spec": {
        "tag": "some-tag",
        "builder": {
            "kind": "ClusterBuilder",
            "name": "some-ccb"
        },
        "source": {
            "blob": {
                "url": "some-blob"
            },
            "subPath": "some-path"
        },
        "build": {
            "env": [
                {
                    "name": "key1",
                    "value": "value1"
                },
                {
                    "name": "key2",
                    "value": "value2"
                }
            ],
            "resources": {}
        }
    },
    "status": {}
}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						img,
					},
					Args: []string{
						"some-image",
						"--blob", "some-blob",
						"--output", "json",
						"--wait",
					},
					ExpectedOutput: resourceJSON,
					ExpectPatches: []string{
						`{"spec":{"source":{"blob":{"url":"some-blob"},"git":null}}}`,
					},
				}.TestKpack(t, cmdFunc)
				assert.Len(t, fakeImageWaiter.Calls, 0)
			})
		})

		when("dry-run flag is used", func() {
			when("no parameters are provided", func() {
				it("informs user of nothing to patch", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							img,
						},
						Args: []string{
							"some-image",
							"--dry-run",
						},
						ExpectedOutput: "nothing to patch\n",
					}.TestKpack(t, cmdFunc)
				})
			})

			it("does not patch and prints result message with dry run indicated", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						img,
					},
					Args: []string{
						"some-image",
						"--blob", "some-blob",
						"--dry-run",
						"--wait",
					},
					ExpectedOutput: `"some-image" patched (dry run)
`,
				}.TestKpack(t, cmdFunc)
				assert.Len(t, fakeImageWaiter.Calls, 0)
			})

			when("output flag is used", func() {
				it("does not patch and prints the resource output", func() {
					const resourceYAML = `apiVersion: kpack.io/v1alpha1
kind: Image
metadata:
  creationTimestamp: null
  name: some-image
  namespace: some-default-namespace
spec:
  build:
    env:
    - name: key1
      value: value1
    - name: key2
      value: value2
    resources: {}
  builder:
    kind: ClusterBuilder
    name: some-ccb
  source:
    blob:
      url: some-blob
    subPath: some-path
  tag: some-tag
status: {}
`

					testhelpers.CommandTest{
						Objects: []runtime.Object{
							img,
						},
						Args: []string{
							"some-image",
							"--blob", "some-blob",
							"--output", "yaml",
							"--dry-run",
							"--wait",
						},
						ExpectedOutput: resourceYAML,
					}.TestKpack(t, cmdFunc)
					assert.Len(t, fakeImageWaiter.Calls, 0)
				})
			})
		})
	})
}

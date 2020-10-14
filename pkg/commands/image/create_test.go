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
	srcfakes "github.com/pivotal/build-service-cli/pkg/registry/fakes"
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
	fakeImageWaiter := &fakes.FakeImageWaiter{}

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return imgcmds.NewCreateCommand(clientSetProvider, imageFactory, func(set k8s.ClientSet) imgcmds.ImageWaiter {
			return fakeImageWaiter
		})
	}

	when("a namespace is provided", func() {
		const namespace = "some-namespace"
		var cacheSize = resource.MustParse("2G")

		when("the image config is valid", func() {
			it("creates the image and wait on the image", func() {
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
					ExpectedOutput: `Image "some-image" created
`,
					ExpectCreates: []runtime.Object{
						expectedImage,
					},
				}.TestKpack(t, cmdFunc)

				assert.Len(t, fakeImageWaiter.Calls, 1)
				assert.Equal(t, fakeImageWaiter.Calls[0], expectedImage)
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
					ExpectedOutput: `Image "some-image" created
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
				ExpectedOutput: `Uploading to 'some-registry.io/some-repo-source'...
Image "some-image" created
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
				ExpectedOutput: `Image "some-image" created
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
				ExpectedOutput: `Image "some-image" created
`,
				ExpectCreates: []runtime.Object{
					expectedImage,
				},
			}.TestKpack(t, cmdFunc)

			assert.Len(t, fakeImageWaiter.Calls, 0)
		})
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
					ExpectedOutput: `Image "some-image" created (dry run)
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
}

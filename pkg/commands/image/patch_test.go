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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	cmdFakes "github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	imgcmds "github.com/vmware-tanzu/kpack-cli/pkg/commands/image"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	registryfakes "github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestImagePatchCommand(t *testing.T) {
	spec.Run(t, "TestImagePatchCommand", testImagePatchCommand)
}

func testImagePatchCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	fakeSourceUploader := registryfakes.NewSourceUploader("some-registry.io/some-repo-source:source-id")
	registryUtilProvider := registryfakes.UtilProvider{
		FakeSourceUploader: fakeSourceUploader,
	}

	fakeImageWaiter := &cmdFakes.FakeImageWaiter{}

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return imgcmds.NewPatchCommand(clientSetProvider, registryUtilProvider, func(set k8s.ClientSet) imgcmds.ImageWaiter {
			return fakeImageWaiter
		})
	}

	existingImage := &v1alpha1.Image{
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
		it("informs user of no change in patch and does not wait", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					existingImage,
				},
				Args: []string{
					"some-image",
				},
				ExpectedOutput: `Patching Image...
Image "some-image" patched (no change)
`,
			}.TestKpack(t, cmdFunc)
			assert.Len(t, fakeImageWaiter.Calls, 0)
		})
	})

	when("patching source", func() {
		when("patching the sub path", func() {
			it("can patch it with an empty string", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						existingImage,
					},
					Args: []string{
						"some-image",
						"--sub-path", "",
					},
					ExpectedOutput: `Patching Image...
Image "some-image" patched
`,
					ExpectPatches: []string{
						`{"spec":{"source":{"subPath":null}}}`,
					},
				}.TestKpack(t, cmdFunc)
				assert.Len(t, fakeImageWaiter.Calls, 0)
			})

			it("can patch it with a non-empty string", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						existingImage,
					},
					Args: []string{
						"some-image",
						"--sub-path", "a-new-path",
					},
					ExpectedOutput: `Patching Image...
Image "some-image" patched
`,
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
					existingImage,
				},
				Args: []string{
					"some-image",
					"--blob", "some-blob",
				},
				ExpectedOutput: `Patching Image...
Image "some-image" patched
`,
				ExpectPatches: []string{
					`{"spec":{"source":{"blob":{"url":"some-blob"},"git":null}}}`,
				},
			}.TestKpack(t, cmdFunc)
			assert.Len(t, fakeImageWaiter.Calls, 0)
		})

		it("can change git revision if existing source is git", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					existingImage,
				},
				Args: []string{
					"some-image",
					"--git-revision", "some-new-revision",
				},
				ExpectedOutput: `Patching Image...
Image "some-image" patched
`,
				ExpectPatches: []string{
					`{"spec":{"source":{"git":{"revision":"some-new-revision"}}}}`,
				},
			}.TestKpack(t, cmdFunc)
			assert.Len(t, fakeImageWaiter.Calls, 0)
		})

		it("git revision defaults to main if not provided with git", func() {
			existingImage.Spec.Source = v1alpha1.SourceConfig{
				Blob: &v1alpha1.Blob{
					URL: "some-blob",
				},
			}

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					existingImage,
				},
				Args: []string{
					"some-image",
					"--git", "some-new-git-url",
				},
				ExpectedOutput: `Patching Image...
Image "some-image" patched
`,
				ExpectPatches: []string{
					`{"spec":{"source":{"blob":null,"git":{"revision":"main","url":"some-new-git-url"}}}}`,
				},
			}.TestKpack(t, cmdFunc)
			assert.Len(t, fakeImageWaiter.Calls, 0)
		})

	})

	when("patching the builder", func() {
		it("can patch the builder", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					existingImage,
				},
				Args: []string{
					"some-image",
					"--builder", "some-builder",
				},
				ExpectedOutput: `Patching Image...
Image "some-image" patched
`,
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
					existingImage,
				},
				Args: []string{
					"some-image",
					"-d", "key2",
				},
				ExpectedOutput: `Patching Image...
Image "some-image" patched
`,
				ExpectPatches: []string{
					`{"spec":{"build":{"env":[{"name":"key1","value":"value1"}]}}}`,
				},
			}.TestKpack(t, cmdFunc)
			assert.Len(t, fakeImageWaiter.Calls, 0)
		})

		it("can update existing env vars", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					existingImage,
				},
				Args: []string{
					"some-image",
					"-e", "key1=some-other-value",
				},
				ExpectedOutput: `Patching Image...
Image "some-image" patched
`,
				ExpectPatches: []string{
					`{"spec":{"build":{"env":[{"name":"key1","value":"some-other-value"},{"name":"key2","value":"value2"}]}}}`,
				},
			}.TestKpack(t, cmdFunc)
			assert.Len(t, fakeImageWaiter.Calls, 0)
		})

		it("can add new env vars", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					existingImage,
				},
				Args: []string{
					"some-image",
					"-e", "key3=value3",
				},
				ExpectedOutput: `Patching Image...
Image "some-image" patched
`,
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
				existingImage,
			},
			Args: []string{
				"some-image",
				"--cache-size", "3G",
			},
			ExpectedOutput: `Patching Image...
Image "some-image" patched
`,
			ExpectPatches: []string{
				`{"spec":{"cacheSize":"3G"}}`,
			},
		}.TestKpack(t, cmdFunc)
		assert.Len(t, fakeImageWaiter.Calls, 0)
	})

	it("will wait on the image update if requested", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				existingImage,
			},
			Args: []string{
				"some-image",
				"--git-revision", "some-new-revision",
				"--registry-ca-cert-path", "some-cert-path",
				"--registry-verify-certs",
				"--wait",
			},
			ExpectedOutput: `Patching Image...
Image "some-image" patched
`,
			ExpectPatches: []string{
				`{"spec":{"source":{"git":{"revision":"some-new-revision"}}}}`,
			},
		}.TestKpack(t, cmdFunc)

		expectedWaitImage := existingImage.DeepCopy()
		expectedWaitImage.Spec.Source.Git.Revision = "some-new-revision"

		assert.Len(t, fakeImageWaiter.Calls, 1)
		assert.Equal(t, fakeImageWaiter.Calls[0], expectedWaitImage)
	})

	when("output flag is used", func() {
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
					existingImage,
				},
				Args: []string{
					"some-image",
					"--blob", "some-blob",
					"--output", "yaml",
					"--wait",
				},
				ExpectedOutput: resourceYAML,
				ExpectedErrorOutput: `Patching Image...
`,
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
					existingImage,
				},
				Args: []string{
					"some-image",
					"--blob", "some-blob",
					"--output", "json",
					"--wait",
				},
				ExpectedOutput: resourceJSON,
				ExpectedErrorOutput: `Patching Image...
`,
				ExpectPatches: []string{
					`{"spec":{"source":{"blob":{"url":"some-blob"},"git":null}}}`,
				},
			}.TestKpack(t, cmdFunc)
			assert.Len(t, fakeImageWaiter.Calls, 0)
		})

		when("there are no changes in the patch", func() {
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
    git:
      revision: some-revision
      url: some-git-url
    subPath: some-path
  tag: some-tag
status: {}
`

			it("can output unpatched resource in requested format and does not wait", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						existingImage,
					},
					Args: []string{
						"some-image",
						"--output", "yaml",
					},
					ExpectedErrorOutput: `Patching Image...
`,
					ExpectedOutput: resourceYAML,
				}.TestKpack(t, cmdFunc)
				assert.Len(t, fakeImageWaiter.Calls, 0)
			})
		})
	})

	when("dry-run flag is used", func() {
		fakeSourceUploader.SetSkipUpload(true)

		it("does not patch and prints result message with dry run indicated", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					existingImage,
				},
				Args: []string{
					"some-image",
					"--local-path", "some-local-path",
					"--sub-path", "some-sub-path",
					"--env", "some-key=some-val",
					"--dry-run",
					"--wait",
				},
				ExpectedOutput: `Patching Image... (dry run)
	Skipping 'some-registry.io/some-repo-source:source-id'
Image "some-image" patched (dry run)
`,
			}.TestKpack(t, cmdFunc)
			assert.Len(t, fakeImageWaiter.Calls, 0)
		})

		when("there are no changes in the patch", func() {
			it("does not patch and informs of no change", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						existingImage,
					},
					Args: []string{
						"some-image",
						"--dry-run",
					},
					ExpectedOutput: `Patching Image... (dry run)
Image "some-image" patched (dry run)
`,
				}.TestKpack(t, cmdFunc)
			})
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
    registry:
      image: some-registry.io/some-repo-source:source-id
    subPath: some-sub-path
  tag: some-tag
status: {}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						existingImage,
					},
					Args: []string{
						"some-image",
						"--local-path", "some-local-path",
						"--sub-path", "some-sub-path",
						"--dry-run",
						"--output", "yaml",
						"--wait",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Patching Image... (dry run)
	Skipping 'some-registry.io/some-repo-source:source-id'
`,
				}.TestKpack(t, cmdFunc)
				assert.Len(t, fakeImageWaiter.Calls, 0)
			})
		})
	})

	when("dry-run-with-image-upload flag is used", func() {
		it("does not patch and prints result message with dry run indicated", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					existingImage,
				},
				Args: []string{
					"some-image",
					"--local-path", "some-local-path",
					"--sub-path", "some-sub-path",
					"--dry-run-with-image-upload",
					"--wait",
				},
				ExpectedOutput: `Patching Image... (dry run with image upload)
	Uploading 'some-registry.io/some-repo-source:source-id'
Image "some-image" patched (dry run with image upload)
`,
			}.TestKpack(t, cmdFunc)
			assert.Len(t, fakeImageWaiter.Calls, 0)
		})

		when("there are no changes in the patch", func() {
			it("does not patch and informs of no change", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						existingImage,
					},
					Args: []string{
						"some-image",
						"--dry-run-with-image-upload",
					},
					ExpectedOutput: `Patching Image... (dry run with image upload)
Image "some-image" patched (dry run with image upload)
`,
				}.TestKpack(t, cmdFunc)
			})
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
    registry:
      image: some-registry.io/some-repo-source:source-id
    subPath: some-sub-path
  tag: some-tag
status: {}
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						existingImage,
					},
					Args: []string{
						"some-image",
						"--local-path", "some-local-path",
						"--sub-path", "some-sub-path",
						"--dry-run-with-image-upload",
						"--output", "yaml",
						"--wait",
					},
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Patching Image... (dry run with image upload)
	Uploading 'some-registry.io/some-repo-source:source-id'
`,
				}.TestKpack(t, cmdFunc)
				assert.Len(t, fakeImageWaiter.Calls, 0)
			})
		})
	})
}

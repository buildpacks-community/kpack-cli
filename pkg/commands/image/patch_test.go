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

	imgcmds "github.com/pivotal/build-service-cli/pkg/commands/image"
	"github.com/pivotal/build-service-cli/pkg/image"
	"github.com/pivotal/build-service-cli/pkg/image/fakes"
	"github.com/pivotal/build-service-cli/pkg/k8s"
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

	it("will wait on the image update if requested", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				img,
			},
			Args: []string{
				"some-image",
				"--git-revision", "some-new-revision",
				"--registry-ca-cert-path", "some-cert-path",
				"--registry-verify-certs",
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
			const resourceYAML = `metadata:
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
				const resourceYAML = `metadata:
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
}

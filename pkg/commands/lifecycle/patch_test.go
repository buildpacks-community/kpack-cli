// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package lifecycle_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/buildpacks-community/kpack-cli/pkg/commands/lifecycle"
	registryfakes "github.com/buildpacks-community/kpack-cli/pkg/registry/fakes"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUpdateCommand(t *testing.T) {
	spec.Run(t, "TestUpdateCommand", testUpdateCommand)
}

func testUpdateCommand(t *testing.T, when spec.G, it spec.S) {
	const deprecationWarning = `Command "patch" is deprecated, This command will be removed in a future release.
Please use 'kp clusterlifecycle' commands instead.

`
	fakeRegistryUtilProvider := &registryfakes.UtilProvider{
		FakeFetcher: registryfakes.NewLifecycleImageFetcher(
			registryfakes.LifecycleInfo{
				Metadata: "value-not-validated-by-cli",
				ImageInfo: registryfakes.ImageInfo{
					Ref:    "some-registry.io/repo/lifecycle-image",
					Digest: "lifecycle-image-digest",
				},
			},
		),
	}

	cmdFunc := func(k8sClient *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeK8sProvider(k8sClient, "")
		return lifecycle.NewUpdateCommand(clientSetProvider, fakeRegistryUtilProvider)
	}

	kpConfig := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kp-config",
			Namespace: "kpack",
		},
		Data: map[string]string{
			"default.repository": "default-registry.io/default-repo",
		},
	}

	lifecycleImageConfig := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "lifecycle-image",
			Namespace: "kpack",
		},
		Data: map[string]string{},
	}

	updatedLifecycleImageConfig := lifecycleImageConfig.DeepCopy()
	updatedLifecycleImageConfig.Data["image"] = "default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest"

	it("errors when lifecycle-image configmap is not found", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				kpConfig,
			},
			Args: []string{
				"--image", "some-registry.io/repo/lifecycle-image",
			},
			ExpectErr:           true,
			ExpectedOutput:      deprecationWarning + "Patching lifecycle config...\n",
			ExpectedErrorOutput: "Error: configmap \"lifecycle-image\" not found in \"kpack\" namespace\n",
		}.TestK8s(t, cmdFunc)
	})

	it("errors when io.buildpacks.lifecycle.metadata label is not set on given image", func() {
		fetcher := &registryfakes.Fetcher{}
		fetcher.AddImage("some-registry.io/repo/image-without-metadata", registryfakes.NewFakeImage("some-digest"))
		fakeRegistryUtilProvider.FakeFetcher = fetcher

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				kpConfig,
				lifecycleImageConfig,
			},
			Args: []string{
				"--image", "some-registry.io/repo/image-without-metadata",
			},
			ExpectErr:           true,
			ExpectedOutput:      deprecationWarning + "Patching lifecycle config...\n",
			ExpectedErrorOutput: "Error: image missing lifecycle metadata\n",
		}.TestK8s(t, cmdFunc)
	})

	it("errors when default.repository key is not found in kp-config configmap", func() {
		badKpConfig := &corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{},
		}

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				badKpConfig,
				lifecycleImageConfig,
			},
			Args: []string{
				"--image", "some-registry.io/repo/lifecycle-image",
			},
			ExpectErr:           true,
			ExpectedOutput:      deprecationWarning + "Patching lifecycle config...\n",
			ExpectedErrorOutput: "Error: failed to get default repository: use \"kp config default-repository\" to set\n",
		}.TestK8s(t, cmdFunc)
	})

	it("updates lifecycle-image ConfigMap", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				kpConfig,
				lifecycleImageConfig,
			},
			Args: []string{
				"--image", "some-registry.io/repo/lifecycle-image",
			},
			ExpectPatches: []string{
				`{"data":{"image":"default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest"}}`,
			},
			ExpectedOutput: deprecationWarning + `Patching lifecycle config...
	Uploading 'default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest'
Patched lifecycle config
`,
		}.TestK8s(t, cmdFunc)
	})

	when("output flag is used", func() {
		it("can output in yaml format", func() {
			const resourceYAML = `apiVersion: v1
data:
  image: default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: lifecycle-image
  namespace: kpack
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					kpConfig,
					lifecycleImageConfig,
				},
				Args: []string{
					"--image", "some-registry.io/repo/lifecycle-image",
					"--output", "yaml",
				},
				ExpectPatches: []string{
					`{"data":{"image":"default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest"}}`,
				},
				ExpectedOutput: deprecationWarning + resourceYAML,
				ExpectedErrorOutput: `Patching lifecycle config...
	Uploading 'default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest'
`,
			}.TestK8s(t, cmdFunc)
		})

		it("can output in json format", func() {
			const resourceJSON = `{
    "kind": "ConfigMap",
    "apiVersion": "v1",
    "metadata": {
        "name": "lifecycle-image",
        "namespace": "kpack",
        "creationTimestamp": null
    },
    "data": {
        "image": "default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest"
    }
}
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					kpConfig,
					lifecycleImageConfig,
				},
				Args: []string{
					"--image", "some-registry.io/repo/lifecycle-image",
					"--output", "json",
				},
				ExpectPatches: []string{
					`{"data":{"image":"default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest"}}`,
				},
				ExpectedOutput: deprecationWarning + resourceJSON,
				ExpectedErrorOutput: `Patching lifecycle config...
	Uploading 'default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest'
`,
			}.TestK8s(t, cmdFunc)
		})
	})

	when("dry-run flag is used", func() {
		it("does not update lifecycle-image configmap and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					kpConfig,
					lifecycleImageConfig,
				},
				Args: []string{
					"--image", "some-registry.io/repo/lifecycle-image",
					"--dry-run",
				},
				ExpectedOutput: deprecationWarning + `Patching lifecycle config... (dry run)
	Skipping 'default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest'
Patched lifecycle config (dry run)
`,
			}.TestK8s(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("does not update lifecycle-image configmap and prints the resource output", func() {
				const resourceYAML = `apiVersion: v1
data:
  image: default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: lifecycle-image
  namespace: kpack
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						kpConfig,
						lifecycleImageConfig,
					},
					Args: []string{
						"--image", "some-registry.io/repo/lifecycle-image",
						"--dry-run",
						"--output", "yaml",
					},
					ExpectedOutput: deprecationWarning + resourceYAML,
					ExpectedErrorOutput: `Patching lifecycle config... (dry run)
	Skipping 'default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest'
`,
				}.TestK8s(t, cmdFunc)
			})
		})
	})

	when("dry-run-with-image-upload flag is used", func() {
		it("does not update lifecycle-image configmap and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					kpConfig,
					lifecycleImageConfig,
				},
				Args: []string{
					"--image", "some-registry.io/repo/lifecycle-image",
					"--dry-run-with-image-upload",
				},
				ExpectedOutput: deprecationWarning + `Patching lifecycle config... (dry run with image upload)
	Uploading 'default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest'
Patched lifecycle config (dry run with image upload)
`,
			}.TestK8s(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("does not update lifecycle-image configmap and prints the resource output", func() {
				const resourceYAML = `apiVersion: v1
data:
  image: default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: lifecycle-image
  namespace: kpack
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						kpConfig,
						lifecycleImageConfig,
					},
					Args: []string{
						"--image", "some-registry.io/repo/lifecycle-image",
						"--dry-run-with-image-upload",
						"--output", "yaml",
					},
					ExpectedOutput: deprecationWarning + resourceYAML,
					ExpectedErrorOutput: `Patching lifecycle config... (dry run with image upload)
	Uploading 'default-registry.io/default-repo/lifecycle@sha256:lifecycle-image-digest'
`,
				}.TestK8s(t, cmdFunc)
			})
		})
	})
}

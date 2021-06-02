// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package lifecycle_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/lifecycle"
	registryfakes "github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestUpdateCommand(t *testing.T) {
	spec.Run(t, "TestUpdateCommand", testUpdateCommand)
}

func testUpdateCommand(t *testing.T, when spec.G, it spec.S) {

	fakeRelocator := &registryfakes.Relocator{}
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
		FakeRelocator: fakeRelocator,
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
			"canonical.repository": "canonical-registry.io/canonical-repo",
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
	updatedLifecycleImageConfig.Data["image"] = "canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest"

	it("errors when lifecycle-image configmap is not found", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				kpConfig,
			},
			Args: []string{
				"--image", "some-registry.io/repo/lifecycle-image",
			},
			ExpectErr: true,
			ExpectedOutput: `Updating lifecycle image...
Error: configmap "lifecycle-image" not found in "kpack" namespace
`,
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
			ExpectErr: true,
			ExpectedOutput: `Updating lifecycle image...
Error: image missing lifecycle metadata
`,
		}.TestK8s(t, cmdFunc)
	})

	it("errors when kp-config configmap is not found", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				lifecycleImageConfig,
			},
			Args: []string{
				"--image", "some-registry.io/repo/lifecycle-image",
			},
			ExpectErr: true,
			ExpectedOutput: `Updating lifecycle image...
Error: failed to get canonical repository: configmaps "kp-config" not found
`,
		}.TestK8s(t, cmdFunc)
	})

	it("errors when canonical.repository key is not found in kp-config configmap", func() {
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
			ExpectErr: true,
			ExpectedOutput: `Updating lifecycle image...
Error: failed to get canonical repository: key "canonical.repository" not found in configmap "kp-config"
`,
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
			ExpectUpdates: []clientgotesting.UpdateActionImpl{
				{
					Object: updatedLifecycleImageConfig,
				},
			},
			ExpectedOutput: `Updating lifecycle image...
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Updated lifecycle image
`,
		}.TestK8s(t, cmdFunc)
	})

	when("output flag is used", func() {
		it("can output in yaml format", func() {
			const resourceYAML = `apiVersion: v1
data:
  image: canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest
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
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: updatedLifecycleImageConfig,
					},
				},
				ExpectedOutput: resourceYAML,
				ExpectedErrorOutput: `Updating lifecycle image...
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
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
        "image": "canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest"
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
				ExpectUpdates: []clientgotesting.UpdateActionImpl{
					{
						Object: updatedLifecycleImageConfig,
					},
				},
				ExpectedOutput: resourceJSON,
				ExpectedErrorOutput: `Updating lifecycle image...
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
`,
			}.TestK8s(t, cmdFunc)
		})
	})

	when("dry-run flag is used", func() {
		fakeRelocator.SetSkip(true)

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
				ExpectedOutput: `Updating lifecycle image... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Updated lifecycle image (dry run)
`,
			}.TestK8s(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("does not update lifecycle-image configmap and prints the resource output", func() {
				const resourceYAML = `apiVersion: v1
data:
  image: canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest
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
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Updating lifecycle image... (dry run)
	Skipping 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
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
				ExpectedOutput: `Updating lifecycle image... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
Updated lifecycle image (dry run with image upload)
`,
			}.TestK8s(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("does not update lifecycle-image configmap and prints the resource output", func() {
				const resourceYAML = `apiVersion: v1
data:
  image: canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest
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
					ExpectedOutput: resourceYAML,
					ExpectedErrorOutput: `Updating lifecycle image... (dry run with image upload)
	Uploading 'canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest'
`,
				}.TestK8s(t, cmdFunc)
			})
		})
	})
}

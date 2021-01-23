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

	"github.com/pivotal/build-service-cli/pkg/commands/lifecycle"
	registryfakes "github.com/pivotal/build-service-cli/pkg/registry/fakes"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestUpdateCommand(t *testing.T) {
	spec.Run(t, "TestUpdateCommand", testUpdateCommand)
}

func testUpdateCommand(t *testing.T, when spec.G, it spec.S) {

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
		FakeRelocator: &registryfakes.Relocator{},
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

	it("updates lifecycle-image ConfigMap", func() {
		updatedLifecycleImageConfig := lifecycleImageConfig.DeepCopy()
		updatedLifecycleImageConfig.Data["image"] = "canonical-registry.io/canonical-repo/lifecycle@sha256:lifecycle-image-digest"

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
			ExpectedOutput: `Updated lifecycle image
`,
		}.TestK8s(t, cmdFunc)
	})

	it("errors when io.buildpacks.lifecycle.metadata label is not set on given image", func() {
		fetcher := &registryfakes.Fetcher{}
		fetcher.AddImage("some-registry.io/repo/image-without-metadata", registryfakes.NewFakeImage("some-digest"))
		fakeRegistryUtilProvider.FakeFetcher = fetcher

		testhelpers.CommandTest{
			Args: []string{
				"--image", "some-registry.io/repo/image-without-metadata",
			},
			ExpectErr: true,
			ExpectedOutput: `Error: image missing lifecycle metadata
`,
		}.TestK8s(t, cmdFunc)
	})

	it("errors when kp-config configmap is not found", func() {
		testhelpers.CommandTest{
			Args: []string{
				"--image", "some-registry.io/repo/lifecycle-image",
			},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical repository: configmaps "kp-config" not found
`,
		}.TestK8s(t, cmdFunc)
	})

	it("errors when canonical.repository key is not found in kp-config configmap", func() {
		badConfig := &corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{},
		}

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				badConfig,
			},
			Args: []string{
				"--image", "some-registry.io/repo/lifecycle-image",
			},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical repository: key "canonical.repository" not found in configmap "kp-config"
`,
		}.TestK8s(t, cmdFunc)
	})

	it("errors when lifecycle-image configmap is not found", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				kpConfig,
			},
			Args: []string{
				"--image", "some-registry.io/repo/lifecycle-image",
			},
			ExpectErr: true,
			ExpectedOutput: `Error: configmap "lifecycle-image" not found in "kpack" namespace
`,
		}.TestK8s(t, cmdFunc)
	})
}

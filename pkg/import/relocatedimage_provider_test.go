package _import

import (
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/kpack/pkg/registry/registryfakes"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
)

func TestRelocatedImageProvider(t *testing.T) {
	spec.Run(t, "TestRelocatedImageProvider", testRelocatedImageProvider)
}

func testRelocatedImageProvider(t *testing.T, when spec.G, it spec.S) {
	when("calculating relocated image location", func() {
		it("fetches the prerelocated image digest", func() {
			fetcher := &fakeFetcher{Images: map[string]v1.Image{
				"some-registry.com/some-repo/image@sha256:some-digest": fakes.NewFakeImage("some-digest"),
			}}

			relocatedImageProvider := NewDefaultRelocatedImageProvider(fetcher)
			keychain := &registryfakes.FakeKeychain{Name: "someKeychain"}
			kpConfig := config.NewKpConfig("my-registy.com/my-repo", corev1.ObjectReference{Name: "service account"})

			srcImage := "some-registry.com/some-repo/image@sha256:some-digest"

			image, err := relocatedImageProvider.RelocatedImage(keychain, kpConfig, srcImage)
			require.NoError(t, err)

			assert.Equal(t, "my-registy.com/my-repo@sha256:some-digest", image)
		})
	})
}
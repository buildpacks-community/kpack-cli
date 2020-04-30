package image_test

import (
	"fmt"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/pivotal/build-service-cli/pkg/image"
	"github.com/pivotal/build-service-cli/pkg/registry/fakes"
)

func TestImageUploader(t *testing.T) {
	spec.Run(t, "testImageUploader", testImageUploader)
}

func testImageUploader(t *testing.T, when spec.G, it spec.S) {
	fetcher := &fakes.Fetcher{}
	relocator := &fakes.FakeRelocator{}
	uploader := &image.Uploader{
		Fetcher:   fetcher,
		Relocator: relocator,
	}

	when("image file is provided", func() {
		it("it uploads to registry", func() {
			ref, _, err := uploader.Upload("kpackcr.org/somepath", "new-image-name", "testdata/test-image.tar")
			require.NoError(t, err)

			const expectedRef = "kpackcr.org/somepath/new-image-name@sha256:c486cfa1439f5ca6e19f5572a1c589070f475be1d246a6827fe326cc9e6738c6"
			require.Equal(t, expectedRef, ref)
		})
	})

	when("remote location", func() {
		it("it uploads to registry", func() {
			testImage, err := random.Image(10, 10)
			require.NoError(t, err)

			fetcher.AddImage(testImage, "some/remote-bp")

			ref, _, err := uploader.Upload("kpackcr.org/somepath", "new-image-name", "some/remote-bp")
			require.NoError(t, err)

			digest, err := testImage.Digest()
			require.NoError(t, err)

			expectedRef := fmt.Sprintf("kpackcr.org/somepath/new-image-name@%s", digest)
			require.Equal(t, expectedRef, ref)
		})
	})
}

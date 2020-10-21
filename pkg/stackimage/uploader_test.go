// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package stackimage

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/pivotal/build-service-cli/pkg/image/fakes"
	"github.com/pivotal/build-service-cli/pkg/registry"
)

func TestBuildpackageUploader(t *testing.T) {
	spec.Run(t, "testBuildpackageUploader", testBuildpackageUploader)
}

func testBuildpackageUploader(t *testing.T, when spec.G, it spec.S) {
	fetcher := &fakes.Fetcher{}
	relocator := &fakes.Relocator{}
	uploader := &Uploader{
		Fetcher:   fetcher,
		Relocator: relocator,
	}

	when("UploadStackImages", func() {
		it("it uploads to registry", func() {
			testBuildImage, err := random.Image(10, 10)
			require.NoError(t, err)
			testRunImage, err := random.Image(10, 10)
			require.NoError(t, err)

			fetcher.AddImage("some/remote-build", testBuildImage)
			fetcher.AddImage("some/remote-run", testRunImage)

			bldDigest, err := testBuildImage.Digest()
			require.NoError(t, err)
			runDigest, err := testRunImage.Digest()
			require.NoError(t, err)

			bldImage, runImage, err := uploader.UploadStackImages("some/remote-build", "some/remote-run", "kpackcr.org/somepath", registry.TLSConfig{}, ioutil.Discard)
			require.NoError(t, err)

			expectedBldImage := fmt.Sprintf("kpackcr.org/somepath/build@%s", bldDigest)
			expectedRunImage := fmt.Sprintf("kpackcr.org/somepath/run@%s", runDigest)
			require.Equal(t, expectedBldImage, bldImage)
			require.Equal(t, expectedRunImage, runImage)
			require.Equal(t, 2, relocator.CallCount())
		})
	})

	when("ValidateStackIDs", func() {
		it("returns no error with same id", func() {
			testBuildImage, err := random.Image(10, 10)
			require.NoError(t, err)
			testRunImage, err := random.Image(10, 10)
			require.NoError(t, err)

			testBuildImage, err = imagehelpers.SetStringLabel(testBuildImage, "io.buildpacks.stack.id", "some-id")
			require.NoError(t, err)
			testRunImage, err = imagehelpers.SetStringLabel(testRunImage, "io.buildpacks.stack.id", "some-id")
			require.NoError(t, err)

			fetcher.AddImage("some/remote-build", testBuildImage)
			fetcher.AddImage("some/remote-run", testRunImage)

			stackID, err := uploader.ValidateStackIDs("some/remote-build", "some/remote-run", registry.TLSConfig{})
			require.NoError(t, err)

			require.Equal(t, "some-id", stackID)
		})

		it("returns error when ids differ", func() {
			testBuildImage, err := random.Image(10, 10)
			require.NoError(t, err)
			testRunImage, err := random.Image(10, 10)
			require.NoError(t, err)

			testBuildImage, err = imagehelpers.SetStringLabel(testBuildImage, "io.buildpacks.stack.id", "some-id")
			require.NoError(t, err)
			testRunImage, err = imagehelpers.SetStringLabel(testRunImage, "io.buildpacks.stack.id", "some-other-id")
			require.NoError(t, err)

			fetcher.AddImage("some/remote-build", testBuildImage)
			fetcher.AddImage("some/remote-run", testRunImage)

			_, err = uploader.ValidateStackIDs("some/remote-build", "some/remote-run", registry.TLSConfig{})
			require.EqualError(t, err, "build stack 'some-id' does not match run stack 'some-other-id'")
		})
	})

	when("UploadedBuildImageRef", func() {
		it("it returns the relocated build image reference without relocating", func() {
			testImage, err := random.Image(10, 10)
			require.NoError(t, err)

			fetcher.AddImage("some/remote", testImage)

			ref, err := uploader.UploadedBuildImageRef("some/remote", "kpackcr.org/somepath", registry.TLSConfig{})
			require.NoError(t, err)

			digest, err := testImage.Digest()
			require.NoError(t, err)

			expectedImage := fmt.Sprintf("kpackcr.org/somepath/build@%s", digest)
			require.Equal(t, expectedImage, ref)
		})
	})

	when("UploadedRunImageRef", func() {
		it("it returns the relocated run image reference without relocating", func() {
			testImage, err := random.Image(10, 10)
			require.NoError(t, err)

			fetcher.AddImage("some/remote", testImage)

			ref, err := uploader.UploadedRunImageRef("some/remote", "kpackcr.org/somepath", registry.TLSConfig{})
			require.NoError(t, err)

			digest, err := testImage.Digest()
			require.NoError(t, err)

			expectedImage := fmt.Sprintf("kpackcr.org/somepath/run@%s", digest)
			require.Equal(t, expectedImage, ref)
		})
	})
}

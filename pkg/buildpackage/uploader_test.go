// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package buildpackage

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

	when("UploadBuildpackage", func() {
		when("cnb file is provided", func() {
			it("it uploads to registry", func() {

				image, err := uploader.UploadBuildpackage(ioutil.Discard, "kpackcr.org/somepath", "testdata/sample-bp.cnb", registry.TLSConfig{})
				require.NoError(t, err)

				const expectedFixture = "kpackcr.org/somepath/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"
				require.Equal(t, expectedFixture, image)
				require.Equal(t, 1, relocator.CallCount())
			})
		})

		when("remote location", func() {
			it("it uploads to registry", func() {
				testImage, err := random.Image(10, 10)
				require.NoError(t, err)

				testImage, err = imagehelpers.SetStringLabel(testImage, "io.buildpacks.buildpackage.metadata", `{"id": "sample-buildpack/name"}`)
				require.NoError(t, err)

				fetcher.AddImage("some/remote-bp", testImage)

				image, err := uploader.UploadBuildpackage(ioutil.Discard, "kpackcr.org/somepath", "some/remote-bp", registry.TLSConfig{})
				require.NoError(t, err)

				digest, err := testImage.Digest()
				require.NoError(t, err)

				expectedImage := fmt.Sprintf("kpackcr.org/somepath/sample-buildpack_name@%s", digest)
				require.Equal(t, expectedImage, image)
				require.Equal(t, 1, relocator.CallCount())
			})
		})
	})

	when("UploadBuildpackage", func() {
		when("cnb file is provided", func() {
			it("it returns the relocated reference without relocating", func() {

				ref, err := uploader.RelocatedBuildpackage("kpackcr.org/somepath", "testdata/sample-bp.cnb", registry.TLSConfig{})
				require.NoError(t, err)

				const expectedFixture = "kpackcr.org/somepath/sample_buildpackage@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"
				require.Equal(t, expectedFixture, ref)
				require.Equal(t, 0, relocator.CallCount())
			})
		})

		when("remote location", func() {
			it("it returns the relocated reference without relocating", func() {
				testImage, err := random.Image(10, 10)
				require.NoError(t, err)

				testImage, err = imagehelpers.SetStringLabel(testImage, "io.buildpacks.buildpackage.metadata", `{"id": "sample-buildpack/name"}`)
				require.NoError(t, err)

				fetcher.AddImage("some/remote-bp", testImage)

				ref, err := uploader.RelocatedBuildpackage("kpackcr.org/somepath", "some/remote-bp", registry.TLSConfig{})
				require.NoError(t, err)

				digest, err := testImage.Digest()
				require.NoError(t, err)

				expectedImage := fmt.Sprintf("kpackcr.org/somepath/sample-buildpack_name@%s", digest)
				require.Equal(t, expectedImage, ref)
				require.Equal(t, 0, relocator.CallCount())
			})
		})
	})
}

// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package buildpackage

import (
	"fmt"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/pivotal/kpack/pkg/registry/registryfakes"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
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
	fakeKeychain := &registryfakes.FakeKeychain{}

	when("UploadBuildpackage", func() {
		when("cnb file is provided", func() {
			it("it uploads to registry", func() {
				image, metadata, err := uploader.UploadBuildpackage(fakeKeychain, "testdata/sample-bp.cnb", "kpackcr.org/somepath")
				require.NoError(t, err)

				const expectedFixture = "kpackcr.org/somepath@sha256:37d646bec2453ab05fe57288ede904dfd12f988dbc964e3e764c41c1bd3b58bf"
				require.Equal(t, expectedFixture, image)
				require.Equal(t, 1, relocator.CallCount())

				require.Equal(t, Metadata{Id: "sample/buildpackage", Version: "0.0.1"}, metadata)
			})
		})

		when("remote location", func() {
			it("it uploads to registry", func() {
				testImage, err := random.Image(10, 10)
				require.NoError(t, err)

				testImage, err = imagehelpers.SetStringLabel(testImage, "io.buildpacks.buildpackage.metadata", `{"id": "sample-buildpack/name", "version":"some-version"}`)
				require.NoError(t, err)

				fetcher.AddImage("some/remote-bp", testImage)

				image, metadata, err := uploader.UploadBuildpackage(fakeKeychain, "some/remote-bp", "kpackcr.org/somepath")
				require.NoError(t, err)

				digest, err := testImage.Digest()
				require.NoError(t, err)

				expectedImage := fmt.Sprintf("kpackcr.org/somepath@%s", digest)
				require.Equal(t, expectedImage, image)
				require.Equal(t, 1, relocator.CallCount())

				require.Equal(t, Metadata{Id: "sample-buildpack/name", Version: "some-version"}, metadata)
			})
		})
	})
}

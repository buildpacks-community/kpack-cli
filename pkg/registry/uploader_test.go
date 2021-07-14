// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package registry_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/registry/registryfakes"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
)

func TestUploader(t *testing.T) {
	spec.Run(t, "Test Uploader", testUploader)
}

func testUploader(t *testing.T, when spec.G, it spec.S) {
	const (
		testdataDigest = "sha256:f261a0e333140b47e51e1cf5a045e704a47842d56ee33962770001582ac6656a"
		testZipDigest  = "sha256:0a35946b7420d1f6bf0d1397c86cef1f00b08a8575d6045f20914a6845c1376c"
	)

	when("Upload", func() {
		var (
			fakeRelocator = &fakes.Relocator{}
			uploader      = registry.DefaultSourceUploader{
				Relocator: fakeRelocator,
			}
		)

		it("relocates local contents to registry", func() {
			_, err := uploader.Upload(&registryfakes.FakeKeychain{}, "myregistry.com/blah", "testdata/sample")
			require.NoError(t, err)

			require.Equal(t, 1, fakeRelocator.CallCount())

			_, image, _ := fakeRelocator.RelocateCall(0)
			digest, err := image.Digest()
			require.NoError(t, err)
			require.Equal(t, testdataDigest, digest.String())
		})

		it("relocates local zip to registry", func() {
			_, err := uploader.Upload(&registryfakes.FakeKeychain{}, "myregistry.com/blah", "testdata/sample.zip")
			require.NoError(t, err)

			require.Equal(t, 1, fakeRelocator.CallCount())

			_, image, _ := fakeRelocator.RelocateCall(0)
			digest, err := image.Digest()
			require.NoError(t, err)
			require.Equal(t, testZipDigest, digest.String())
		})


		it("returns err on path to invalid zip", func() {
			_, err := uploader.Upload(&registryfakes.FakeKeychain{}, "myregistry.com/blah", "testdata/sample/app")
			require.EqualError(t, err, "local path must be a directory or zip")

			require.Equal(t, 0, fakeRelocator.CallCount())
		})

	})
}

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
		testdataDigest = "sha256:22bd7fd05ebd1bca0d8f8f1e97620ed76b7ff31f41d471b71d023259414c0e15"
		testZipDigest  = "sha256:5fcd0d53ec99c419061fb4b92d160fa5531791e9a8660447c958879b0777b89a"
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

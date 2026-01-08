// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package lifecycleimage

import (
	"fmt"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	kpackregistryfakes "github.com/pivotal/kpack/pkg/registry/registryfakes"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	registryfakes "github.com/buildpacks-community/kpack-cli/pkg/registry/fakes"
)

func TestLifecycleUploader(t *testing.T) {
	spec.Run(t, "testLifecycleUploader", testLifecycleUploader)
}

func testLifecycleUploader(t *testing.T, when spec.G, it spec.S) {
	fetcher := &registryfakes.Fetcher{}
	relocator := &registryfakes.Relocator{}
	uploader := &Uploader{
		Fetcher:   fetcher,
		Relocator: relocator,
	}
	fakeKeychain := &kpackregistryfakes.FakeKeychain{}

	when("UploadLifecycleImage", func() {
		it("uploads lifecycle image to registry", func() {
			testLifecycleImage, err := random.Image(10, 10)
			require.NoError(t, err)

			fetcher.AddImage("some/remote-lifecycle", testLifecycleImage)

			lifecycleDigest, err := testLifecycleImage.Digest()
			require.NoError(t, err)

			relocatedImage, err := uploader.UploadLifecycleImage(fakeKeychain, "some/remote-lifecycle", "kpackcr.org/lifecycle")
			require.NoError(t, err)

			expectedImage := fmt.Sprintf("kpackcr.org/lifecycle@%s", lifecycleDigest)
			require.Equal(t, expectedImage, relocatedImage)
			require.Equal(t, 1, relocator.CallCount())
		})
	})

	when("ValidateLifecycleImage", func() {
		it("returns no error when lifecycle has required labels", func() {
			testLifecycleImage, err := random.Image(10, 10)
			require.NoError(t, err)

			testLifecycleImage, err = imagehelpers.SetStringLabel(testLifecycleImage, "io.buildpacks.lifecycle.version", "0.17.0")
			require.NoError(t, err)
			testLifecycleImage, err = imagehelpers.SetStringLabel(testLifecycleImage, "io.buildpacks.lifecycle.apis", `{"buildpack":{"supported":["0.2","0.10"]},"platform":{"supported":["0.3","0.12"]}}`)
			require.NoError(t, err)

			fetcher.AddImage("some/remote-lifecycle", testLifecycleImage)

			err = uploader.ValidateLifecycleImage(fakeKeychain, "some/remote-lifecycle")
			require.NoError(t, err)
		})

		it("returns error when lifecycle missing version label", func() {
			testLifecycleImage, err := random.Image(10, 10)
			require.NoError(t, err)

			testLifecycleImage, err = imagehelpers.SetStringLabel(testLifecycleImage, "io.buildpacks.lifecycle.apis", `{"buildpack":{"supported":["0.2","0.10"]},"platform":{"supported":["0.3","0.12"]}}`)
			require.NoError(t, err)

			fetcher.AddImage("some/remote-lifecycle", testLifecycleImage)

			err = uploader.ValidateLifecycleImage(fakeKeychain, "some/remote-lifecycle")
			require.EqualError(t, err, "missing label io.buildpacks.lifecycle.version")
		})

		it("returns error when lifecycle missing apis label", func() {
			testLifecycleImage, err := random.Image(10, 10)
			require.NoError(t, err)

			testLifecycleImage, err = imagehelpers.SetStringLabel(testLifecycleImage, "io.buildpacks.lifecycle.version", "0.17.0")
			require.NoError(t, err)

			fetcher.AddImage("some/remote-lifecycle", testLifecycleImage)

			err = uploader.ValidateLifecycleImage(fakeKeychain, "some/remote-lifecycle")
			require.EqualError(t, err, "missing label io.buildpacks.lifecycle.apis")
		})
	})
}

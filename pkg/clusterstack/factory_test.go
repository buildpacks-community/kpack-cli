// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"fmt"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/pivotal/build-service-cli/pkg/image/fakes"
	"github.com/pivotal/build-service-cli/pkg/registry"
)

func TestFactory(t *testing.T) {
	spec.Run(t, "testFactory", testFactory)
}

func testFactory(t *testing.T, when spec.G, it spec.S) {
	fetcher := &fakes.Fetcher{}
	relocator := &fakes.Relocator{}
	factory := &Factory{
		Fetcher:    fetcher,
		Relocator:  relocator,
		TLSConfig:  registry.TLSConfig{},
		Repository: "kpackcr.org/somepath",
	}

	when("RelocatedBuildImage", func() {
		it("it returns the relocated references for build and run", func() {
			testImage, err := random.Image(10, 10)
			require.NoError(t, err)

			testImage, err = imagehelpers.SetStringLabel(testImage, "io.buildpacks.buildpackage.metadata", `{"id": "sample-buildpack/name"}`)
			require.NoError(t, err)

			fetcher.AddImage("some/remote-bp", testImage)

			digest, err := testImage.Digest()
			require.NoError(t, err)

			buildRef, err := factory.RelocatedBuildImage("some/remote-bp")
			require.NoError(t, err)
			runRef, err := factory.RelocatedRunImage("some/remote-bp")
			require.NoError(t, err)

			expectedBuildImage := fmt.Sprintf("kpackcr.org/somepath/build@%s", digest)
			expectedRunImage := fmt.Sprintf("kpackcr.org/somepath/run@%s", digest)

			require.Equal(t, expectedBuildImage, buildRef)
			require.Equal(t, expectedRunImage, runRef)
		})
	})
}

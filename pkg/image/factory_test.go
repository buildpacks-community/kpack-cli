// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/pivotal/build-service-cli/pkg/image"
	srcfakes "github.com/pivotal/build-service-cli/pkg/source/fakes"
)

func TestImageFactory(t *testing.T) {
	spec.Run(t, "TestImageFactory", testImageFactory)
}

func testImageFactory(t *testing.T, when spec.G, it spec.S) {
	factory := &image.Factory{
		SourceUploader: &srcfakes.SourceUploader{
			ImageRef: "",
		},
	}

	it("sets type metadata", func() {
		factory.Blob = "some-blob"
		img, err := factory.MakeImage("test-name", "test-namespace", "test-registry.io/test-image")
		require.NoError(t, err)

		require.Equal(t, "Image", img.Kind)
		require.Equal(t, "build.pivotal.io/v1alpha1", img.APIVersion)
	})

	when("no params are set", func() {
		it("returns an error message", func() {
			_, err := factory.MakeImage("test-name", "test-namespace", "test-registry.io/test-image")
			require.EqualError(t, err, "image source must be one of git, blob, or local-path")
		})
	})

	when("too many params are set", func() {
		it("returns an error message", func() {
			factory.GitRepo = "some-git-repo"
			factory.Blob = "some-blob"
			factory.LocalPath = "some-local-path"
			_, err := factory.MakeImage("test-name", "test-namespace", "test-registry.io/test-image")
			require.EqualError(t, err, "image source must be one of git, blob, or local-path")
		})
	})

	when("git is missing git revision", func() {
		it("returns an error message", func() {
			factory.GitRepo = "some-dockerhub-id"
			_, err := factory.MakeImage("test-name", "test-namespace", "test-registry.io/test-image")
			require.EqualError(t, err, "missing parameter git-revision")
		})
	})

	when("both builder and cluster builder are provided", func() {
		it("returns an error message", func() {
			factory.Blob = "some-blob"
			factory.Builder = "some-builder"
			factory.ClusterBuilder = "some-cluster-builder"
			_, err := factory.MakeImage("test-name", "test-namespace", "test-registry.io/test-image")
			require.EqualError(t, err, "must provide one of builder or cluster-builder")
		})
	})

	when("an env var has an equal sign in the value", func() {
		it("handles the env var", func() {
			factory.Blob = "some-blob"
			factory.Env = append(factory.Env, `BP_MAVEN_BUILD_ARGUMENTS="-Dmaven.test.skip=true -Pk8s package"`)
			img, err := factory.MakeImage("test-name", "test-namespace", "test-registry.io/test-image")
			require.NoError(t, err)
			require.Len(t, img.Env(), 1)
			require.Equal(t, "BP_MAVEN_BUILD_ARGUMENTS", img.Env()[0].Name)
			require.Equal(t, `"-Dmaven.test.skip=true -Pk8s package"`, img.Env()[0].Value)
		})
	})
}

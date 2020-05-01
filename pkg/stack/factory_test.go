package stack_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/pivotal/build-service-cli/pkg/stack"
)

func TestStackFactory(t *testing.T) {
	spec.Run(t, "TestStackFactory", testStackFactory)
}

func testStackFactory(t *testing.T, when spec.G, it spec.S) {
	factory := &stack.Factory{}

	when("no params are set", func() {
		it("returns an error message", func() {
			_, err := factory.MakeStack("test-name")
			require.EqualError(t, err, "image source must be one of git, blob, or local-path")
		})
	})

	when("too many params are set", func() {
		it("returns an error message", func() {
			//factory.GitRepo = "some-git-repo"
			//factory.Blob = "some-blob"
			//factory.LocalPath = "some-local-path"
			//_, err := factory.MakeStack("test-name", "test-namespace", "test-registry.io/test-image")
			//require.EqualError(t, err, "image source must be one of git, blob, or local-path")
		})
	})

	when("git is missing git revision", func() {
		it("returns an error message", func() {
			//factory.GitRepo = "some-dockerhub-id"
			//_, err := factory.MakeStack("test-name", "test-namespace", "test-registry.io/test-image")
			//require.EqualError(t, err, "missing parameter git-revision")
		})
	})

	when("both builder and cluster builder are provided", func() {
		it("returns an error message", func() {
			//factory.Blob = "some-blob"
			//factory.Builder = "some-builder"
			//factory.ClusterBuilder = "some-cluster-builder"
			//_, err := factory.MakeStack("test-name", "test-namespace", "test-registry.io/test-image")
			//require.EqualError(t, err, "must provide one of builder or cluster-builder")
		})
	})
}

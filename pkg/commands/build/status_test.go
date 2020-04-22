package build_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/commands/build"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestBuildStatusCommand(t *testing.T) {
	spec.Run(t, "TestBuildStatusCommand", testBuildStatusCommand)
}

func testBuildStatusCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		image                       = "test-image"
		defaultNamespace            = "some-default-namespace"
		expectedOutputForMostRecent = `Image:      repo.com/image-3:tag
Status:     BUILDING
Reasons:    TRIGGER

Builder:      some-repo.com/my-builder
Run Image:    some-repo.com/run-image

Source:    Local Source

BUILDPACK ID    BUILDPACK VERSION
bp-id-1         bp-version-1
bp-id-2         bp-version-2
`
		expectedOutputForBuildNumber = `Image:      repo.com/image-1:tag
Status:     SUCCESS
Reasons:    CONFIG

Builder:      some-repo.com/my-builder
Run Image:    some-repo.com/run-image

Source:    Local Source

BUILDPACK ID    BUILDPACK VERSION
bp-id-1         bp-version-1
bp-id-2         bp-version-2
`
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return build.NewStatusCommand(clientSet, defaultNamespace)
	}

	when("getting build status", func() {
		when("in the default namespace", func() {
			when("the build exists", func() {
				when("the build flag is provided", func() {
					it("shows the build status", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.MakeTestBuilds(image, defaultNamespace),
							Args:           []string{image, "-b", "1"},
							ExpectedOutput: expectedOutputForBuildNumber,
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the build flag is not provided", func() {
					it("shows the build status of the most recent build", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.MakeTestBuilds(image, defaultNamespace),
							Args:           []string{image},
							ExpectedOutput: expectedOutputForMostRecent,
						}.TestKpack(t, cmdFunc)
					})
				})
			})

			when("the build does not exist", func() {
				when("the build flag is provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.MakeTestBuilds(image, defaultNamespace),
							Args:           []string{image, "-b", "123"},
							ExpectErr:      true,
							ExpectedOutput: "Error: build \"123\" not found\n",
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the build flag was not provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Args:           []string{image},
							ExpectErr:      true,
							ExpectedOutput: "Error: no builds found\n",
						}.TestKpack(t, cmdFunc)
					})
				})
			})
		})

		when("in a given namespace", func() {
			const namespace = "some-namespace"

			when("the build exists", func() {
				when("the build flag is provided", func() {
					it("gets the build status", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.MakeTestBuilds(image, namespace),
							Args:           []string{image, "-b", "1", "-n", namespace},
							ExpectedOutput: expectedOutputForBuildNumber,
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the build flag is not provided", func() {
					it("shows the build status of the most recent build", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.MakeTestBuilds(image, namespace),
							Args:           []string{image, "-n", namespace},
							ExpectedOutput: expectedOutputForMostRecent,
						}.TestKpack(t, cmdFunc)
					})
				})
			})

			when("the build does not exist", func() {
				when("the build flag is provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.MakeTestBuilds(image, namespace),
							Args:           []string{image, "-b", "123", "-n", namespace},
							ExpectErr:      true,
							ExpectedOutput: "Error: build \"123\" not found\n",
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the build flag was not provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Args:           []string{image, "-n", namespace},
							ExpectErr:      true,
							ExpectedOutput: "Error: no builds found\n",
						}.TestKpack(t, cmdFunc)
					})
				})
			})
		})
	})
}

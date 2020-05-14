package build_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/commands/build"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestBuildListCommand(t *testing.T) {
	spec.Run(t, "TestBuildListCommand", testBuildListCommand)
}

func testBuildListCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		image            = "test-image"
		defaultNamespace = "some-default-namespace"
		expectedOutput   = `BUILD    STATUS      IMAGE                   STARTED                FINISHED               REASON
1        SUCCESS     repo.com/image-1:tag    0001-01-01 00:00:00    0001-01-01 00:00:00    CONFIG
2        FAILURE     repo.com/image-2:tag    0001-01-01 01:00:00    0001-01-01 00:00:00    COMMIT+
3        BUILDING    repo.com/image-3:tag    0001-01-01 05:00:00                           TRIGGER

`
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		contextProvider := testhelpers.NewFakeKpackContextProvider(defaultNamespace, clientSet)
		return build.NewListCommand(contextProvider)
	}

	when("listing builds", func() {
		when("in the default namespace", func() {
			when("there are builds", func() {
				it("lists the builds", func() {
					testhelpers.CommandTest{
						Objects:        testhelpers.MakeTestBuilds(image, defaultNamespace),
						Args:           []string{image},
						ExpectedOutput: expectedOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("there are no builds", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						Args:           []string{image},
						ExpectErr:      true,
						ExpectedOutput: "Error: no builds found\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})

		when("in a given namespace", func() {
			const namespace = "some-namespace"

			when("there are builds", func() {
				it("lists the builds", func() {
					testhelpers.CommandTest{
						Objects:        testhelpers.MakeTestBuilds(image, namespace),
						Args:           []string{image, "-n", namespace},
						ExpectedOutput: expectedOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("there are no builds", func() {
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
}

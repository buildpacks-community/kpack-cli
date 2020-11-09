// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

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
		expectedOutput   = `BUILD    STATUS      IMAGE                         REASON
1        SUCCESS     repo.com/image-1:tag          CONFIG
2        FAILURE     repo.com/image-2:tag          COMMIT+
3        BUILDING    repo.com/image-3:tag          TRIGGER
1        BUILDING    repo.com/other-image-1:tag    UNKNOWN

`
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return build.NewListCommand(clientSetProvider)
	}

	when("listing builds", func() {
		when("in the default namespace", func() {
			when("there are builds", func() {
				it("lists the builds", func() {
					testhelpers.CommandTest{
						Objects:        testhelpers.MakeTestBuilds(image, defaultNamespace),
						Args:           nil,
						ExpectedOutput: expectedOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("there are no builds", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						Args:           nil,
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
						Args:           []string{"-n", namespace},
						ExpectedOutput: expectedOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("there are no builds", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						Args:           []string{"-n", namespace},
						ExpectErr:      true,
						ExpectedOutput: "Error: no builds found\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})

		when("an image is specified", func() {
			const expectedOutput = `BUILD    STATUS      IMAGE                   REASON
1        SUCCESS     repo.com/image-1:tag    CONFIG
2        FAILURE     repo.com/image-2:tag    COMMIT+
3        BUILDING    repo.com/image-3:tag    TRIGGER

`
			when("there are builds", func() {
				it("lists the builds of the image", func() {
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
	})
}

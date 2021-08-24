// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package build_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/build"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestBuildLogsCommand(t *testing.T) {
	spec.Run(t, "TestBuildLogsCommand", testBuildLogsCommand)
}

func testBuildLogsCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		image            = "test-image"
		defaultNamespace = "some-default-namespace"
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return build.NewLogsCommand(clientSetProvider)
	}

	when("getting build logs", func() {
		when("in the default namespace", func() {
			when("the build does not exist", func() {
				when("the build flag is provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.BuildsToRuntimeObjs(testhelpers.MakeTestBuilds(image, defaultNamespace)),
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
			when("the build does not exist", func() {
				when("the build flag is provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.BuildsToRuntimeObjs(testhelpers.MakeTestBuilds(image, namespace)),
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

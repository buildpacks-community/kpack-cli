// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package buildpack_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/buildpack"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestBuildpackDeleteCommand(t *testing.T) {
	spec.Run(t, "TestBuildpackDeleteCommand", testBuildpackDeleteCommand)
}

func testBuildpackDeleteCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return buildpack.NewDeleteCommand(clientSetProvider)
	}

	when("a namespace has been provided", func() {
		when("a buildpack is available", func() {
			it("deletes the buildpack", func() {
				bp := &v1alpha2.Buildpack{
					ObjectMeta: v1.ObjectMeta{
						Name:      "some-buildpack",
						Namespace: "test-namespace",
					},
				}
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						bp,
					},
					Args: []string{"-n", "test-namespace", "some-buildpack"},
					ExpectedOutput: `Buildpack "some-buildpack" deleted
`,
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							ActionImpl: clientgotesting.ActionImpl{
								Namespace: "test-namespace",
							},
							Name: bp.Name,
						},
					},
				}.TestKpack(t, cmdFunc)
			})
		})
		when("a buildpack is not available", func() {
			it("returns an error", func() {
				testhelpers.CommandTest{
					Objects: nil,
					Args:    []string{"-n", "test-namespace", "some-buildpack"},
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							ActionImpl: clientgotesting.ActionImpl{
								Namespace: "test-namespace",
							},

							Name: "some-buildpack",
						},
					},
					ExpectedErrorOutput: "Error: buildpacks.kpack.io \"some-buildpack\" not found\n",
					ExpectErr:           true,
				}.TestKpack(t, cmdFunc)
			})
		})
	})

	when("a namespace has not been provided", func() {
		when("a buildpack is available", func() {
			it("deletes the buildpack", func() {
				bp := &v1alpha2.Buildpack{
					ObjectMeta: v1.ObjectMeta{
						Name:      "some-buildpack",
						Namespace: defaultNamespace,
					},
				}
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						bp,
					},
					Args: []string{"some-buildpack"},
					ExpectedOutput: `Buildpack "some-buildpack" deleted
`,
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							ActionImpl: clientgotesting.ActionImpl{
								Namespace: defaultNamespace,
							},

							Name: bp.Name,
						},
					},
				}.TestKpack(t, cmdFunc)
			})
		})
		when("a buildpack is not available", func() {
			it("returns an error", func() {
				testhelpers.CommandTest{
					Objects: nil,
					Args:    []string{"some-buildpack"},
					ExpectDeletes: []clientgotesting.DeleteActionImpl{
						{
							ActionImpl: clientgotesting.ActionImpl{
								Namespace: defaultNamespace,
							},

							Name: "some-buildpack",
						},
					},
					ExpectedErrorOutput: "Error: buildpacks.kpack.io \"some-buildpack\" not found\n",
					ExpectErr:           true,
				}.TestKpack(t, cmdFunc)
			})
		})
	})

}

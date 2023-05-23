// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuildpack_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterbuildpack"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestClusterBuildpackDeleteCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuildpackDeleteCommand", testClusterBuildpackDeleteCommand)
}

func testClusterBuildpackDeleteCommand(t *testing.T, when spec.G, it spec.S) {
	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterbuildpack.NewDeleteCommand(clientSetProvider)
	}

	when("a buildpack is available", func() {
		it("deletes the buildpack", func() {
			bp := &v1alpha2.ClusterBuildpack{
				ObjectMeta: v1.ObjectMeta{
					Name: "some-buildpack",
				},
			}
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					bp,
				},
				Args: []string{"some-buildpack"},
				ExpectedOutput: `Cluster Buildpack "some-buildpack" deleted
`,
				ExpectDeletes: []clientgotesting.DeleteActionImpl{
					{
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
						Name: "some-buildpack",
					},
				},
				ExpectedErrorOutput: "Error: clusterbuildpacks.kpack.io \"some-buildpack\" not found\n",
				ExpectErr:           true,
			}.TestKpack(t, cmdFunc)
		})
	})
}

// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterbuilder"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestClusterBuilderDeleteCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuilderDeleteCommand", testClusterBuilderDeleteCommand)
}

func testClusterBuilderDeleteCommand(t *testing.T, when spec.G, it spec.S) {

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterbuilder.NewDeleteCommand(clientSetProvider)
	}

	when("a clusterbuilder is available", func() {
		it("deletes the clusterbuilder", func() {
			clusterBuilder := &v1alpha1.ClusterBuilder{
				ObjectMeta: v1.ObjectMeta{
					Name: "some-clusterbuilder",
				},
			}
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					clusterBuilder,
				},
				Args: []string{"some-clusterbuilder"},
				ExpectedOutput: `ClusterBuilder "some-clusterbuilder" deleted
`,
				ExpectDeletes: []clientgotesting.DeleteActionImpl{
					{
						Name: clusterBuilder.Name,
					},
				},
			}.TestKpack(t, cmdFunc)
		})
	})

	when("a clusterbuilder is not available", func() {
		it("returns an error", func() {
			testhelpers.CommandTest{
				Objects: nil,
				Args:    []string{"some-clusterbuilder"},
				ExpectDeletes: []clientgotesting.DeleteActionImpl{
					{
						Name: "some-clusterbuilder",
					},
				},
				ExpectedOutput: "Error: clusterbuilders.kpack.io \"some-clusterbuilder\" not found\n",
				ExpectErr:      true,
			}.TestKpack(t, cmdFunc)
		})
	})
}

// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterstack"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestClusterStackDeleteCommand(t *testing.T) {
	spec.Run(t, "TestClusterStackDeleteCommand", testClusterStackDeleteCommand)
}

func testClusterStackDeleteCommand(t *testing.T, when spec.G, it spec.S) {

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterstack.NewDeleteCommand(clientSetProvider)
	}

	when("a stack is available", func() {
		it("deletes the stack", func() {
			stack := &v1alpha2.ClusterStack{
				ObjectMeta: v1.ObjectMeta{
					Name: "some-stack",
				},
			}
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					stack,
				},
				Args:           []string{"some-stack"},
				ExpectedOutput: "ClusterStack \"some-stack\" deleted\n",
				ExpectDeletes: []clientgotesting.DeleteActionImpl{
					{
						Name: stack.Name,
					},
				},
			}.TestKpack(t, cmdFunc)
		})
	})

	when("a stack is not available", func() {
		it("returns an error", func() {
			testhelpers.CommandTest{
				Objects: nil,
				Args:    []string{"some-stack"},
				ExpectDeletes: []clientgotesting.DeleteActionImpl{
					{
						Name: "some-stack",
					},
				},
				ExpectedErrorOutput: "Error: clusterstacks.kpack.io \"some-stack\" not found\n",
				ExpectErr:           true,
			}.TestKpack(t, cmdFunc)
		})
	})
}

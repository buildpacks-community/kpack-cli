// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package stack_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/commands/stack"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestStackDeleteCommand(t *testing.T) {
	spec.Run(t, "TestStackDeleteCommand", testStackDeleteCommand)
}

func testStackDeleteCommand(t *testing.T, when spec.G, it spec.S) {

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return stack.NewDeleteCommand(clientSetProvider)
	}

	when("a stack is available", func() {
		it("deletes the stack", func() {
			stack := &expv1alpha1.Stack{
				ObjectMeta: v1.ObjectMeta{
					Name: "some-stack",
				},
			}
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					stack,
				},
				Args:           []string{"some-stack"},
				ExpectedOutput: "\"some-stack\" deleted\n",
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
				ExpectedOutput: "Error: stacks.experimental.kpack.pivotal.io \"some-stack\" not found\n",
				ExpectErr:      true,
			}.TestKpack(t, cmdFunc)
		})
	})
}

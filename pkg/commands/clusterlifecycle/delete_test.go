// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterlifecycle_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/buildpacks-community/kpack-cli/pkg/commands/clusterlifecycle"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
)

func TestClusterLifecycleDeleteCommand(t *testing.T) {
	spec.Run(t, "TestClusterLifecycleDeleteCommand", testClusterLifecycleDeleteCommand)
}

func testClusterLifecycleDeleteCommand(t *testing.T, when spec.G, it spec.S) {

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterlifecycle.NewDeleteCommand(clientSetProvider)
	}

	when("a lifecycle is available", func() {
		it("deletes the lifecycle", func() {
			lifecycle := &v1alpha2.ClusterLifecycle{
				ObjectMeta: v1.ObjectMeta{
					Name: "some-lifecycle",
				},
			}
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					lifecycle,
				},
				Args:           []string{"some-lifecycle"},
				ExpectedOutput: "ClusterLifecycle \"some-lifecycle\" deleted\n",
				ExpectDeletes: []clientgotesting.DeleteActionImpl{
					{
						Name: lifecycle.Name,
					},
				},
			}.TestKpack(t, cmdFunc)
		})
	})

	when("a lifecycle is not available", func() {
		it("returns an error", func() {
			testhelpers.CommandTest{
				Objects: nil,
				Args:    []string{"some-lifecycle"},
				ExpectDeletes: []clientgotesting.DeleteActionImpl{
					{
						Name: "some-lifecycle",
					},
				},
				ExpectedErrorOutput: "Error: clusterlifecycles.kpack.io \"some-lifecycle\" not found\n",
				ExpectErr:           true,
			}.TestKpack(t, cmdFunc)
		})
	})
}

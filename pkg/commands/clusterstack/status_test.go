// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterStackStatusCommand(t *testing.T) {
	spec.Run(t, "TestClusterStackStatusCommand", testClusterStackStatusCommand)
}

func testClusterStackStatusCommand(t *testing.T, when spec.G, it spec.S) {
	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterstack.NewStatusCommand(clientSetProvider)
	}

	when("the stack exists", func() {
		stck := &expv1alpha1.ClusterStack{
			ObjectMeta: metav1.ObjectMeta{
				Name: "some-stack",
			},
			Status: expv1alpha1.ClusterStackStatus{
				ResolvedClusterStack: expv1alpha1.ResolvedClusterStack{
					Id: "some-stack-id",
					BuildImage: expv1alpha1.ClusterStackStatusImage{
						LatestImage: "some-run-image",
					},
					RunImage: expv1alpha1.ClusterStackStatusImage{
						LatestImage: "some-build-image",
					},
					Mixins: []string{"mixin1", "mixin2"},
				},
			},
		}
		it("returns stack details", func() {
			const expectedOutput = `Status:         Unknown
Id:             some-stack-id
Run Image:      some-build-image
Build Image:    some-run-image

`

			testhelpers.CommandTest{
				Objects:        append([]runtime.Object{stck}),
				Args:           []string{"some-stack"},
				ExpectedOutput: expectedOutput,
			}.TestKpack(t, cmdFunc)
		})

		it("includes mixins when --verbose flag is used", func() {
			const expectedOutput = `Status:         Unknown
Id:             some-stack-id
Run Image:      some-build-image
Build Image:    some-run-image
Mixins:         mixin1, mixin2

`

			testhelpers.CommandTest{
				Objects:        append([]runtime.Object{stck}),
				Args:           []string{"some-stack", "--verbose"},
				ExpectedOutput: expectedOutput,
			}.TestKpack(t, cmdFunc)
		})
	})

	when("the stack does not exist", func() {
		it("returns a message that there is no stack", func() {
			testhelpers.CommandTest{
				Args:           []string{"stack-does-not-exist"},
				ExpectErr:      true,
				ExpectedOutput: "Error: clusterstacks.experimental.kpack.pivotal.io \"stack-does-not-exist\" not found\n",
			}.TestKpack(t, cmdFunc)

		})
	})
}

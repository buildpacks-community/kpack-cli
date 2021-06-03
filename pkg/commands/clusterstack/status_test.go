// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterstack"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
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
		stck := &v1alpha1.ClusterStack{
			ObjectMeta: metav1.ObjectMeta{
				Name: "some-stack",
			},
			Status: v1alpha1.ClusterStackStatus{
				ResolvedClusterStack: v1alpha1.ResolvedClusterStack{
					Id: "some-stack-id",
					BuildImage: v1alpha1.ClusterStackStatusImage{
						LatestImage: "some-run-image",
					},
					RunImage: v1alpha1.ClusterStackStatusImage{
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

		when("the status is not ready", func() {
			it("prints the status message", func() {
				stck.Status.Conditions = append(stck.Status.Conditions, corev1alpha1.Condition{
					Type:    corev1alpha1.ConditionReady,
					Status:  corev1.ConditionFalse,
					Message: "some sample message",
				})

				const expectedOutput = `Status:         Not Ready - some sample message
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
		})
	})

	when("the stack does not exist", func() {
		it("returns a message that there is no stack", func() {
			testhelpers.CommandTest{
				Args:           []string{"stack-does-not-exist"},
				ExpectErr:      true,
				ExpectedOutput: "Error: clusterstacks.kpack.io \"stack-does-not-exist\" not found\n",
			}.TestKpack(t, cmdFunc)

		})
	})
}

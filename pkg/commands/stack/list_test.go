// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package stack_test

import (
	"testing"

	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/stack"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestStackListCommand(t *testing.T) {
	spec.Run(t, "TestStackListCommand", testStackListCommand)
}

func testStackListCommand(t *testing.T, when spec.G, it spec.S) {
	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return stack.NewListCommand(clientSetProvider)
	}

	when("the namespaces has images", func() {
		it("returns a table of image details", func() {
			stack1 := &expv1alpha1.Stack{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-stack-1",
				},
				Status: expv1alpha1.StackStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionFalse,
							},
						},
					},
					ResolvedStack: expv1alpha1.ResolvedStack{
						Id: "stack-id-1",
					},
				},
			}
			stack2 := &expv1alpha1.Stack{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-stack-2",
				},
				Status: expv1alpha1.StackStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
					ResolvedStack: expv1alpha1.ResolvedStack{
						Id: "stack-id-2",
					},
				},
			}
			stack3 := &expv1alpha1.Stack{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-stack-3",
				},
				Status: expv1alpha1.StackStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionUnknown,
							},
						},
					},
					ResolvedStack: expv1alpha1.ResolvedStack{
						Id: "stack-id-3",
					},
				},
			}

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					stack1,
					stack2,
					stack3,
				},
				ExpectedOutput: `NAME            READY      ID
test-stack-1    False      stack-id-1
test-stack-2    True       stack-id-2
test-stack-3    Unknown    stack-id-3

`,
			}.TestKpack(t, cmdFunc)
		})

		when("there are no stacks", func() {
			it("returns a message that no stacks were found", func() {
				testhelpers.CommandTest{
					ExpectErr:      true,
					ExpectedOutput: "Error: no stacks found\n",
				}.TestKpack(t, cmdFunc)

			})
		})
	})
}

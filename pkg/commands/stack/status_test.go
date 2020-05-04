package stack_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/stack"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestStackStatusCommand(t *testing.T) {
	spec.Run(t, "TestStackStatusCommand", testStackStatusCommand)
}

func testStackStatusCommand(t *testing.T, when spec.G, it spec.S) {
	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return stack.NewStatusCommand(clientSet)
	}

	when("the stack exists", func() {
		it("returns stack details", func() {
			stck := &expv1alpha1.Stack{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-stack",
				},
				Status: expv1alpha1.StackStatus{
					ResolvedStack: expv1alpha1.ResolvedStack{
						Id: "some-stack-id",
						BuildImage: expv1alpha1.StackStatusImage{
							LatestImage: "some-run-image",
						},
						RunImage: expv1alpha1.StackStatusImage{
							LatestImage: "some-build-image",
						},
						Mixins: []string{"mixin1", "mixin2"},
					},
				},
			}

			const expectedOutput = `Status:         Unknown
Id:             some-stack-id
Run Image:      some-build-image
Build Image:    some-run-image
Mixins:         mixin1, mixin2

`

			testhelpers.CommandTest{
				Objects:        append([]runtime.Object{stck}),
				Args:           []string{"some-stack"},
				ExpectedOutput: expectedOutput,
			}.TestKpack(t, cmdFunc)
		})
	})

	when("the stack does not exist", func() {
		it("returns a message that there is no stack", func() {
			testhelpers.CommandTest{
				Args:           []string{"stack-does-not-exist"},
				ExpectErr:      true,
				ExpectedOutput: "Error: stacks.experimental.kpack.pivotal.io \"stack-does-not-exist\" not found\n",
			}.TestKpack(t, cmdFunc)

		})
	})
}

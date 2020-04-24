package clusterbuilder_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/commands/clusterbuilder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterBuilderDeleteCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuilderDeleteCommand", testClusterBuilderDeleteCommand)
}

func testClusterBuilderDeleteCommand(t *testing.T, when spec.G, it spec.S) {

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		return clusterbuilder.NewDeleteCommand(clientSet)
	}

	when("a clusterbuilder is available", func() {
		it("deletes the clusterbuilder", func() {
			clusterBuilder := &expv1alpha1.CustomClusterBuilder{
				ObjectMeta: v1.ObjectMeta{
					Name: "some-clusterbuilder",
				},
			}
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					clusterBuilder,
				},
				Args:           []string{"some-clusterbuilder"},
				ExpectedOutput: "\"some-clusterbuilder\" deleted\n",
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
				ExpectedOutput: "Error: customclusterbuilders.experimental.kpack.pivotal.io \"some-clusterbuilder\" not found\n",
				ExpectErr:      true,
			}.TestKpack(t, cmdFunc)
		})
	})
}

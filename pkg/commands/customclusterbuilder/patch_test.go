package customclusterbuilder_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/customclusterbuilder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestCustomClusterBuilderPatchCommand(t *testing.T) {
	spec.Run(t, "TestCustomClusterBuilderPatchCommand", testCustomClusterBuilderPatchCommand)
}

func testCustomClusterBuilderPatchCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		builder = &expv1alpha1.CustomClusterBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.CustomBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-builder",
			},
			Spec: expv1alpha1.CustomClusterBuilderSpec{
				CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
					Tag:   "some-registry.com/test-builder",
					Stack: "some-stack",
					Store: "some-store",
					Order: []expv1alpha1.OrderEntry{
						{
							Group: []expv1alpha1.BuildpackRef{
								{
									BuildpackInfo: expv1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []expv1alpha1.BuildpackRef{
								{
									BuildpackInfo: expv1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
						},
					},
				},
				ServiceAccountRef: corev1.ObjectReference{
					Name:      "some-service-account",
					Namespace: "some-namespace",
				},
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return customclusterbuilder.NewPatchCommand(clientSetProvider)
	}

	it("patches a CustomClusterBuilder", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				builder,
			},
			Args: []string{
				builder.Name,
				"--stack", "some-other-stack",
				"--store", "some-other-store",
				"--order", "./testdata/patched-order.yaml",
			},
			ExpectedOutput: "\"test-builder\" patched\n",
			ExpectPatches: []string{
				`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"}]},{"group":[{"id":"org.cloudfoundry.fake-bp"}]}],"stack":"some-other-stack","store":"some-other-store"}}`,
			},
		}.TestKpack(t, cmdFunc)
	})

	it("does not patch if there are no changes", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				builder,
			},
			Args: []string{
				builder.Name,
			},
			ExpectedOutput: "nothing to patch\n",
		}.TestKpack(t, cmdFunc)
	})
}

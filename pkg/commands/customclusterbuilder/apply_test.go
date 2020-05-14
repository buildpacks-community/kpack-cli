package customclusterbuilder_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"

	"github.com/pivotal/build-service-cli/pkg/commands/customclusterbuilder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterBuilderApplyCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuilderApplyCommand", testClusterBuilderApplyCommand)
}

func testClusterBuilderApplyCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	var (
		expectedBuilder = &expv1alpha1.CustomClusterBuilder{
			TypeMeta: v1.TypeMeta{
				Kind:       expv1alpha1.CustomClusterBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "test-builder",
			},
			Spec: expv1alpha1.CustomClusterBuilderSpec{
				CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
					Tag:   "some-registry.com/test-builder",
					Stack: "test-stack",
					Store: "test-store",
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
					Namespace: "some-namespace",
					Name:      "some-service-account",
				},
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		contextProvider := testhelpers.NewFakeKpackContextProvider(defaultNamespace, clientSet)
		return customclusterbuilder.NewApplyCommand(contextProvider)
	}

	when("a valid cluster builder config exists", func() {
		it("returns a success message", func() {
			testhelpers.CommandTest{
				Args: []string{"-f", "./testdata/builder.yaml"},
				ExpectedOutput: `"test-builder" applied
`,
				ExpectCreates: []runtime.Object{
					expectedBuilder,
				},
			}.TestKpack(t, cmdFunc)
		})

		when("a valid cluster builder config is applied for an existing cluster builder", func() {
			it("updates the cluster builder", func() {
				existingImage := expectedBuilder.DeepCopy()
				existingImage.Spec.Stack = "some-other-stack"

				testhelpers.CommandTest{
					Args: []string{"-f", "./testdata/builder.yaml"},
					Objects: []runtime.Object{
						existingImage,
					},
					ExpectedOutput: `"test-builder" applied
`,
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedBuilder,
						},
					},
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}

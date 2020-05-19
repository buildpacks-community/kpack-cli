package custombuilder_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/custombuilder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestCustomBuilderPatchCommand(t *testing.T) {
	spec.Run(t, "TestCustomBuilderPatchCommand", testCustomBuilderPatchCommand)
}

func testCustomBuilderPatchCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	var (
		builder = &expv1alpha1.CustomBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.CustomBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder",
				Namespace: "some-namespace",
			},
			Spec: expv1alpha1.CustomNamespacedBuilderSpec{
				CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
					Tag:   "some-registry.com/test-builder",
					Stack: "some-stack",
					Store: "default",
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
				ServiceAccount: "default",
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return custombuilder.NewPatchCommand(clientSetProvider)
	}

	it("patches a CustomBuilder", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				builder,
			},
			Args: []string{
				builder.Name,
				"--stack", "some-other-stack",
				"--order", "./testdata/patched-order.yaml",
				"-n", builder.Namespace,
			},
			ExpectedOutput: "\"test-builder\" patched\n",
			ExpectPatches: []string{
				`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"}]},{"group":[{"id":"org.cloudfoundry.fake-bp"}]}],"stack":"some-other-stack"}}`,
			},
		}.TestKpack(t, cmdFunc)
	})

	it("patches a CustomBuilder in the default namespace", func() {
		builder.Namespace = defaultNamespace

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				builder,
			},
			Args: []string{
				builder.Name,
				"--stack", "some-other-stack",
				"--order", "./testdata/patched-order.yaml",
			},
			ExpectedOutput: "\"test-builder\" patched\n",
			ExpectPatches: []string{
				`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"}]},{"group":[{"id":"org.cloudfoundry.fake-bp"}]}],"stack":"some-other-stack"}}`,
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
				"-n", builder.Namespace,
			},
			ExpectedOutput: "nothing to patch\n",
		}.TestKpack(t, cmdFunc)
	})
}

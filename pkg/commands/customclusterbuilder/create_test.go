// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package customclusterbuilder_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/customclusterbuilder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestCustomClusterBuilderCreateCommand(t *testing.T) {
	spec.Run(t, "TestCustomBuilderCreateCommand", testCustomClusterBuilderCreateCommand)
}

func testCustomClusterBuilderCreateCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		expectedBuilder = &expv1alpha1.CustomClusterBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.CustomClusterBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "test-builder",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry.com/test-builder","stack":"some-stack","store":"some-store","order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"build-service","name":"ccb-service-account"}},"status":{"stack":{}}}`,
				},
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
					Namespace: "build-service",
					Name:      "ccb-service-account",
				},
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return customclusterbuilder.NewCreateCommand(clientSetProvider)
	}

	it("creates a CustomClusterBuilder", func() {
		testhelpers.CommandTest{
			Args: []string{
				expectedBuilder.Name,
				expectedBuilder.Spec.Tag,
				"--stack", expectedBuilder.Spec.Stack,
				"--store", expectedBuilder.Spec.Store,
				"--order", "./testdata/order.yaml",
			},
			ExpectedOutput: `"test-builder" created
`,
			ExpectCreates: []runtime.Object{
				expectedBuilder,
			},
		}.TestKpack(t, cmdFunc)
	})

	it("creates a CustomClusterBuilder with the default stack", func() {
		expectedBuilder.Spec.Stack = "default"
		expectedBuilder.Spec.Store = "default"
		expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry.com/test-builder","stack":"default","store":"default","order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"build-service","name":"ccb-service-account"}},"status":{"stack":{}}}`

		testhelpers.CommandTest{
			Args: []string{
				expectedBuilder.Name,
				expectedBuilder.Spec.Tag,
				"--order", "./testdata/order.yaml",
			},
			ExpectedOutput: "\"test-builder\" created\n",
			ExpectCreates: []runtime.Object{
				expectedBuilder,
			},
		}.TestKpack(t, cmdFunc)
	})

}

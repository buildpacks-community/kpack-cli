// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/builder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestBuilderPatchCommand(t *testing.T) {
	spec.Run(t, "TestBuilderPatchCommand", testBuilderPatchCommand)
}

func testBuilderPatchCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	var (
		bldr = &v1alpha1.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.BuilderKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder",
				Namespace: "some-namespace",
			},
			Spec: v1alpha1.NamespacedBuilderSpec{
				BuilderSpec: v1alpha1.BuilderSpec{
					Tag: "some-registry.com/test-builder",
					Stack: corev1.ObjectReference{
						Name: "some-stack",
						Kind: v1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "some-store",
						Kind: v1alpha1.ClusterStoreKind,
					},
					Order: []v1alpha1.OrderEntry{
						{
							Group: []v1alpha1.BuildpackRef{
								{
									BuildpackInfo: v1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []v1alpha1.BuildpackRef{
								{
									BuildpackInfo: v1alpha1.BuildpackInfo{
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
		return builder.NewPatchCommand(clientSetProvider)
	}

	it("patches a Builder", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				bldr,
			},
			Args: []string{
				bldr.Name,
				"--stack", "some-other-stack",
				"--store", "some-other-store",
				"--order", "./testdata/patched-order.yaml",
				"-n", bldr.Namespace,
			},
			ExpectedOutput: "\"test-builder\" patched\n",
			ExpectPatches: []string{
				`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"}]},{"group":[{"id":"org.cloudfoundry.fake-bp"}]}],"stack":{"name":"some-other-stack"},"store":{"name":"some-other-store"}}}`,
			},
		}.TestKpack(t, cmdFunc)
	})

	it("patches a Builder in the default namespace", func() {
		bldr.Namespace = defaultNamespace

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				bldr,
			},
			Args: []string{
				bldr.Name,
				"--stack", "some-other-stack",
				"--store", "some-other-store",
				"--order", "./testdata/patched-order.yaml",
			},
			ExpectedOutput: "\"test-builder\" patched\n",
			ExpectPatches: []string{
				`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"}]},{"group":[{"id":"org.cloudfoundry.fake-bp"}]}],"stack":{"name":"some-other-stack"},"store":{"name":"some-other-store"}}}`,
			},
		}.TestKpack(t, cmdFunc)
	})

	it("does not patch if there are no changes", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				bldr,
			},
			Args: []string{
				bldr.Name,
				"-n", bldr.Namespace,
			},
			ExpectedOutput: "nothing to patch\n",
		}.TestKpack(t, cmdFunc)
	})
}

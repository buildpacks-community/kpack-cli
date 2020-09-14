// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/clusterbuilder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestClusterBuilderPatchCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuilderPatchCommand", testClusterBuilderPatchCommand)
}

func testClusterBuilderPatchCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		builder = &v1alpha1.ClusterBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.BuilderKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-builder",
			},
			Spec: v1alpha1.ClusterBuilderSpec{
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
				ServiceAccountRef: corev1.ObjectReference{
					Name:      "some-service-account",
					Namespace: "some-namespace",
				},
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterbuilder.NewPatchCommand(clientSetProvider)
	}

	it("patches a ClusterBuilder", func() {
		testhelpers.CommandTest{
			Objects: []runtime.Object{
				builder,
			},
			Args: []string{
				builder.Name,
				"--tag", "some-other-tag",
				"--stack", "some-other-stack",
				"--store", "some-other-store",
				"--order", "./testdata/patched-order.yaml",
			},
			ExpectedOutput: "\"test-builder\" patched\n",
			ExpectPatches: []string{
				`{"spec":{"order":[{"group":[{"id":"org.cloudfoundry.test-bp"}]},{"group":[{"id":"org.cloudfoundry.fake-bp"}]}],"stack":{"name":"some-other-stack"},"store":{"name":"some-other-store"},"tag":"some-other-tag"}}`,
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

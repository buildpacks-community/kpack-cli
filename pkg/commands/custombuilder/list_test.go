// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package custombuilder_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/custombuilder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestBuilderListCommand(t *testing.T) {
	spec.Run(t, "TestBuilderListCommand", testBuilderListCommand)
}

func testBuilderListCommand(t *testing.T, when spec.G, it spec.S) {

	const (
		expectedOutput = `NAME              READY    STACK                          IMAGE
test-builder-1    true     io.buildpacks.stacks.centos    some-registry.com/test-builder-1:tag
test-builder-2    false                                   
test-builder-3    true     io.buildpacks.stacks.bionic    some-registry.com/test-builder-3:tag

`
		defaultNamespace = "some-default-namespace"
	)

	var (
		defaultNamespacedCustomBuilder1 = &expv1alpha1.CustomBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.CustomBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-1",
				Namespace: defaultNamespace,
			},
			Spec: expv1alpha1.CustomNamespacedBuilderSpec{
				CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
					Tag: "some-registry.com/test-builder-1",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: expv1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
						Kind: expv1alpha1.ClusterStoreKind,
					},
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
			},
			Status: expv1alpha1.CustomBuilderStatus{
				BuilderStatus: v1alpha1.BuilderStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
					Stack: v1alpha1.BuildStack{
						RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9203847098234",
						ID:       "io.buildpacks.stacks.centos",
					},
					LatestImage: "some-registry.com/test-builder-1:tag",
				},
			},
		}
		defaultNamespacedCustomBuilder2 = &expv1alpha1.CustomBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.CustomBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-2",
				Namespace: defaultNamespace,
			},
			Spec: expv1alpha1.CustomNamespacedBuilderSpec{
				CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
					Tag: "some-registry.com/test-builder-2",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: expv1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
						Kind: expv1alpha1.ClusterStoreKind,
					},
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
			},
			Status: expv1alpha1.CustomBuilderStatus{
				BuilderStatus: v1alpha1.BuilderStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionFalse,
							},
						},
					},
				},
			},
		}
		defaultNamespacedCustomBuilder3 = &expv1alpha1.CustomBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.CustomBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-3",
				Namespace: defaultNamespace,
			},
			Spec: expv1alpha1.CustomNamespacedBuilderSpec{
				CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
					Tag: "some-registry.com/test-builder-3",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: expv1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
						Kind: expv1alpha1.ClusterStoreKind,
					},
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
			},
			Status: expv1alpha1.CustomBuilderStatus{
				BuilderStatus: v1alpha1.BuilderStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
					Stack: v1alpha1.BuildStack{
						RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9fasdfa847098234",
						ID:       "io.buildpacks.stacks.bionic",
					},
					LatestImage: "some-registry.com/test-builder-3:tag",
				},
			},
		}

		otherNamespacedCustomBuilder1 = &expv1alpha1.CustomBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.CustomBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-1",
				Namespace: "test-namespace",
			},
			Spec: expv1alpha1.CustomNamespacedBuilderSpec{
				CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
					Tag: "some-registry.com/test-builder-1",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: expv1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
						Kind: expv1alpha1.ClusterStoreKind,
					},
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
			},
			Status: expv1alpha1.CustomBuilderStatus{
				BuilderStatus: v1alpha1.BuilderStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
					Stack: v1alpha1.BuildStack{
						RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9203847098234",
						ID:       "io.buildpacks.stacks.centos",
					},
					LatestImage: "some-registry.com/test-builder-1:tag",
				},
			},
		}
		otherNamespacedCustomBuilder2 = &expv1alpha1.CustomBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.CustomBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-2",
				Namespace: "test-namespace",
			},
			Spec: expv1alpha1.CustomNamespacedBuilderSpec{
				CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
					Tag: "some-registry.com/test-builder-2",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: expv1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
						Kind: expv1alpha1.ClusterStoreKind,
					},
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
			},
			Status: expv1alpha1.CustomBuilderStatus{
				BuilderStatus: v1alpha1.BuilderStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionFalse,
							},
						},
					},
				},
			},
		}
		otherNamespacedCustomBuilder3 = &expv1alpha1.CustomBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.CustomBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-3",
				Namespace: "test-namespace",
			},
			Spec: expv1alpha1.CustomNamespacedBuilderSpec{
				CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
					Tag: "some-registry.com/test-builder-3",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: expv1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
						Kind: expv1alpha1.ClusterStoreKind,
					},
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
			},
			Status: expv1alpha1.CustomBuilderStatus{
				BuilderStatus: v1alpha1.BuilderStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
					Stack: v1alpha1.BuildStack{
						RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9fasdfa847098234",
						ID:       "io.buildpacks.stacks.bionic",
					},
					LatestImage: "some-registry.com/test-builder-3:tag",
				},
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return custombuilder.NewListCommand(clientSetProvider)
	}

	when("listing clusterbuilder", func() {
		when("namespace is provided", func() {
			when("there are builders in the namespace", func() {
				it("lists the builders", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							otherNamespacedCustomBuilder1,
							otherNamespacedCustomBuilder2,
							otherNamespacedCustomBuilder3,
						},
						Args:           []string{"-n", "test-namespace"},
						ExpectedOutput: expectedOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("there are no builders in the namespace", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						ExpectErr:      true,
						ExpectedOutput: "Error: no builders found\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})

		when("namespace is not provided", func() {
			when("there are builders in the default namespace", func() {
				it("lists the builders", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							defaultNamespacedCustomBuilder1,
							defaultNamespacedCustomBuilder2,
							defaultNamespacedCustomBuilder3,
						},
						ExpectedOutput: expectedOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("there are no builders in the default namespace", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						ExpectErr:      true,
						ExpectedOutput: "Error: no builders found\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})
	})
}

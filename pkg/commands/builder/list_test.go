// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/builder"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
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
		defaultNamespacedBuilder1 = &v1alpha2.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.BuilderKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-1",
				Namespace: defaultNamespace,
			},
			Spec: v1alpha2.NamespacedBuilderSpec{
				BuilderSpec: v1alpha2.BuilderSpec{
					Tag: "some-registry.com/test-builder-1",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: v1alpha2.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
						Kind: v1alpha2.ClusterStoreKind,
					},
					Order: []v1alpha2.OrderEntry{
						{
							Group: []v1alpha2.BuildpackRef{
								{
									BuildpackInfo: v1alpha2.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []v1alpha2.BuildpackRef{
								{
									BuildpackInfo: v1alpha2.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
						},
					},
				},
			},
			Status: v1alpha2.BuilderStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
				Stack: v1alpha2.BuildStack{
					RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9203847098234",
					ID:       "io.buildpacks.stacks.centos",
				},
				LatestImage: "some-registry.com/test-builder-1:tag",
			},
		}
		defaultNamespacedBuilder2 = &v1alpha2.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.BuilderKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-2",
				Namespace: defaultNamespace,
			},
			Spec: v1alpha2.NamespacedBuilderSpec{
				BuilderSpec: v1alpha2.BuilderSpec{
					Tag: "some-registry.com/test-builder-2",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: v1alpha2.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
						Kind: v1alpha2.ClusterStoreKind,
					},
					Order: []v1alpha2.OrderEntry{
						{
							Group: []v1alpha2.BuildpackRef{
								{
									BuildpackInfo: v1alpha2.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []v1alpha2.BuildpackRef{
								{
									BuildpackInfo: v1alpha2.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
						},
					},
				},
			},
			Status: v1alpha2.BuilderStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
		}
		defaultNamespacedBuilder3 = &v1alpha2.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.BuilderKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-3",
				Namespace: defaultNamespace,
			},
			Spec: v1alpha2.NamespacedBuilderSpec{
				BuilderSpec: v1alpha2.BuilderSpec{
					Tag: "some-registry.com/test-builder-3",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: v1alpha2.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
						Kind: v1alpha2.ClusterStoreKind,
					},
					Order: []v1alpha2.OrderEntry{
						{
							Group: []v1alpha2.BuildpackRef{
								{
									BuildpackInfo: v1alpha2.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []v1alpha2.BuildpackRef{
								{
									BuildpackInfo: v1alpha2.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
						},
					},
				},
			},
			Status: v1alpha2.BuilderStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
				Stack: v1alpha2.BuildStack{
					RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9fasdfa847098234",
					ID:       "io.buildpacks.stacks.bionic",
				},
				LatestImage: "some-registry.com/test-builder-3:tag",
			},
		}

		otherNamespacedBuilder1 = &v1alpha2.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.BuilderKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-1",
				Namespace: "test-namespace",
			},
			Spec: v1alpha2.NamespacedBuilderSpec{
				BuilderSpec: v1alpha2.BuilderSpec{
					Tag: "some-registry.com/test-builder-1",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: v1alpha2.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
						Kind: v1alpha2.ClusterStoreKind,
					},
					Order: []v1alpha2.OrderEntry{
						{
							Group: []v1alpha2.BuildpackRef{
								{
									BuildpackInfo: v1alpha2.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []v1alpha2.BuildpackRef{
								{
									BuildpackInfo: v1alpha2.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
						},
					},
				},
			},
			Status: v1alpha2.BuilderStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
				Stack: v1alpha2.BuildStack{
					RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9203847098234",
					ID:       "io.buildpacks.stacks.centos",
				},
				LatestImage: "some-registry.com/test-builder-1:tag",
			},
		}
		otherNamespacedBuilder2 = &v1alpha2.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.BuilderKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-2",
				Namespace: "test-namespace",
			},
			Spec: v1alpha2.NamespacedBuilderSpec{
				BuilderSpec: v1alpha2.BuilderSpec{
					Tag: "some-registry.com/test-builder-2",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: v1alpha2.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
						Kind: v1alpha2.ClusterStoreKind,
					},
					Order: []v1alpha2.OrderEntry{
						{
							Group: []v1alpha2.BuildpackRef{
								{
									BuildpackInfo: v1alpha2.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []v1alpha2.BuildpackRef{
								{
									BuildpackInfo: v1alpha2.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
						},
					},
				},
			},
			Status: v1alpha2.BuilderStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
		}
		otherNamespacedBuilder3 = &v1alpha2.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.BuilderKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-3",
				Namespace: "test-namespace",
			},
			Spec: v1alpha2.NamespacedBuilderSpec{
				BuilderSpec: v1alpha2.BuilderSpec{
					Tag: "some-registry.com/test-builder-3",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: v1alpha2.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
						Kind: v1alpha2.ClusterStoreKind,
					},
					Order: []v1alpha2.OrderEntry{
						{
							Group: []v1alpha2.BuildpackRef{
								{
									BuildpackInfo: v1alpha2.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []v1alpha2.BuildpackRef{
								{
									BuildpackInfo: v1alpha2.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
						},
					},
				},
			},
			Status: v1alpha2.BuilderStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
				Stack: v1alpha2.BuildStack{
					RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9fasdfa847098234",
					ID:       "io.buildpacks.stacks.bionic",
				},
				LatestImage: "some-registry.com/test-builder-3:tag",
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return builder.NewListCommand(clientSetProvider)
	}

	when("listing clusterbuilder", func() {
		when("namespace is provided", func() {
			when("there are builders in the namespace", func() {
				it("lists the builders", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							otherNamespacedBuilder1,
							otherNamespacedBuilder2,
							otherNamespacedBuilder3,
						},
						Args:           []string{"-n", "test-namespace"},
						ExpectedOutput: expectedOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("there are no builders in the namespace", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: no builders found\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})

		when("namespace is not provided", func() {
			when("there are builders in the default namespace", func() {
				it("lists the builders", func() {
					testhelpers.CommandTest{
						Objects: []runtime.Object{
							defaultNamespacedBuilder1,
							defaultNamespacedBuilder2,
							defaultNamespacedBuilder3,
						},
						ExpectedOutput: expectedOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("there are no builders in the default namespace", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: no builders found\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})
	})
}

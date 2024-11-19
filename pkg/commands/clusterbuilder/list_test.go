// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder_test

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

	"github.com/buildpacks-community/kpack-cli/pkg/commands/clusterbuilder"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
)

func TestClusterBuilderListCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuilderListCommand", testClusterBuilderListCommand)
}

func testClusterBuilderListCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		expectedOutput = `NAME              READY    STACK                          IMAGE
test-builder-1    true     io.buildpacks.stacks.centos    some-registry.com/test-builder-1:tag
test-builder-2    false                                   
test-builder-3    true     io.buildpacks.stacks.bionic    some-registry.com/test-builder-3:tag

`
	)

	var (
		clusterBuilder1 = &v1alpha2.ClusterBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.ClusterBuilderKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-builder-1",
			},
			Spec: v1alpha2.ClusterBuilderSpec{
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
					Order: []v1alpha2.BuilderOrderEntry{
						{
							Group: []v1alpha2.BuilderBuildpackRef{
								{
									BuildpackRef: corev1alpha1.BuildpackRef{
										BuildpackInfo: corev1alpha1.BuildpackInfo{
											Id: "org.cloudfoundry.nodejs",
										},
									},
								},
							},
						},
						{
							Group: []v1alpha2.BuilderBuildpackRef{
								{
									BuildpackRef: corev1alpha1.BuildpackRef{
										BuildpackInfo: corev1alpha1.BuildpackInfo{
											Id: "org.cloudfoundry.go",
										},
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
			Status: v1alpha2.BuilderStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
				Stack: corev1alpha1.BuildStack{
					RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9203847098234",
					ID:       "io.buildpacks.stacks.centos",
				},
				LatestImage: "some-registry.com/test-builder-1:tag",
			},
		}
		clusterBuilder2 = &v1alpha2.ClusterBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.ClusterBuilderKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-builder-2",
			},
			Spec: v1alpha2.ClusterBuilderSpec{
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
					Order: []v1alpha2.BuilderOrderEntry{
						{
							Group: []v1alpha2.BuilderBuildpackRef{
								{
									BuildpackRef: corev1alpha1.BuildpackRef{
										BuildpackInfo: corev1alpha1.BuildpackInfo{
											Id: "org.cloudfoundry.nodejs",
										},
									},
								},
							},
						},
						{
							Group: []v1alpha2.BuilderBuildpackRef{
								{
									BuildpackRef: corev1alpha1.BuildpackRef{
										BuildpackInfo: corev1alpha1.BuildpackInfo{
											Id: "org.cloudfoundry.go",
										},
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
		clusterBuilder3 = &v1alpha2.ClusterBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.ClusterBuilderKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-builder-3",
			},
			Spec: v1alpha2.ClusterBuilderSpec{
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
					Order: []v1alpha2.BuilderOrderEntry{
						{
							Group: []v1alpha2.BuilderBuildpackRef{
								{
									BuildpackRef: corev1alpha1.BuildpackRef{
										BuildpackInfo: corev1alpha1.BuildpackInfo{
											Id: "org.cloudfoundry.nodejs",
										},
									},
								},
							},
						},
						{
							Group: []v1alpha2.BuilderBuildpackRef{
								{
									BuildpackRef: corev1alpha1.BuildpackRef{
										BuildpackInfo: corev1alpha1.BuildpackInfo{
											Id: "org.cloudfoundry.go",
										},
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
			Status: v1alpha2.BuilderStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
				Stack: corev1alpha1.BuildStack{
					RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9fasdfa847098234",
					ID:       "io.buildpacks.stacks.bionic",
				},
				LatestImage: "some-registry.com/test-builder-3:tag",
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterbuilder.NewListCommand(clientSetProvider)
	}

	when("listing clusterbuilder", func() {
		when("there are clusterbuilders", func() {
			it("lists the builders", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						clusterBuilder1,
						clusterBuilder2,
						clusterBuilder3,
					},
					ExpectedOutput: expectedOutput,
				}.TestKpack(t, cmdFunc)
			})
		})

		when("there are no clusterbuilders", func() {
			it("prints an appropriate message", func() {
				testhelpers.CommandTest{
					ExpectErr:           true,
					ExpectedErrorOutput: "Error: no clusterbuilders found\n",
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}

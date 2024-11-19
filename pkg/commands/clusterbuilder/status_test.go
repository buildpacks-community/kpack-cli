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

func TestClusterBuilderStatusCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuilderStatusCommand", testClusterBuilderStatusCommand)
}

func testClusterBuilderStatusCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		expectedReadyOutputUsingSpecOrder = `Status:       Ready
Image:        some-registry.com/test-builder-1:tag
Stack ID:     io.buildpacks.stacks.centos
Run Image:    gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9203847098234

Stack Ref:     
  Name:       test-stack
  Kind:       ClusterStack
Store Ref:     
  Name:       test-store
  Kind:       ClusterStore

BUILDPACK ID               VERSION    HOMEPAGE
org.cloudfoundry.nodejs    v0.2.1     https://github.com/paketo-buildpacks/nodejs
org.cloudfoundry.go        v0.0.3     https://github.com/paketo-buildpacks/go


CLUSTERBUILDPACK NAME         CLUSTERBUILDPACK KIND


sample-cluster-buildpack    ClusterBuildpack

DETECTION ORDER              
Group #1                     
  org.cloudfoundry.nodejs    
Group #2                     
  org.cloudfoundry.go        

`
		expectedReadyOutputUsingStatusOrder = `Status:       Ready
Image:        some-registry.com/test-builder-1:tag
Stack ID:     io.buildpacks.stacks.centos
Run Image:    gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9203847098234

Stack Ref:     
  Name:       test-stack
  Kind:       ClusterStack
Store Ref:     
  Name:       test-store
  Kind:       ClusterStore

BUILDPACK ID               VERSION    HOMEPAGE
org.cloudfoundry.nodejs    v0.2.1     https://github.com/paketo-buildpacks/nodejs
org.cloudfoundry.go        v0.0.3     https://github.com/paketo-buildpacks/go


CLUSTERBUILDPACK NAME         CLUSTERBUILDPACK KIND


sample-cluster-buildpack    ClusterBuildpack

DETECTION ORDER                    
Group #1                           
  org.cloudfoundry.nodejs@0.2.1    
Group #2                           
  org.cloudfoundry.go@0.0.3        

`
		expectedNotReadyOutput = `Status:    Not Ready
Reason:    this builder is not ready for the purpose of a test

`
		expectedUnkownOutput = `Status:    Unknown

`
	)

	var (
		readyClusterBuilder = &v1alpha2.ClusterBuilder{
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
						{
							Group: []v1alpha2.BuilderBuildpackRef{
								{
									ObjectReference: corev1.ObjectReference{
										Name: "sample-cluster-buildpack",
										Kind: "ClusterBuildpack",
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
				BuilderMetadata: corev1alpha1.BuildpackMetadataList{
					{
						Id:       "org.cloudfoundry.nodejs",
						Version:  "v0.2.1",
						Homepage: "https://github.com/paketo-buildpacks/nodejs",
					},
					{
						Id:       "org.cloudfoundry.go",
						Version:  "v0.0.3",
						Homepage: "https://github.com/paketo-buildpacks/go",
					},
				},
				Stack: corev1alpha1.BuildStack{
					RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9203847098234",
					ID:       "io.buildpacks.stacks.centos",
				},
				LatestImage: "some-registry.com/test-builder-1:tag",
			},
		}
		notReadyClusterBuilder = &v1alpha2.ClusterBuilder{
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
							Type:    corev1alpha1.ConditionReady,
							Status:  corev1.ConditionFalse,
							Message: "this builder is not ready for the purpose of a test",
						},
					},
				},
			},
		}
		unknownClusterBuilder = &v1alpha2.ClusterBuilder{
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
			Status: v1alpha2.BuilderStatus{},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterbuilder.NewStatusCommand(clientSetProvider)
	}

	when("getting clusterbuilder status", func() {
		when("the clusterbuilder exists", func() {
			when("the builder is ready", func() {
				when("the order is in the builder status", func() {
					it("shows the build status using status.order", func() {
						readyClusterBuilder.Status.Order = []corev1alpha1.OrderEntry{
							{
								Group: []corev1alpha1.BuildpackRef{
									{
										BuildpackInfo: corev1alpha1.BuildpackInfo{
											Id:      "org.cloudfoundry.nodejs",
											Version: "0.2.1",
										},
									},
								},
							},
							{
								Group: []corev1alpha1.BuildpackRef{
									{
										BuildpackInfo: corev1alpha1.BuildpackInfo{
											Id:      "org.cloudfoundry.go",
											Version: "0.0.3",
										},
									},
								},
							},
						}

						testhelpers.CommandTest{
							Objects:        []runtime.Object{readyClusterBuilder},
							Args:           []string{"test-builder-1"},
							ExpectedOutput: expectedReadyOutputUsingStatusOrder,
						}.TestKpack(t, cmdFunc)
					})

				})
			})

			when("the builder is not ready", func() {
				it("shows the build status of not ready builder", func() {
					testhelpers.CommandTest{
						Objects:        []runtime.Object{notReadyClusterBuilder},
						Args:           []string{"test-builder-2"},
						ExpectedOutput: expectedNotReadyOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("the builder is unknown", func() {
				it("shows the build status of unknown builder", func() {
					testhelpers.CommandTest{
						Objects:        []runtime.Object{unknownClusterBuilder},
						Args:           []string{"test-builder-3"},
						ExpectedOutput: expectedUnkownOutput,
					}.TestKpack(t, cmdFunc)
				})
			})
		})

		when("the clusterbuidler does not exist", func() {
			it("prints an appropriate message", func() {
				testhelpers.CommandTest{
					Args:                []string{"non-existant-builder"},
					ExpectErr:           true,
					ExpectedErrorOutput: "Error: clusterbuilders.kpack.io \"non-existant-builder\" not found\n",
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}

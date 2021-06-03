// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
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

func TestClusterBuilderStatusCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuilderStatusCommand", testClusterBuilderStatusCommand)
}

func testClusterBuilderStatusCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		defaultNamespace                  = "some-default-namespace"
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
		readyDefaultBuilder = &v1alpha1.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterBuilderKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-1",
				Namespace: defaultNamespace,
			},
			Spec: v1alpha1.NamespacedBuilderSpec{
				BuilderSpec: v1alpha1.BuilderSpec{
					Tag: "some-registry.com/test-builder-1",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: v1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
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
			},
			Status: v1alpha1.BuilderStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
				BuilderMetadata: v1alpha1.BuildpackMetadataList{
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
				Stack: v1alpha1.BuildStack{
					RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9203847098234",
					ID:       "io.buildpacks.stacks.centos",
				},
				LatestImage: "some-registry.com/test-builder-1:tag",
			},
		}
		notReadyDefaultBuilder = &v1alpha1.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterBuilderKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-2",
				Namespace: defaultNamespace,
			},
			Spec: v1alpha1.NamespacedBuilderSpec{
				BuilderSpec: v1alpha1.BuilderSpec{
					Tag: "some-registry.com/test-builder-2",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: v1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
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
			},
			Status: v1alpha1.BuilderStatus{
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
		unknownDefaultBuilder = &v1alpha1.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterBuilderKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-3",
				Namespace: defaultNamespace,
			},
			Spec: v1alpha1.NamespacedBuilderSpec{
				BuilderSpec: v1alpha1.BuilderSpec{
					Tag: "some-registry.com/test-builder-3",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: v1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
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
			},
			Status: v1alpha1.BuilderStatus{},
		}

		readyNamespaceBuilder = &v1alpha1.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterBuilderKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-1",
				Namespace: "test-namespace",
			},
			Spec: v1alpha1.NamespacedBuilderSpec{
				BuilderSpec: v1alpha1.BuilderSpec{
					Tag: "some-registry.com/test-builder-1",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: v1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
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
			},
			Status: v1alpha1.BuilderStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
				BuilderMetadata: v1alpha1.BuildpackMetadataList{
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
				Stack: v1alpha1.BuildStack{
					RunImage: "gcr.io/paketo-buildpacks/run@sha256:iweuryaksdjhf9203847098234",
					ID:       "io.buildpacks.stacks.centos",
				},
				LatestImage: "some-registry.com/test-builder-1:tag",
			},
		}
		notReadyNamespaceBuilder = &v1alpha1.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterBuilderKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-2",
				Namespace: "test-namespace",
			},
			Spec: v1alpha1.NamespacedBuilderSpec{
				BuilderSpec: v1alpha1.BuilderSpec{
					Tag: "some-registry.com/test-builder-2",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: v1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
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
			},
			Status: v1alpha1.BuilderStatus{
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
		unknownNamespaceBuilder = &v1alpha1.Builder{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterBuilderKind,
				APIVersion: "kpack.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-builder-3",
				Namespace: "test-namespace",
			},
			Spec: v1alpha1.NamespacedBuilderSpec{
				BuilderSpec: v1alpha1.BuilderSpec{
					Tag: "some-registry.com/test-builder-3",
					Stack: corev1.ObjectReference{
						Name: "test-stack",
						Kind: v1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "test-store",
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
			},
			Status: v1alpha1.BuilderStatus{},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return builder.NewStatusCommand(clientSetProvider)
	}

	when("getting builder status", func() {
		when("in the default namespace", func() {

			when("the builder exists", func() {
				when("the builder is ready", func() {
					when("the order is not in the builder status", func() {
						it("shows the build status falling back to spec.order", func() {
							testhelpers.CommandTest{
								Objects:        []runtime.Object{readyDefaultBuilder},
								Args:           []string{"test-builder-1"},
								ExpectedOutput: expectedReadyOutputUsingSpecOrder,
							}.TestKpack(t, cmdFunc)
						})

					})
					when("the order is in the builder status", func() {
						it("shows the build status using status.order", func() {
							readyDefaultBuilder.Status.Order = []v1alpha1.OrderEntry{
								{
									Group: []v1alpha1.BuildpackRef{
										{
											BuildpackInfo: v1alpha1.BuildpackInfo{
												Id:      "org.cloudfoundry.nodejs",
												Version: "0.2.1",
											},
										},
									},
								},
								{
									Group: []v1alpha1.BuildpackRef{
										{
											BuildpackInfo: v1alpha1.BuildpackInfo{
												Id:      "org.cloudfoundry.go",
												Version: "0.0.3",
											},
										},
									},
								},
							}

							testhelpers.CommandTest{
								Objects:        []runtime.Object{readyDefaultBuilder},
								Args:           []string{"test-builder-1"},
								ExpectedOutput: expectedReadyOutputUsingStatusOrder,
							}.TestKpack(t, cmdFunc)
						})

					})
				})

				when("the builder is not ready", func() {
					it("shows the build status of not ready builder", func() {
						testhelpers.CommandTest{
							Objects:        []runtime.Object{notReadyDefaultBuilder},
							Args:           []string{"test-builder-2"},
							ExpectedOutput: expectedNotReadyOutput,
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the builder is unknown", func() {
					it("shows the build status of unknown builder", func() {
						testhelpers.CommandTest{
							Objects:        []runtime.Object{unknownDefaultBuilder},
							Args:           []string{"test-builder-3"},
							ExpectedOutput: expectedUnkownOutput,
						}.TestKpack(t, cmdFunc)
					})
				})
			})

			when("the builder does not exist", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						Args:           []string{"non-existant-builder"},
						ExpectErr:      true,
						ExpectedOutput: "Error: builders.kpack.io \"non-existant-builder\" not found\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})

		when("in the specified namespace", func() {

			when("the builder exists", func() {
				when("the builder is ready", func() {
					it("shows the build status", func() {
						testhelpers.CommandTest{
							Objects:        []runtime.Object{readyNamespaceBuilder},
							Args:           []string{"test-builder-1", "-n", "test-namespace"},
							ExpectedOutput: expectedReadyOutputUsingSpecOrder,
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the builder is not ready", func() {
					it("shows the build status of not ready builder", func() {
						testhelpers.CommandTest{
							Objects:        []runtime.Object{notReadyNamespaceBuilder},
							Args:           []string{"test-builder-2", "-n", "test-namespace"},
							ExpectedOutput: expectedNotReadyOutput,
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the builder is unknown", func() {
					it("shows the build status of unknown builder", func() {
						testhelpers.CommandTest{
							Objects:        []runtime.Object{unknownNamespaceBuilder},
							Args:           []string{"test-builder-3", "-n", "test-namespace"},
							ExpectedOutput: expectedUnkownOutput,
						}.TestKpack(t, cmdFunc)
					})
				})
			})

			when("the builder does not exist", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						Args:           []string{"non-existant-builder", "-n", "test-namespace"},
						ExpectErr:      true,
						ExpectedOutput: "Error: builders.kpack.io \"non-existant-builder\" not found\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})
	})
}

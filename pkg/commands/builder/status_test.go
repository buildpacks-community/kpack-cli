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
	"github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterbuildpack"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands/buildpack"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestBuilderStatusCommand(t *testing.T) {
	spec.Run(t, "TestBuilderStatusCommand", testBuilderStatusCommand)
}

func testBuilderStatusCommand(t *testing.T, when spec.G, it spec.S) {
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


BUILDPACKNAME         BUILDPACKKIND


sample-buildpack    Buildpack

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


BUILDPACKNAME         BUILDPACKKIND


sample-buildpack    Buildpack

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
		readyDefaultBuilder = &v1alpha2.Builder{
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
										Name: "sample-buildpack",
										Kind: "Buildpack",
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
		notReadyDefaultBuilder = &v1alpha2.Builder{
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
		unknownDefaultBuilder = &v1alpha2.Builder{
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
			},
			Status: v1alpha2.BuilderStatus{},
		}

		readyNamespaceBuilder = &v1alpha2.Builder{
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
										Name: "sample-buildpack",
										Kind: "Buildpack",
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
				Order: []corev1alpha1.OrderEntry{
					{
						Group: []corev1alpha1.BuildpackRef{
							{
								BuildpackInfo: corev1alpha1.BuildpackInfo{
									Id: "org.cloudfoundry.nodejs",
								},
							},
						},
					},
					{
						Group: []corev1alpha1.BuildpackRef{
							{
								BuildpackInfo: corev1alpha1.BuildpackInfo{
									Id: "org.cloudfoundry.go",
								},
							},
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
		notReadyNamespaceBuilder = &v1alpha2.Builder{
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
		unknownNamespaceBuilder = &v1alpha2.Builder{
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
			},
			Status: v1alpha2.BuilderStatus{},
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
					when("the order is in the builder status", func() {
						it("shows the build status using status.order", func() {
							readyDefaultBuilder.Status.Order = []corev1alpha1.OrderEntry{
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
						Args:                []string{"non-existant-builder"},
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: builders.kpack.io \"non-existant-builder\" not found\n",
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
						Args:                []string{"non-existant-builder", "-n", "test-namespace"},
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: builders.kpack.io \"non-existant-builder\" not found\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})
	})
}

func TestClusterBuildpackStatusCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuildpackStatusCommand", testClusterBuildpackStatusCommand)
}

func testClusterBuildpackStatusCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		defaultNamespace    = "some-default-namespace"
		expectedReadyOutput = `Status:    Ready
Source:    some-registry.com/test-buildpack-1

BUILDPACK ID               VERSION    HOMEPAGE
org.cloudfoundry.nodejs    0.2.1      

`
	)

	var (
		readyDefaultClusterBuildpack = &v1alpha2.ClusterBuildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.ClusterBuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-buildpack-1",
			},
			Spec: v1alpha2.ClusterBuildpackSpec{
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-1",
				},
			},
			Status: v1alpha2.ClusterBuildpackStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
				Buildpacks: []corev1alpha1.BuildpackStatus{
					{
						BuildpackInfo: corev1alpha1.BuildpackInfo{
							Id:      "org.cloudfoundry.nodejs",
							Version: "0.2.1",
						},
					},
				},
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterbuildpack.NewStatusCommand(clientSetProvider)
	}

	when("getting buildpack status", func() {
		when("the buildpack exists", func() {
			when("the buildpack is ready", func() {
				it("shows the build status using status.order", func() {
					testhelpers.CommandTest{
						Objects:        []runtime.Object{readyDefaultClusterBuildpack},
						Args:           []string{"test-buildpack-1"},
						ExpectedOutput: expectedReadyOutput,
					}.TestKpack(t, cmdFunc)
				})
			})
		})
	})
}

func TestBuildpackStatusCommand(t *testing.T) {
	spec.Run(t, "TestBuildpackStatusCommand", testBuildpackStatusCommand)
}

func testBuildpackStatusCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		defaultNamespace    = "some-default-namespace"
		expectedReadyOutput = `Status:    Ready
Source:    some-registry.com/test-buildpack-1

BUILDPACK ID               VERSION    HOMEPAGE
org.cloudfoundry.nodejs    0.2.1      

`
	)

	var (
		readyDefaultBuildpack = &v1alpha2.Buildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha2.BuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-buildpack-1",
				Namespace: defaultNamespace,
			},
			Spec: v1alpha2.BuildpackSpec{
				ServiceAccountName: "default",
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-1",
				},
			},
			Status: v1alpha2.BuildpackStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
				Buildpacks: []corev1alpha1.BuildpackStatus{
					{
						BuildpackInfo: corev1alpha1.BuildpackInfo{
							Id:      "org.cloudfoundry.nodejs",
							Version: "0.2.1",
						},
					},
				},
			},
		}
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return buildpack.NewStatusCommand(clientSetProvider)
	}

	when("getting buildpack status", func() {
		when("in the default namespace", func() {
			when("the buildpack exists", func() {
				when("the buildpack is ready", func() {
					it("shows the build status using status.order", func() {
						testhelpers.CommandTest{
							Objects:        []runtime.Object{readyDefaultBuildpack},
							Args:           []string{"test-buildpack-1"},
							ExpectedOutput: expectedReadyOutput,
						}.TestKpack(t, cmdFunc)
					})
				})
		  })
	})
  })
}
// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/image"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestImageListCommand(t *testing.T) {
	spec.Run(t, "TestImageListCommand", testImageListCommand)
}

func testImageListCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return image.NewListCommand(clientSetProvider)
	}

	when("a namespace is provided", func() {
		when("the namespaces has images", func() {
			it("returns a table of image details", func() {
				image1 := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image-1",
						Namespace: "test-namespace",
					},
					Status: v1alpha2.ImageStatus{
						LatestBuildReason: "COMMIT",
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionFalse,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-1@sha256:abcdef123",
					},
				}
				image2 := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image-2",
						Namespace: "test-namespace",
					},
					Status: v1alpha2.ImageStatus{
						LatestBuildReason: "COMMIT",
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionUnknown,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-2@sha256:abcdef123",
					},
				}
				image3 := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image-3",
						Namespace: "test-namespace",
					},
					Spec: v1alpha2.ImageSpec{},
					Status: v1alpha2.ImageStatus{
						LatestBuildReason: "COMMIT",
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionTrue,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-3@sha256:abcdef123",
					},
				}
				notInNamespaceImage := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image-4",
						Namespace: defaultNamespace,
					},
					Status: v1alpha2.ImageStatus{
						LatestBuildReason: "COMMIT",
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionFalse,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-4@sha256:abcdef123",
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						image1,
						image2,
						image3,
						notInNamespaceImage,
					},
					Args: []string{"-n", "test-namespace"},
					ExpectedOutput: `NAME            READY      LATEST REASON    LATEST IMAGE                                      NAMESPACE
test-image-1    False      COMMIT           test-registry.io/test-image-1@sha256:abcdef123    test-namespace
test-image-2    Unknown    COMMIT           test-registry.io/test-image-2@sha256:abcdef123    test-namespace
test-image-3    True       COMMIT           test-registry.io/test-image-3@sha256:abcdef123    test-namespace

`,
				}.TestKpack(t, cmdFunc)
			})

			when("the namespace has no images", func() {
				it("returns a message that the namespace has no images", func() {
					testhelpers.CommandTest{
						Args:                []string{"-n", "test-namespace"},
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: no images found\n",
					}.TestKpack(t, cmdFunc)

				})
			})
		})
	})

	when("a namespace is not provided", func() {
		when("the namespaces has images", func() {
			it("returns a table of image details", func() {
				image1 := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image-1",
						Namespace: defaultNamespace,
					},
					Status: v1alpha2.ImageStatus{
						LatestBuildReason: "COMMIT",
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionFalse,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-1@sha256:abcdef123",
					},
				}

				image2 := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image-2",
						Namespace: defaultNamespace,
					},
					Status: v1alpha2.ImageStatus{
						LatestBuildReason: "COMMIT",
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionUnknown,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-2@sha256:abcdef123",
					},
				}

				image3 := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image-3",
						Namespace: defaultNamespace,
					},
					Spec: v1alpha2.ImageSpec{},
					Status: v1alpha2.ImageStatus{
						LatestBuildReason: "COMMIT",
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionTrue,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-3@sha256:abcdef123",
					},
				}

				notDefaultNamespaceImage := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image-4",
						Namespace: "not-default-namespace",
					},
					Status: v1alpha2.ImageStatus{
						LatestBuildReason: "COMMIT",
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionFalse,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-4@sha256:abcdef123",
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						image1,
						image2,
						image3,
						notDefaultNamespaceImage,
					},
					ExpectedOutput: `NAME            READY      LATEST REASON    LATEST IMAGE                                      NAMESPACE
test-image-1    False      COMMIT           test-registry.io/test-image-1@sha256:abcdef123    some-default-namespace
test-image-2    Unknown    COMMIT           test-registry.io/test-image-2@sha256:abcdef123    some-default-namespace
test-image-3    True       COMMIT           test-registry.io/test-image-3@sha256:abcdef123    some-default-namespace

`,
				}.TestKpack(t, cmdFunc)
			})

			when("the namespace has no images", func() {
				it("returns a message that the namespace has no images", func() {
					testhelpers.CommandTest{
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: no images found\n",
					}.TestKpack(t, cmdFunc)

				})
			})
		})
	})

	when("all namespaces is specified", func() {
		when("the namespaces has images", func() {
			it("returns a table of image details", func() {
				image1 := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image-1",
						Namespace: defaultNamespace,
					},
					Status: v1alpha2.ImageStatus{
						LatestBuildReason: "COMMIT",
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionFalse,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-1@sha256:abcdef123",
					},
				}

				image2 := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image-2",
						Namespace: defaultNamespace,
					},
					Status: v1alpha2.ImageStatus{
						LatestBuildReason: "COMMIT",
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionUnknown,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-2@sha256:abcdef123",
					},
				}

				image3 := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image-3",
						Namespace: defaultNamespace,
					},
					Spec: v1alpha2.ImageSpec{},
					Status: v1alpha2.ImageStatus{
						LatestBuildReason: "COMMIT",
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionTrue,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-3@sha256:abcdef123",
					},
				}

				notDefaultNamespaceImage := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-image-4",
						Namespace: "not-default-namespace",
					},
					Status: v1alpha2.ImageStatus{
						LatestBuildReason: "COMMIT",
						Status: corev1alpha1.Status{
							Conditions: []corev1alpha1.Condition{
								{
									Type:   corev1alpha1.ConditionReady,
									Status: corev1.ConditionFalse,
								},
							},
						},
						LatestImage: "test-registry.io/test-image-4@sha256:abcdef123",
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						image1,
						image2,
						image3,
						notDefaultNamespaceImage,
					},
					Args: []string{"-A"},
					ExpectedOutput: `NAME            READY      LATEST REASON    LATEST IMAGE                                      NAMESPACE
test-image-1    False      COMMIT           test-registry.io/test-image-1@sha256:abcdef123    some-default-namespace
test-image-2    Unknown    COMMIT           test-registry.io/test-image-2@sha256:abcdef123    some-default-namespace
test-image-3    True       COMMIT           test-registry.io/test-image-3@sha256:abcdef123    some-default-namespace
test-image-4    False      COMMIT           test-registry.io/test-image-4@sha256:abcdef123    not-default-namespace

`,
				}.TestKpack(t, cmdFunc)
			})

			when("the namespace has no images", func() {
				it("returns a message that the namespace has no images", func() {
					testhelpers.CommandTest{
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: no images found\n",
					}.TestKpack(t, cmdFunc)

				})
			})
		})
	})
}

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

func TestImageStatusCommand(t *testing.T) {
	spec.Run(t, "TestImageStatusCommand", testImageStatusCommand)
}

func testImageStatusCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		defaultNamespace = "some-default-namespace"
		namespace        = "test-namespace"
		imageName        = "test-image"
	)

	testBuilds := testhelpers.MakeTestBuilds(imageName, defaultNamespace)
	testNamespacedBuilds := testhelpers.MakeTestBuilds(imageName, namespace)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return image.NewStatusCommand(clientSetProvider)
	}

	when("a namespace is provided", func() {
		when("the namespaces has images", func() {
			it("returns a table of image details for git source", func() {
				image := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      imageName,
						Namespace: namespace,
					},
					Spec: v1alpha2.ImageSpec{
						Builder: corev1.ObjectReference{
							Kind: "ClusterBuilder",
							Name: "some-cluster-builder",
						},
						Source: corev1alpha1.SourceConfig{
							Git: &corev1alpha1.Git{
								URL:      "some-git-url",
								Revision: "some-git-revision",
							},
						},
					},
					Status: v1alpha2.ImageStatus{
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
				testNamespacedBuilds[0].Spec.Source = corev1alpha1.SourceConfig{
					Git: &corev1alpha1.Git{
						Revision: "successful-build-git-revision",
					},
				}
				testNamespacedBuilds[2].Spec.Source = corev1alpha1.SourceConfig{
					Git: &corev1alpha1.Git{
						Revision: "failed-build-git-revision",
					},
				}

				const expectedOutput = `Status:         Not Ready
Message:        --
LatestImage:    test-registry.io/test-image-1@sha256:abcdef123

Source
Type:        GitUrl
Url:         some-git-url
Revision:    some-git-revision

Builder Ref
Name:    some-cluster-builder
Kind:    ClusterBuilder

Last Successful Build
Id:              1
Build Reason:    CONFIG
Git Revision:    successful-build-git-revision

BUILDPACK ID    BUILDPACK VERSION    HOMEPAGE
bp-id-1         bp-version-1         mysupercoolsite.com
bp-id-2         bp-version-2         mysupercoolsite2.com

Last Failed Build
Id:              2
Build Reason:    COMMIT,BUILDPACK
Git Revision:    failed-build-git-revision

`

				testhelpers.CommandTest{
					Objects:        append([]runtime.Object{image}, testhelpers.BuildsToRuntimeObjs(testNamespacedBuilds)...),
					Args:           []string{imageName, "-n", namespace},
					ExpectedOutput: expectedOutput,
				}.TestKpack(t, cmdFunc)
			})

			it("returns a table of image details for blob source", func() {
				image := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      imageName,
						Namespace: namespace,
					},
					Spec: v1alpha2.ImageSpec{
						Builder: corev1.ObjectReference{
							Kind: "ClusterBuilder",
							Name: "some-cluster-builder",
						},
						Source: corev1alpha1.SourceConfig{
							Blob: &corev1alpha1.Blob{
								URL: "some-blob-url",
							},
						},
					},
					Status: v1alpha2.ImageStatus{
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

				const expectedOutput = `Status:         Not Ready
Message:        --
LatestImage:    test-registry.io/test-image-1@sha256:abcdef123

Source
Type:    Blob
Url:     some-blob-url

Builder Ref
Name:    some-cluster-builder
Kind:    ClusterBuilder

Last Successful Build
Id:              1
Build Reason:    CONFIG

BUILDPACK ID    BUILDPACK VERSION    HOMEPAGE
bp-id-1         bp-version-1         mysupercoolsite.com
bp-id-2         bp-version-2         mysupercoolsite2.com

Last Failed Build
Id:              2
Build Reason:    COMMIT,BUILDPACK

`

				testhelpers.CommandTest{
					Objects:        append([]runtime.Object{image}, testhelpers.BuildsToRuntimeObjs(testNamespacedBuilds)...),
					Args:           []string{imageName, "-n", namespace},
					ExpectedOutput: expectedOutput,
				}.TestKpack(t, cmdFunc)
			})

			it("returns a table of image details for local source", func() {
				image := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      imageName,
						Namespace: namespace,
					},
					Spec: v1alpha2.ImageSpec{
						Builder: corev1.ObjectReference{
							Kind: "ClusterBuilder",
							Name: "some-cluster-builder",
						},
						Source: corev1alpha1.SourceConfig{
							Registry: &corev1alpha1.Registry{},
						},
					},
					Status: v1alpha2.ImageStatus{
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

				const expectedOutput = `Status:         Not Ready
Message:        --
LatestImage:    test-registry.io/test-image-1@sha256:abcdef123

Source
Type:    Local Source

Builder Ref
Name:    some-cluster-builder
Kind:    ClusterBuilder

Last Successful Build
Id:              1
Build Reason:    CONFIG

BUILDPACK ID    BUILDPACK VERSION    HOMEPAGE
bp-id-1         bp-version-1         mysupercoolsite.com
bp-id-2         bp-version-2         mysupercoolsite2.com

Last Failed Build
Id:              2
Build Reason:    COMMIT,BUILDPACK

`

				testhelpers.CommandTest{
					Objects:        append([]runtime.Object{image}, testhelpers.BuildsToRuntimeObjs(testNamespacedBuilds)...),
					Args:           []string{imageName, "-n", namespace},
					ExpectedOutput: expectedOutput,
				}.TestKpack(t, cmdFunc)
			})

			when("the namespace has no images", func() {
				it("returns a message that the namespace has no images", func() {
					testhelpers.CommandTest{
						Args:                []string{imageName, "-n", namespace},
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: images.kpack.io \"test-image\" not found\n",
					}.TestKpack(t, cmdFunc)

				})
			})
		})
	})

	when("a namespace is not provided", func() {
		when("the namespaces has images", func() {
			it("returns a table of image details", func() {
				image := &v1alpha2.Image{
					ObjectMeta: v1.ObjectMeta{
						Name:      imageName,
						Namespace: defaultNamespace,
					},
					Spec: v1alpha2.ImageSpec{
						Builder: corev1.ObjectReference{
							Kind: "ClusterBuilder",
							Name: "some-cluster-builder",
						},
					},
					Status: v1alpha2.ImageStatus{
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

				const expectedOutput = `Status:         Not Ready
Message:        --
LatestImage:    test-registry.io/test-image-1@sha256:abcdef123

Source
Type:    Local Source

Builder Ref
Name:    some-cluster-builder
Kind:    ClusterBuilder

Last Successful Build
Id:              1
Build Reason:    CONFIG

BUILDPACK ID    BUILDPACK VERSION    HOMEPAGE
bp-id-1         bp-version-1         mysupercoolsite.com
bp-id-2         bp-version-2         mysupercoolsite2.com

Last Failed Build
Id:              2
Build Reason:    COMMIT,BUILDPACK

`

				testhelpers.CommandTest{
					Objects:        append([]runtime.Object{image}, testhelpers.BuildsToRuntimeObjs(testBuilds)...),
					Args:           []string{imageName},
					ExpectedOutput: expectedOutput,
				}.TestKpack(t, cmdFunc)
			})

			when("the namespace has no images", func() {
				it("returns a message that the namespace has no images", func() {
					testhelpers.CommandTest{
						Args:                []string{imageName},
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: images.kpack.io \"test-image\" not found\n",
					}.TestKpack(t, cmdFunc)

				})
			})
		})
	})

	when("an image has no successful builds", func() {
		it("does not display buildpack metadata heading", func() {
			image := &v1alpha2.Image{
				ObjectMeta: v1.ObjectMeta{
					Name:      imageName,
					Namespace: defaultNamespace,
				},
				Spec: v1alpha2.ImageSpec{
					Builder: corev1.ObjectReference{
						Kind: "ClusterBuilder",
						Name: "some-cluster-builder",
					},
				},
				Status: v1alpha2.ImageStatus{
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

			const expectedOutput = `Status:         Not Ready
Message:        --
LatestImage:    test-registry.io/test-image-1@sha256:abcdef123

Source
Type:    Local Source

Builder Ref
Name:    some-cluster-builder
Kind:    ClusterBuilder

Last Successful Build
Id:              --
Build Reason:    --

Last Failed Build
Id:              --
Build Reason:    --

`
			testhelpers.CommandTest{
				Objects:        append([]runtime.Object{image}),
				Args:           []string{imageName},
				ExpectedOutput: expectedOutput,
			}.TestKpack(t, cmdFunc)
		})
	})
}

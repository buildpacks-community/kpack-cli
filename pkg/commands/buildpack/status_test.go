// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package buildpack_test

import (
	"testing"

	buildv1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/buildpack"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

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
		expectedNotReadyOutput = `Status:    Not Ready
Reason:    this buildpack is not ready for the purpose of a test

`
		expectedUnkownOutput = `Status:    Unknown

`
	)

	var (
		readyDefaultBuildpack = &buildv1alpha2.Buildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       buildv1alpha2.BuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-buildpack-1",
				Namespace: defaultNamespace,
			},
			Spec: buildv1alpha2.BuildpackSpec{
				ServiceAccountName: "default",
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-1",
				},
			},
			Status: buildv1alpha2.BuildpackStatus{
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
		notReadyDefaultBuildpack = &buildv1alpha2.Buildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       buildv1alpha2.BuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-buildpack-2",
				Namespace: defaultNamespace,
			},
			Spec: buildv1alpha2.BuildpackSpec{
				ServiceAccountName: "default",
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-2",
				},
			},
			Status: buildv1alpha2.BuildpackStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:    corev1alpha1.ConditionReady,
							Status:  corev1.ConditionFalse,
							Message: "this buildpack is not ready for the purpose of a test",
						},
					},
				},
			},
		}
		unknownDefaultBuildpack = &buildv1alpha2.Buildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       buildv1alpha2.BuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-buildpack-3",
				Namespace: defaultNamespace,
			},
			Spec: buildv1alpha2.BuildpackSpec{
				ServiceAccountName: "default",
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-3",
				},
			},
			Status: buildv1alpha2.BuildpackStatus{},
		}

		readyNamespaceBuildpack = &buildv1alpha2.Buildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       buildv1alpha2.BuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-buildpack-1",
				Namespace: "test-namespace",
			},
			Spec: buildv1alpha2.BuildpackSpec{
				ServiceAccountName: "default",
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-1",
				},
			},
			Status: buildv1alpha2.BuildpackStatus{
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
		notReadyNamespaceBuildpack = &buildv1alpha2.Buildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       buildv1alpha2.BuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-buildpack-2",
				Namespace: "test-namespace",
			},
			Spec: buildv1alpha2.BuildpackSpec{
				ServiceAccountName: "default",
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-2",
				},
			},
			Status: buildv1alpha2.BuildpackStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:    corev1alpha1.ConditionReady,
							Status:  corev1.ConditionFalse,
							Message: "this buildpack is not ready for the purpose of a test",
						},
					},
				},
			},
		}
		unknownNamespaceBuildpack = &buildv1alpha2.Buildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       buildv1alpha2.BuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-buildpack-3",
				Namespace: "test-namespace",
			},
			Spec: buildv1alpha2.BuildpackSpec{
				ServiceAccountName: "default",
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-3",
				},
			},
			Status: buildv1alpha2.BuildpackStatus{},
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

				when("the buildpack is not ready", func() {
					it("shows the build status of not ready buildpack", func() {
						testhelpers.CommandTest{
							Objects:        []runtime.Object{notReadyDefaultBuildpack},
							Args:           []string{"test-buildpack-2"},
							ExpectedOutput: expectedNotReadyOutput,
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the buildpack is unknown", func() {
					it("shows the build status of unknown buildpack", func() {
						testhelpers.CommandTest{
							Objects:        []runtime.Object{unknownDefaultBuildpack},
							Args:           []string{"test-buildpack-3"},
							ExpectedOutput: expectedUnkownOutput,
						}.TestKpack(t, cmdFunc)
					})
				})
			})

			when("the buildpack does not exist", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						Args:                []string{"non-existant-buildpack"},
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: buildpacks.kpack.io \"non-existant-buildpack\" not found\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})

		when("in the specified namespace", func() {
			when("the buildpack exists", func() {
				when("the buildpack is ready", func() {
					it("shows the build status", func() {
						testhelpers.CommandTest{
							Objects:        []runtime.Object{readyNamespaceBuildpack},
							Args:           []string{"test-buildpack-1", "-n", "test-namespace"},
							ExpectedOutput: expectedReadyOutput,
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the buildpack is not ready", func() {
					it("shows the build status of not ready buildpack", func() {
						testhelpers.CommandTest{
							Objects:        []runtime.Object{notReadyNamespaceBuildpack},
							Args:           []string{"test-buildpack-2", "-n", "test-namespace"},
							ExpectedOutput: expectedNotReadyOutput,
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the buildpack is unknown", func() {
					it("shows the build status of unknown buildpack", func() {
						testhelpers.CommandTest{
							Objects:        []runtime.Object{unknownNamespaceBuildpack},
							Args:           []string{"test-buildpack-3", "-n", "test-namespace"},
							ExpectedOutput: expectedUnkownOutput,
						}.TestKpack(t, cmdFunc)
					})
				})
			})

			when("the buildpack does not exist", func() {
				it("prints an appropriate message", func() {
					testhelpers.CommandTest{
						Args:                []string{"non-existant-buildpack", "-n", "test-namespace"},
						ExpectErr:           true,
						ExpectedErrorOutput: "Error: buildpacks.kpack.io \"non-existant-buildpack\" not found\n",
					}.TestKpack(t, cmdFunc)
				})
			})
		})
	})
}

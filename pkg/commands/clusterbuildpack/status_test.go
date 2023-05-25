// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuildpack_test

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

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/clusterbuildpack"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestClusterBuildpackStatusCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuildpackStatusCommand", testClusterBuildpackStatusCommand)
}

func testClusterBuildpackStatusCommand(t *testing.T, when spec.G, it spec.S) {
	const (
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
		readyDefaultClusterBuildpack = &buildv1alpha2.ClusterBuildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       buildv1alpha2.ClusterBuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-buildpack-1",
			},
			Spec: buildv1alpha2.ClusterBuildpackSpec{
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-1",
				},
				ServiceAccountRef: &corev1.ObjectReference{
					Namespace: "some-namespace",
					Name:      "some-serviceaccount",
				},
			},
			Status: buildv1alpha2.ClusterBuildpackStatus{
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
		notReadyDefaultClusterBuildpack = &buildv1alpha2.ClusterBuildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       buildv1alpha2.ClusterBuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-buildpack-2",
			},
			Spec: buildv1alpha2.ClusterBuildpackSpec{
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-2",
				},
				ServiceAccountRef: &corev1.ObjectReference{
					Namespace: "some-namespace",
					Name:      "some-serviceaccount",
				},
			},
			Status: buildv1alpha2.ClusterBuildpackStatus{
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
		unknownDefaultClusterBuildpack = &buildv1alpha2.ClusterBuildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       buildv1alpha2.ClusterBuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-buildpack-3",
			},
			Spec: buildv1alpha2.ClusterBuildpackSpec{
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-3",
				},
				ServiceAccountRef: &corev1.ObjectReference{
					Namespace: "some-namespace",
					Name:      "some-serviceaccount",
				},
			},
			Status: buildv1alpha2.ClusterBuildpackStatus{},
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

			when("the buildpack is not ready", func() {
				it("shows the build status of not ready buildpack", func() {
					testhelpers.CommandTest{
						Objects:        []runtime.Object{notReadyDefaultClusterBuildpack},
						Args:           []string{"test-buildpack-2"},
						ExpectedOutput: expectedNotReadyOutput,
					}.TestKpack(t, cmdFunc)
				})
			})

			when("the buildpack is unknown", func() {
				it("shows the build status of unknown buildpack", func() {
					testhelpers.CommandTest{
						Objects:        []runtime.Object{unknownDefaultClusterBuildpack},
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
					ExpectedErrorOutput: "Error: clusterbuildpacks.kpack.io \"non-existant-buildpack\" not found\n",
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}

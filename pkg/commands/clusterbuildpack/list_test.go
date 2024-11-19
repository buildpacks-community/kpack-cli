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

	"github.com/buildpacks-community/kpack-cli/pkg/commands/clusterbuildpack"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
)

func TestClusterBuildpackListCommand(t *testing.T) {
	spec.Run(t, "TestClusterBuildpackListCommand", testClusterBuildpackListCommand)
}

func testClusterBuildpackListCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		cbp1 *buildv1alpha2.ClusterBuildpack
		cbp2 *buildv1alpha2.ClusterBuildpack
		cbp3 *buildv1alpha2.ClusterBuildpack
	)

	it.Before(func() {
		cbp1 = &buildv1alpha2.ClusterBuildpack{
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

		cbp2 = &buildv1alpha2.ClusterBuildpack{
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
			},
			Status: buildv1alpha2.ClusterBuildpackStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						{
							Type:   corev1alpha1.ConditionReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
				Buildpacks: []corev1alpha1.BuildpackStatus{
					{
						BuildpackInfo: corev1alpha1.BuildpackInfo{
							Id:      "org.cloudfoundry.go",
							Version: "0.0.3",
						},
					},
				},
			},
		}

		cbp3 = &buildv1alpha2.ClusterBuildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       buildv1alpha2.ClusterBuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-buildpack-3",
				Namespace: "test-namespace",
			},
			Spec: buildv1alpha2.ClusterBuildpackSpec{
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-3",
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
							Id:      "org.cloudfoundry.java",
							Version: "1.2.3",
						},
					},
				},
			},
		}
	})

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterbuildpack.NewListCommand(clientSetProvider)
	}

	when("there are buildpacks", func() {
		it("lists the buildpacks", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					cbp1,
					cbp2,
					cbp3,
				},
				ExpectedOutput: `NAME                READY    IMAGE
test-buildpack-1    true     some-registry.com/test-buildpack-1
test-buildpack-2    false    some-registry.com/test-buildpack-2
test-buildpack-3    true     some-registry.com/test-buildpack-3

`,
			}.TestKpack(t, cmdFunc)
		})
	})

	when("there are no buildpacks", func() {
		it("prints an appropriate message", func() {
			testhelpers.CommandTest{
				ExpectErr:           true,
				ExpectedErrorOutput: "Error: no cluster buildpacks found\n",
			}.TestKpack(t, cmdFunc)
		})
	})
}

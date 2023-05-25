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

func TestBuildpackListCommand(t *testing.T) {
	spec.Run(t, "TestBuildpackListCommand", testBuildpackListCommand)
}

func testBuildpackListCommand(t *testing.T, when spec.G, it spec.S) {
	const defaultNamespace = "some-default-namespace"

	var (
		buildpack1 *buildv1alpha2.Buildpack
		buildpack2 *buildv1alpha2.Buildpack
		buildpack3 *buildv1alpha2.Buildpack
	)

	it.Before(func() {
		buildpack1 = &buildv1alpha2.Buildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       buildv1alpha2.BuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-buildpack-1",
				Namespace: defaultNamespace,
			},
			Spec: buildv1alpha2.BuildpackSpec{
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

		buildpack2 = &buildv1alpha2.Buildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       buildv1alpha2.BuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-buildpack-2",
				Namespace: defaultNamespace,
			},
			Spec: buildv1alpha2.BuildpackSpec{
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-2",
				},
			},
			Status: buildv1alpha2.BuildpackStatus{
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

		buildpack3 = &buildv1alpha2.Buildpack{
			TypeMeta: metav1.TypeMeta{
				Kind:       buildv1alpha2.BuildpackKind,
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-buildpack-3",
				Namespace: "test-namespace",
			},
			Spec: buildv1alpha2.BuildpackSpec{
				ImageSource: corev1alpha1.ImageSource{
					Image: "some-registry.com/test-buildpack-3",
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
							Id:      "org.cloudfoundry.java",
							Version: "1.2.3",
						},
					},
				},
			},
		}
	})

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return buildpack.NewListCommand(clientSetProvider)
	}

	when("namespace is not provided", func() {
		when("there are buildpacks in the default namespace", func() {
			it("lists the buildpacks", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						buildpack1,
						buildpack2,
						buildpack3,
					},
					ExpectedOutput: `NAME                READY    IMAGE
test-buildpack-1    true     some-registry.com/test-buildpack-1
test-buildpack-2    false    some-registry.com/test-buildpack-2

`,
				}.TestKpack(t, cmdFunc)
			})
		})

		when("there are no buildpacks in the default namespace", func() {
			it("prints an appropriate message", func() {
				testhelpers.CommandTest{
					ExpectErr:           true,
					ExpectedErrorOutput: "Error: no buildpacks found\n",
				}.TestKpack(t, cmdFunc)
			})
		})
	})

	when("namespace is provided", func() {
		when("there are buildpacks in the namespace", func() {
			it("lists the buildpacks", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						buildpack1,
						buildpack2,
						buildpack3,
					},
					Args: []string{"-n", "test-namespace"},
					ExpectedOutput: `NAME                READY    IMAGE
test-buildpack-3    true     some-registry.com/test-buildpack-3

`,
				}.TestKpack(t, cmdFunc)
			})
		})

		when("there are no buildpacks in the namespace", func() {
			it("prints an appropriate message", func() {
				testhelpers.CommandTest{
					ExpectErr:           true,
					ExpectedErrorOutput: "Error: no buildpacks found\n",
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}

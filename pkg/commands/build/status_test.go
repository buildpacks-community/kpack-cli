// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package build_test

import (
	"testing"
	"time"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pivotal/build-service-cli/pkg/commands/build"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestBuildStatusCommand(t *testing.T) {
	spec.Run(t, "TestBuildStatusCommand", testBuildStatusCommand)
}

func testBuildStatusCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		image                       = "test-image"
		defaultNamespace            = "some-default-namespace"
		expectedOutputForMostRecent = `Image:            repo.com/image-3:tag
Status:           BUILDING
Build Reasons:    TRIGGER

Pod Name:    pod-three

Builder:      some-repo.com/my-builder
Run Image:    some-repo.com/run-image

Source:    Local Source

BUILDPACK ID    BUILDPACK VERSION
bp-id-1         bp-version-1
bp-id-2         bp-version-2

`
		expectedOutputForBuildNumber = `Image:            repo.com/image-1:tag
Status:           SUCCESS
Build Reasons:    CONFIG

Pod Name:    pod-one

Builder:      some-repo.com/my-builder
Run Image:    some-repo.com/run-image

Source:    Local Source

BUILDPACK ID    BUILDPACK VERSION
bp-id-1         bp-version-1
bp-id-2         bp-version-2

`
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)
		return build.NewStatusCommand(clientSetProvider)
	}

	when("getting build status", func() {
		when("in the default namespace", func() {
			when("the build exists", func() {
				when("the build flag is provided", func() {
					it("shows the build status", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.MakeTestBuilds(image, defaultNamespace),
							Args:           []string{image, "-b", "1"},
							ExpectedOutput: expectedOutputForBuildNumber,
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the build flag is not provided", func() {
					it("shows the build status of the most recent build", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.MakeTestBuilds(image, defaultNamespace),
							Args:           []string{image},
							ExpectedOutput: expectedOutputForMostRecent,
						}.TestKpack(t, cmdFunc)
					})
				})
			})

			when("the build does not exist", func() {
				when("the build flag is provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.MakeTestBuilds(image, defaultNamespace),
							Args:           []string{image, "-b", "123"},
							ExpectErr:      true,
							ExpectedOutput: "Error: build \"123\" not found\n",
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the build flag was not provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Args:           []string{image},
							ExpectErr:      true,
							ExpectedOutput: "Error: no builds found\n",
						}.TestKpack(t, cmdFunc)
					})
				})
			})
		})

		when("in a given namespace", func() {
			const namespace = "some-namespace"

			when("the build exists", func() {
				when("the build flag is provided", func() {
					it("gets the build status", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.MakeTestBuilds(image, namespace),
							Args:           []string{image, "-b", "1", "-n", namespace},
							ExpectedOutput: expectedOutputForBuildNumber,
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the build flag is not provided", func() {
					it("shows the build status of the most recent build", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.MakeTestBuilds(image, namespace),
							Args:           []string{image, "-n", namespace},
							ExpectedOutput: expectedOutputForMostRecent,
						}.TestKpack(t, cmdFunc)
					})
				})
			})

			when("the build does not exist", func() {
				when("the build flag is provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Objects:        testhelpers.MakeTestBuilds(image, namespace),
							Args:           []string{image, "-b", "123", "-n", namespace},
							ExpectErr:      true,
							ExpectedOutput: "Error: build \"123\" not found\n",
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the build flag was not provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Args:           []string{image, "-n", namespace},
							ExpectErr:      true,
							ExpectedOutput: "Error: no builds found\n",
						}.TestKpack(t, cmdFunc)
					})
				})
			})
		})

		when("build status returns a reason and message", func() {
			it("displays status reason and status message", func() {
				expectedOutput := `Image:             repo.com/image-3:tag
Status:            BUILDING
Build Reasons:     TRIGGER
Status Reason:     some-reason
Status Message:    some-message

Pod Name:    some-pod

Builder:      some-repo.com/my-builder
Run Image:    some-repo.com/run-image

Source:    Local Source

BUILDPACK ID    BUILDPACK VERSION
bp-id-1         bp-version-1
bp-id-2         bp-version-2

`
				bld := &v1alpha1.Build{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "bld-three",
						Namespace:         "some-default-namespace",
						CreationTimestamp: metav1.Time{Time: time.Time{}.Add(5 * time.Hour)},
						Labels: map[string]string{
							v1alpha1.ImageLabel:       image,
							v1alpha1.BuildNumberLabel: "3",
						},
						Annotations: map[string]string{
							v1alpha1.BuildReasonAnnotation: "TRIGGER",
						},
					},
					Spec: v1alpha1.BuildSpec{
						Builder: v1alpha1.BuildBuilderSpec{
							Image: "some-repo.com/my-builder",
						},
					},
					Status: v1alpha1.BuildStatus{
						Status: corev1alpha1.Status{
							Conditions: corev1alpha1.Conditions{
								{
									Type:    corev1alpha1.ConditionSucceeded,
									Status:  corev1.ConditionUnknown,
									Reason:  "some-reason",
									Message: "some-message",
								},
							},
						},
						BuildMetadata: v1alpha1.BuildpackMetadataList{
							{
								Id:      "bp-id-1",
								Version: "bp-version-1",
							},
							{
								Id:      "bp-id-2",
								Version: "bp-version-2",
							},
						},
						Stack: v1alpha1.BuildStack{
							RunImage: "some-repo.com/run-image",
						},
						LatestImage: "repo.com/image-3:tag",
						PodName: "some-pod",
					},
				}
				testhelpers.CommandTest{
					Objects:        []runtime.Object{bld},
					Args:           []string{image},
					ExpectedOutput: expectedOutput,
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}

// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterlifecycle_test

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

	"github.com/buildpacks-community/kpack-cli/pkg/commands/clusterlifecycle"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
)

func TestClusterLifecycleStatusCommand(t *testing.T) {
	spec.Run(t, "TestClusterLifecycleStatusCommand", testClusterLifecycleStatusCommand)
}

func testClusterLifecycleStatusCommand(t *testing.T, when spec.G, it spec.S) {
	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterlifecycle.NewStatusCommand(clientSetProvider)
	}

	when("the lifecycle exists", func() {
		lifecycle := &v1alpha2.ClusterLifecycle{
			ObjectMeta: metav1.ObjectMeta{
				Name: "some-lifecycle",
			},
			Status: v1alpha2.ClusterLifecycleStatus{
				ResolvedClusterLifecycle: v1alpha2.ResolvedClusterLifecycle{
					Version: "0.17.0",
					Image: v1alpha2.ClusterLifecycleStatusImage{
						LatestImage: "registry.io/repo/lifecycle@sha256:abc123",
					},
					APIs: v1alpha2.LifecycleAPIs{
						Buildpack: v1alpha2.APIVersions{
							Supported:  []string{"0.2", "0.3", "0.10"},
							Deprecated: []string{"0.2"},
						},
						Platform: v1alpha2.APIVersions{
							Supported:  []string{"0.3", "0.12"},
							Deprecated: []string{},
						},
					},
				},
			},
		}

		it("returns lifecycle details", func() {
			const expectedOutput = `Status:     Unknown
Image:      registry.io/repo/lifecycle@sha256:abc123
Version:    0.17.0

`

			testhelpers.CommandTest{
				Objects:        []runtime.Object{lifecycle},
				Args:           []string{"some-lifecycle"},
				ExpectedOutput: expectedOutput,
			}.TestKpack(t, cmdFunc)
		})

		it("includes buildpack APIs when --verbose flag is used", func() {
			const expectedOutput = `Status:                       Unknown
Image:                        registry.io/repo/lifecycle@sha256:abc123
Version:                      0.17.0
Supported Buildpack APIs:     0.2, 0.3, 0.10
Deprecated Buildpack APIs:    0.2

`

			testhelpers.CommandTest{
				Objects:        []runtime.Object{lifecycle},
				Args:           []string{"some-lifecycle", "--verbose"},
				ExpectedOutput: expectedOutput,
			}.TestKpack(t, cmdFunc)
		})

		when("the status is ready", func() {
			it("prints Ready status", func() {
				readyLifecycle := lifecycle.DeepCopy()
				readyLifecycle.Status.Conditions = []corev1alpha1.Condition{
					{
						Type:   corev1alpha1.ConditionReady,
						Status: corev1.ConditionTrue,
					},
				}

				const expectedOutput = `Status:     Ready
Image:      registry.io/repo/lifecycle@sha256:abc123
Version:    0.17.0

`

				testhelpers.CommandTest{
					Objects:        []runtime.Object{readyLifecycle},
					Args:           []string{"some-lifecycle"},
					ExpectedOutput: expectedOutput,
				}.TestKpack(t, cmdFunc)
			})
		})

		when("the status is not ready", func() {
			it("prints the status message", func() {
				notReadyLifecycle := lifecycle.DeepCopy()
				notReadyLifecycle.Status.Conditions = []corev1alpha1.Condition{
					{
						Type:    corev1alpha1.ConditionReady,
						Status:  corev1.ConditionFalse,
						Message: "some sample message",
					},
				}

				const expectedOutput = `Status:     Not Ready - some sample message
Image:      registry.io/repo/lifecycle@sha256:abc123
Version:    0.17.0

`

				testhelpers.CommandTest{
					Objects:        []runtime.Object{notReadyLifecycle},
					Args:           []string{"some-lifecycle"},
					ExpectedOutput: expectedOutput,
				}.TestKpack(t, cmdFunc)
			})
		})
	})

	when("the lifecycle does not exist", func() {
		it("returns a message that there is no lifecycle", func() {
			testhelpers.CommandTest{
				Args:                []string{"lifecycle-does-not-exist"},
				ExpectErr:           true,
				ExpectedErrorOutput: "Error: clusterlifecycles.kpack.io \"lifecycle-does-not-exist\" not found\n",
			}.TestKpack(t, cmdFunc)
		})
	})
}

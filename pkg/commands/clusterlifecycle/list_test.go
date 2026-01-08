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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/buildpacks-community/kpack-cli/pkg/commands/clusterlifecycle"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
)

func TestClusterLifecycleListCommand(t *testing.T) {
	spec.Run(t, "TestClusterLifecycleListCommand", testClusterLifecycleListCommand)
}

func testClusterLifecycleListCommand(t *testing.T, when spec.G, it spec.S) {
	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackClusterProvider(clientSet)
		return clusterlifecycle.NewListCommand(clientSetProvider)
	}

	when("the cluster has lifecycles", func() {
		it("returns a table of lifecycle details", func() {
			lifecycle1 := &v1alpha2.ClusterLifecycle{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-lifecycle-1",
				},
				Status: v1alpha2.ClusterLifecycleStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionFalse,
							},
						},
					},
					ResolvedClusterLifecycle: v1alpha2.ResolvedClusterLifecycle{
						Version: "0.17.0",
						Image: v1alpha2.ClusterLifecycleStatusImage{
							LatestImage: "registry.io/repo/lifecycle@sha256:abc123",
						},
					},
				},
			}
			lifecycle2 := &v1alpha2.ClusterLifecycle{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-lifecycle-2",
				},
				Status: v1alpha2.ClusterLifecycleStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
					ResolvedClusterLifecycle: v1alpha2.ResolvedClusterLifecycle{
						Version: "0.18.0",
						Image: v1alpha2.ClusterLifecycleStatusImage{
							LatestImage: "registry.io/repo/lifecycle@sha256:def456",
						},
					},
				},
			}
			lifecycle3 := &v1alpha2.ClusterLifecycle{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-lifecycle-3",
				},
				Status: v1alpha2.ClusterLifecycleStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionUnknown,
							},
						},
					},
					ResolvedClusterLifecycle: v1alpha2.ResolvedClusterLifecycle{
						Version: "0.16.5",
						Image: v1alpha2.ClusterLifecycleStatusImage{
							LatestImage: "registry.io/repo/lifecycle@sha256:ghi789",
						},
					},
				},
			}

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					lifecycle1,
					lifecycle2,
					lifecycle3,
				},
				ExpectedOutput: `NAME                READY      VERSION    IMAGE
test-lifecycle-1    False      0.17.0     registry.io/repo/lifecycle@sha256:abc123
test-lifecycle-2    True       0.18.0     registry.io/repo/lifecycle@sha256:def456
test-lifecycle-3    Unknown    0.16.5     registry.io/repo/lifecycle@sha256:ghi789

`,
			}.TestKpack(t, cmdFunc)
		})

		when("there are no lifecycles", func() {
			it("returns a message that no lifecycles were found", func() {
				testhelpers.CommandTest{
					ExpectErr:           true,
					ExpectedErrorOutput: "Error: no clusterlifecycles found\n",
				}.TestKpack(t, cmdFunc)

			})
		})
	})
}

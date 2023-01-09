// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import_test

import (
	"testing"

	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	importpkg "github.com/vmware-tanzu/kpack-cli/pkg/import"
)

func TestDescriptor(t *testing.T) {
	spec.Run(t, "TestDescriptor", testDescriptor)
}

func testDescriptor(t *testing.T, when spec.G, it spec.S) {
	desc := importpkg.DependencyDescriptor{
		DefaultClusterStack:   "some-stack",
		DefaultClusterBuilder: "some-cb",
		ClusterStores: []importpkg.ClusterStore{
			{
				Name: "some-store",
				Sources: []importpkg.Source{
					{
						Image: "some-store-image",
					},
				},
			},
		},
		ClusterStacks: []importpkg.ClusterStack{
			{
				Name: "some-stack",
				BuildImage: importpkg.Source{
					Image: "build-image",
				},
				RunImage: importpkg.Source{
					Image: "run-image",
				},
			},
		},
		ClusterBuilders: []importpkg.ClusterBuilder{
			{
				Name:         "some-cb",
				ClusterStack: "some-stack",
				ClusterStore: "some-store",
				Order: []corev1alpha1.OrderEntry{
					{
						Group: []corev1alpha1.BuildpackRef{
							{
								BuildpackInfo: corev1alpha1.BuildpackInfo{
									Id:      "some-buildpack",
									Version: "1.2.3",
								},
								Optional: false,
							},
						},
					},
				},
			},
		},
	}
	when("#Validate", func() {

		it("validates successfully", func() {
			require.NoError(t, desc.Validate())
		})

		when("there is a duplicate store name", func() {
			desc.ClusterStores = append(desc.ClusterStores, importpkg.ClusterStore{
				Name: "some-store",
			})

			it("fails validation", func() {
				require.Error(t, desc.Validate())
			})
		})

		when("there is a duplicate stack name", func() {
			desc.ClusterStacks = append(desc.ClusterStacks, importpkg.ClusterStack{
				Name: "some-stack",
			})

			it("fails validation", func() {
				require.Error(t, desc.Validate())
			})
		})

		when("there is a duplicate cb name", func() {
			desc.ClusterBuilders = append(desc.ClusterBuilders, importpkg.ClusterBuilder{
				Name:         "some-cb",
				ClusterStack: "some-stack",
				ClusterStore: "some-store",
			})

			it("fails validation", func() {
				require.Error(t, desc.Validate())
			})
		})

		when("the default stack does not exist", func() {
			desc.DefaultClusterStack = "does-not-exist"

			it("fails validation", func() {
				require.Error(t, desc.Validate())
			})
		})

		when("there is no default clusterstack", func() {
			desc.DefaultClusterStack = ""

			it("validates successfully", func() {
				require.NoError(t, desc.Validate())
			})
		})

		when("the default clusterbuilder does not exist", func() {
			desc.DefaultClusterBuilder = "does-not-exist"

			it("fails validation", func() {
				require.Error(t, desc.Validate())
			})
		})

		when("there is no default clusterbuilder", func() {
			desc.DefaultClusterBuilder = ""

			it("validates successfully", func() {
				require.NoError(t, desc.Validate())
			})
		})
	})

	when("#GetClusterStacks", func() {
		it("returns the cluster stacks and the default cluster stack", func() {
			stacks := desc.GetClusterStacks()
			expectedStacks := []importpkg.ClusterStack{
				{Name: "some-stack", BuildImage: importpkg.Source{Image: "build-image"}, RunImage: importpkg.Source{Image: "run-image"}},
				{Name: "default", BuildImage: importpkg.Source{Image: "build-image"}, RunImage: importpkg.Source{Image: "run-image"}}}
			require.Equal(t, expectedStacks, stacks)
		})
	})

	when("#GetClusterBuilders", func() {
		it("returns the cluster builders and the default cluster builder", func() {
			builders := desc.GetClusterBuilders()
			expectedBuilders := []importpkg.ClusterBuilder{
				{Name: "some-cb", ClusterStack: "some-stack", ClusterStore: "some-store", Order: []corev1alpha1.OrderEntry{{Group: []corev1alpha1.BuildpackRef{{BuildpackInfo: corev1alpha1.BuildpackInfo{Id: "some-buildpack", Version: "1.2.3"}, Optional: false}}}}},
				{Name: "default", ClusterStack: "some-stack", ClusterStore: "some-store", Order: []corev1alpha1.OrderEntry{{Group: []corev1alpha1.BuildpackRef{{BuildpackInfo: corev1alpha1.BuildpackInfo{Id: "some-buildpack", Version: "1.2.3"}, Optional: false}}}}},
			}
			require.Equal(t, expectedBuilders, builders)
		})
	})

	when("checking for deprecations of defaults", func() {
		it("returns true if default stack is deprecated", func() {
			desc.ClusterStacks[0].Deprecated = true
			require.True(t, desc.IsDefaultStackDeprecated())
		})
		it("returns true if default builder is deprecated", func() {
			desc.ClusterBuilders[0].Deprecated = true
			require.True(t, desc.IsDefaultBuilderDeprecated())
		})
	})
}

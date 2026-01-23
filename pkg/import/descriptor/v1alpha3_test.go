// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package descriptor_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/buildpacks-community/kpack-cli/pkg/import/descriptor"
)

func TestDescriptorV1Alpha3(t *testing.T) {
	spec.Run(t, "TestDescriptorV1Alpha3", testDescriptorV1Alpha3)
}

func testDescriptorV1Alpha3(t *testing.T, when spec.G, it spec.S) {
	when("#ToV1", func() {
		descV1Alpha3 := descriptor.DependencyDescriptorV1Alpha3{
			DefaultClusterStack:   "some-stack",
			DefaultClusterBuilder: "some-ccb",
			Lifecycle: descriptor.Lifecycle{
				Image: "some-lifecycle-image",
			},
			ClusterStores: []descriptor.ClusterStore{
				{
					Name: "some-store",
					Sources: []descriptor.Source{
						{
							Image: "some-store-image",
						},
					},
				},
			},
			ClusterStacks: []descriptor.ClusterStack{
				{
					Name: "some-stack",
					BuildImage: descriptor.Source{
						Image: "build-image",
					},
					RunImage: descriptor.Source{
						Image: "run-image",
					},
				},
			},
			ClusterBuilders: []descriptor.ClusterBuilder{
				{
					Name:         "some-ccb",
					ClusterStack: "some-stack",
					ClusterStore: "some-store",
				},
			},
		}

		it("converts successfully", func() {
			v1 := descV1Alpha3.ToV1()
			require.Equal(t, descriptor.APIVersionV1, v1.APIVersion)
			require.Len(t, v1.ClusterLifecycles, 1)
			require.Equal(t, v1alpha2.DefaultLifecycleName, v1.ClusterLifecycles[0].Name)
			require.Equal(t, "some-lifecycle-image", v1.ClusterLifecycles[0].Image)
			require.Empty(t, v1.ClusterBuildpacks)
		})

		it("converts with empty lifecycle", func() {
			descV1Alpha3.Lifecycle = descriptor.Lifecycle{}
			v1 := descV1Alpha3.ToV1()
			require.Len(t, v1.ClusterLifecycles, 0)
		})
	})
}

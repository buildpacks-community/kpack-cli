// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	importpkg "github.com/vmware-tanzu/kpack-cli/pkg/import"
)

func TestDescriptorV1(t *testing.T) {
	spec.Run(t, "TestDescriptorV1", testDescriptorV1)
}

func testDescriptorV1(t *testing.T, when spec.G, it spec.S) {
	when("#ToNextVersion", func() {
		descV1 := importpkg.DependencyDescriptorV1{
			DefaultStack:          "some-stack",
			DefaultClusterBuilder: "some-ccb",
			Stores: []importpkg.ClusterStore{
				{
					Name: "some-store",
					Sources: []importpkg.Source{
						{
							Image: "some-store-image",
						},
					},
				},
			},
			Stacks: []importpkg.ClusterStack{
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
			ClusterBuilders: []importpkg.ClusterBuilderV1{
				{
					Name:  "some-ccb",
					Stack: "some-stack",
					Store: "some-store",
					Order: []v1alpha1.OrderEntry{
						{
							Group: []v1alpha1.BuildpackRef{
								{
									BuildpackInfo: v1alpha1.BuildpackInfo{
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

		it("converts successfully", func() {
			d := descV1.ToNextVersion()
			require.NoError(t, d.Validate())
		})
	})
}

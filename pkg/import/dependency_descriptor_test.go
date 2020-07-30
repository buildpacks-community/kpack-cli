package _import_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	importpkg "github.com/pivotal/build-service-cli/pkg/import"
)

func TestDescriptorValidate(t *testing.T) {
	spec.Run(t, "TestDescriptorValidate", testDescriptorValidate)
}

func testDescriptorValidate(t *testing.T, when spec.G, it spec.S) {
	desc := importpkg.DependencyDescriptor{
		DefaultStack:          "some-stack",
		DefaultClusterBuilder: "some-ccb",
		Stores: []importpkg.Store{
			{
				Name: "some-store",
				Sources: []importpkg.Source{
					{
						Image: "some-store-image",
					},
				},
			},
		},
		Stacks: []importpkg.Stack{
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

	it("validates successfully", func() {
		require.NoError(t, desc.Validate())
	})

	when("there is a duplicate store name", func() {
		desc.Stores = append(desc.Stores, importpkg.Store{
			Name: "some-store",
		})

		it("fails validation", func() {
			require.Error(t, desc.Validate())
		})
	})

	when("there is a duplicate stack name", func() {
		desc.Stacks = append(desc.Stacks, importpkg.Stack{
			Name: "some-stack",
		})

		it("fails validation", func() {
			require.Error(t, desc.Validate())
		})
	})

	when("there is a duplicate ccb name", func() {
		desc.ClusterBuilders = append(desc.ClusterBuilders, importpkg.ClusterBuilder{
			Name:  "some-ccb",
			Stack: "some-stack",
			Store: "some-store",
		})

		it("fails validation", func() {
			require.Error(t, desc.Validate())
		})
	})

	when("the default stack does not exist", func() {
		desc.DefaultStack = "does-not-exist"

		it("fails validation", func() {
			require.Error(t, desc.Validate())
		})
	})

	when("the default ccb does not exist", func() {
		desc.DefaultStack = "does-not-exist"

		it("fails validation", func() {
			require.Error(t, desc.Validate())
		})
	})

	when("the ccb uses a stack that does not exist", func() {
		desc.ClusterBuilders[0].Stack = "does-not-exist"

		it("fails validation", func() {
			require.Error(t, desc.Validate())
		})
	})

	when("the ccb uses a store that does not exist", func() {
		desc.ClusterBuilders[0].Store = "does-not-exist"

		it("fails validation", func() {
			require.Error(t, desc.Validate())
		})
	})
}

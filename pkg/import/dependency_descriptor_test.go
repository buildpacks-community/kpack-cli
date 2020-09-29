package _import_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	importpkg "github.com/pivotal/build-service-cli/pkg/import"
)

func TestDescriptor(t *testing.T) {
	spec.Run(t, "TestDescriptor", testDescriptor)
}

func testDescriptor(t *testing.T, when spec.G, it spec.S) {
	when("#Validate", func() {
		desc := importpkg.DependencyDescriptor{
			DefaultClusterStack:   "some-stack",
			DefaultClusterBuilder: "some-ccb",
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
					Name:         "some-ccb",
					ClusterStack: "some-stack",
					ClusterStore: "some-store",
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

		when("there is a duplicate ccb name", func() {
			desc.ClusterBuilders = append(desc.ClusterBuilders, importpkg.ClusterBuilder{
				Name:         "some-ccb",
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

		when("the default ccb does not exist", func() {
			desc.DefaultClusterStack = "does-not-exist"

			it("fails validation", func() {
				require.Error(t, desc.Validate())
			})
		})

		when("the ccb uses a stack that does not exist", func() {
			desc.ClusterBuilders[0].ClusterStack = "does-not-exist"

			it("fails validation", func() {
				require.Error(t, desc.Validate())
			})
		})

		when("the ccb uses a store that does not exist", func() {
			desc.ClusterBuilders[0].ClusterStore = "does-not-exist"

			it("fails validation", func() {
				require.Error(t, desc.Validate())
			})
		})
	})
}

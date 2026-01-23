// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	importpkg "github.com/buildpacks-community/kpack-cli/pkg/import"
)

func TestDescriptor(t *testing.T) {
	spec.Run(t, "TestDescriptor", testDescriptor)
}

func testDescriptor(t *testing.T, when spec.G, it spec.S) {
	desc := importpkg.DependencyDescriptor{
		DefaultClusterStack:   "some-stack",
		DefaultClusterBuilder: "some-cb",
		ClusterLifecycles: []importpkg.ClusterLifecycle{
			{
				Name:  "some-lifecycle",
				Image: "lifecycle-image",
			},
		},
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
				Order: []v1alpha2.BuilderOrderEntry{
					{
						Group: []v1alpha2.BuilderBuildpackRef{
							{
								BuildpackRef: corev1alpha1.BuildpackRef{
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
		},
	}
	when("#ValidateDescriptor", func() {

		it("validates successfully", func() {
			require.NoError(t, importpkg.ValidateDescriptor(desc))
		})

		when("there is a duplicate lifecycle name", func() {
			it("fails validation with the duplicate name in the error message", func() {
				descWithDupe := importpkg.DependencyDescriptor{
					ClusterLifecycles: []importpkg.ClusterLifecycle{
						{Name: "my-lifecycle", Image: "image1"},
						{Name: "my-lifecycle", Image: "image2"},
					},
				}
				err := importpkg.ValidateDescriptor(descWithDupe)
				require.Error(t, err)
				require.Contains(t, err.Error(), "duplicate cluster lifecycle name 'my-lifecycle'")
			})
		})

		when("there is a lifecycle with empty name", func() {
			it("fails validation", func() {
				descWithEmptyName := importpkg.DependencyDescriptor{
					ClusterLifecycles: []importpkg.ClusterLifecycle{
						{
							Name:  "",
							Image: "some-image",
						},
					},
				}
				err := importpkg.ValidateDescriptor(descWithEmptyName)
				require.Error(t, err)
				require.Contains(t, err.Error(), "cluster lifecycle name cannot be empty")
			})
		})

		when("there is a buildpack with empty name", func() {
			it("fails validation", func() {
				descWithEmptyName := importpkg.DependencyDescriptor{
					ClusterBuildpacks: []importpkg.ClusterBuildpack{
						{
							Name:  "",
							Image: "some-image",
						},
					},
				}
				err := importpkg.ValidateDescriptor(descWithEmptyName)
				require.Error(t, err)
				require.Contains(t, err.Error(), "cluster buildpack name cannot be empty")
			})
		})

		when("there is a duplicate buildpack name", func() {
			it("fails validation with the duplicate name in the error message", func() {
				descWithDupe := importpkg.DependencyDescriptor{
					ClusterBuildpacks: []importpkg.ClusterBuildpack{
						{Name: "my-bp", Image: "image1"},
						{Name: "my-bp", Image: "image2"},
					},
				}
				err := importpkg.ValidateDescriptor(descWithDupe)
				require.Error(t, err)
				require.Contains(t, err.Error(), "duplicate cluster buildpack name 'my-bp'")
			})
		})

		when("there is a duplicate store name", func() {
			it("fails validation with the duplicate name in the error message", func() {
				descWithDupe := importpkg.DependencyDescriptor{
					ClusterStores: []importpkg.ClusterStore{
						{Name: "my-store", Sources: []importpkg.Source{{Image: "image1"}}},
						{Name: "my-store", Sources: []importpkg.Source{{Image: "image2"}}},
					},
				}
				err := importpkg.ValidateDescriptor(descWithDupe)
				require.Error(t, err)
				require.Contains(t, err.Error(), "duplicate store name 'my-store'")
			})
		})

		when("there is a duplicate stack name", func() {
			it("fails validation with the duplicate name in the error message", func() {
				descWithDupe := importpkg.DependencyDescriptor{
					ClusterStacks: []importpkg.ClusterStack{
						{Name: "my-stack", BuildImage: importpkg.Source{Image: "build1"}, RunImage: importpkg.Source{Image: "run1"}},
						{Name: "my-stack", BuildImage: importpkg.Source{Image: "build2"}, RunImage: importpkg.Source{Image: "run2"}},
					},
				}
				err := importpkg.ValidateDescriptor(descWithDupe)
				require.Error(t, err)
				require.Contains(t, err.Error(), "duplicate stack name 'my-stack'")
			})
		})

		when("there is a duplicate cb name", func() {
			it("fails validation with the duplicate name in the error message", func() {
				descWithDupe := importpkg.DependencyDescriptor{
					ClusterBuilders: []importpkg.ClusterBuilder{
						{Name: "my-builder", ClusterStack: "stack1", ClusterStore: "store1"},
						{Name: "my-builder", ClusterStack: "stack2", ClusterStore: "store2"},
					},
				}
				err := importpkg.ValidateDescriptor(descWithDupe)
				require.Error(t, err)
				require.Contains(t, err.Error(), "duplicate cluster builder name 'my-builder'")
			})
		})

		when("the default stack does not exist", func() {
			desc.DefaultClusterStack = "does-not-exist"

			it("fails validation", func() {
				require.Error(t, importpkg.ValidateDescriptor(desc))
			})
		})

		when("there is no default clusterstack", func() {
			desc.DefaultClusterStack = ""

			it("validates successfully", func() {
				require.NoError(t, importpkg.ValidateDescriptor(desc))
			})
		})

		when("the default clusterbuilder does not exist", func() {
			desc.DefaultClusterBuilder = "does-not-exist"

			it("fails validation", func() {
				require.Error(t, importpkg.ValidateDescriptor(desc))
			})
		})

		when("there is no default clusterbuilder", func() {
			desc.DefaultClusterBuilder = ""

			it("validates successfully", func() {
				require.NoError(t, importpkg.ValidateDescriptor(desc))
			})
		})
	})

	when("#GetClusterStacks", func() {
		it("returns the cluster stacks and the default cluster stack", func() {
			stacks := importpkg.GetClusterStacks(desc)
			expectedStacks := []importpkg.ClusterStack{
				{Name: "some-stack", BuildImage: importpkg.Source{Image: "build-image"}, RunImage: importpkg.Source{Image: "run-image"}},
				{Name: "default", BuildImage: importpkg.Source{Image: "build-image"}, RunImage: importpkg.Source{Image: "run-image"}}}
			require.Equal(t, expectedStacks, stacks)
		})
	})

	when("#GetClusterBuilders", func() {
		it("returns the cluster builders and the default cluster builder", func() {
			builders := importpkg.GetClusterBuilders(desc)
			expectedBuilders := []importpkg.ClusterBuilder{
				{
					Name:         "some-cb",
					ClusterStack: "some-stack",
					ClusterStore: "some-store",
					Order: []v1alpha2.BuilderOrderEntry{
						{
							Group: []v1alpha2.BuilderBuildpackRef{
								{
									BuildpackRef: corev1alpha1.BuildpackRef{
										BuildpackInfo: corev1alpha1.BuildpackInfo{
											Id: "some-buildpack", Version: "1.2.3",
										},
										Optional: false,
									},
								},
							},
						},
					},
				},
				{
					Name:         "default",
					ClusterStack: "some-stack",
					ClusterStore: "some-store",
					Order: []v1alpha2.BuilderOrderEntry{
						{
							Group: []v1alpha2.BuilderBuildpackRef{
								{
									BuildpackRef: corev1alpha1.BuildpackRef{
										BuildpackInfo: corev1alpha1.BuildpackInfo{
											Id: "some-buildpack", Version: "1.2.3",
										},
										Optional: false,
									},
								},
							},
						},
					},
				},
			}
			require.Equal(t, expectedBuilders, builders)
		})
	})

	when("#GetClusterLifecycles", func() {
		when("there is a default lifecycle", func() {
			it("returns the cluster lifecycles and the default cluster lifecycle", func() {
				descWithDefault := importpkg.DependencyDescriptor{
					DefaultClusterLifecycle: "some-lifecycle",
					ClusterLifecycles: []importpkg.ClusterLifecycle{
						{Name: "some-lifecycle", Image: "lifecycle-image"},
						{Name: "other-lifecycle", Image: "other-image"},
					},
				}
				lifecycles := importpkg.GetClusterLifecycles(descWithDefault)
				expectedLifecycles := []importpkg.ClusterLifecycle{
					{Name: "some-lifecycle", Image: "lifecycle-image"},
					{Name: "other-lifecycle", Image: "other-image"},
					{Name: v1alpha2.DefaultLifecycleName, Image: "lifecycle-image"},
				}
				require.Equal(t, expectedLifecycles, lifecycles)
			})
		})

		when("there is no default lifecycle", func() {
			it("returns only the cluster lifecycles", func() {
				descWithoutDefault := importpkg.DependencyDescriptor{
					ClusterLifecycles: []importpkg.ClusterLifecycle{
						{Name: "some-lifecycle", Image: "lifecycle-image"},
					},
				}
				lifecycles := importpkg.GetClusterLifecycles(descWithoutDefault)
				expectedLifecycles := []importpkg.ClusterLifecycle{
					{Name: "some-lifecycle", Image: "lifecycle-image"},
				}
				require.Equal(t, expectedLifecycles, lifecycles)
			})
		})
	})

	when("#GetClusterBuildpacks", func() {
		when("there is a default buildpack", func() {
			it("returns the cluster buildpacks and the default cluster buildpack", func() {
				descWithDefault := importpkg.DependencyDescriptor{
					DefaultClusterBuildpack: "some-buildpack",
					ClusterBuildpacks: []importpkg.ClusterBuildpack{
						{Name: "some-buildpack", Image: "buildpack-image"},
						{Name: "other-buildpack", Image: "other-image"},
					},
				}
				buildpacks := importpkg.GetClusterBuildpacks(descWithDefault)
				expectedBuildpacks := []importpkg.ClusterBuildpack{
					{Name: "some-buildpack", Image: "buildpack-image"},
					{Name: "other-buildpack", Image: "other-image"},
					{Name: "default", Image: "buildpack-image"},
				}
				require.Equal(t, expectedBuildpacks, buildpacks)
			})
		})

		when("there is no default buildpack", func() {
			it("returns only the cluster buildpacks", func() {
				descWithoutDefault := importpkg.DependencyDescriptor{
					ClusterBuildpacks: []importpkg.ClusterBuildpack{
						{Name: "some-buildpack", Image: "buildpack-image"},
					},
				}
				buildpacks := importpkg.GetClusterBuildpacks(descWithoutDefault)
				expectedBuildpacks := []importpkg.ClusterBuildpack{
					{Name: "some-buildpack", Image: "buildpack-image"},
				}
				require.Equal(t, expectedBuildpacks, buildpacks)
			})
		})
	})

	when("validating default lifecycle", func() {
		when("the default lifecycle does not exist", func() {
			it("fails validation", func() {
				descWithBadDefault := importpkg.DependencyDescriptor{
					DefaultClusterLifecycle: "does-not-exist",
					ClusterLifecycles: []importpkg.ClusterLifecycle{
						{Name: "some-lifecycle", Image: "lifecycle-image"},
					},
				}
				err := importpkg.ValidateDescriptor(descWithBadDefault)
				require.Error(t, err)
				require.Contains(t, err.Error(), "default cluster lifecycle 'does-not-exist' not found")
			})
		})

		when("there is no default cluster lifecycle", func() {
			it("validates successfully", func() {
				descWithoutDefault := importpkg.DependencyDescriptor{
					ClusterLifecycles: []importpkg.ClusterLifecycle{
						{Name: "some-lifecycle", Image: "lifecycle-image"},
					},
				}
				require.NoError(t, importpkg.ValidateDescriptor(descWithoutDefault))
			})
		})
	})

	when("validating default buildpack", func() {
		when("the default buildpack does not exist", func() {
			it("fails validation", func() {
				descWithBadDefault := importpkg.DependencyDescriptor{
					DefaultClusterBuildpack: "does-not-exist",
					ClusterBuildpacks: []importpkg.ClusterBuildpack{
						{Name: "some-buildpack", Image: "buildpack-image"},
					},
				}
				err := importpkg.ValidateDescriptor(descWithBadDefault)
				require.Error(t, err)
				require.Contains(t, err.Error(), "default cluster buildpack 'does-not-exist' not found")
			})
		})

		when("there is no default cluster buildpack", func() {
			it("validates successfully", func() {
				descWithoutDefault := importpkg.DependencyDescriptor{
					ClusterBuildpacks: []importpkg.ClusterBuildpack{
						{Name: "some-buildpack", Image: "buildpack-image"},
					},
				}
				require.NoError(t, importpkg.ValidateDescriptor(descWithoutDefault))
			})
		})
	})
}

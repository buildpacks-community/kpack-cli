// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import_test

import (
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/registry/registryfakes"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sclevine/spec"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	importpkg "github.com/vmware-tanzu/kpack-cli/pkg/import"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
)

func TestImportDiffer(t *testing.T) {
	spec.Run(t, "TestImportDiffer", testImportDiffer)
}

type FakeRefGetter struct{}

func NewFakeRefGetter() *FakeRefGetter {
	return &FakeRefGetter{}
}

func (rg *FakeRefGetter) RelocatedBuildpackage(keychain authn.Keychain, kpConfig config.KpConfig, image string) (string, error) {
	return image, nil
}

func (rg *FakeRefGetter) RelocatedBuildImage(keychain authn.Keychain, kpConfig config.KpConfig, image string) (string, error) {
	return image, nil
}

func (rg *FakeRefGetter) RelocatedRunImage(keychain authn.Keychain, kpConfig config.KpConfig, image string) (string, error) {
	return image, nil
}

func (rg *FakeRefGetter) RelocatedLifecycleImage(keychain authn.Keychain, kpConfig config.KpConfig, image string) (string, error) {
	return image, nil
}

func testImportDiffer(t *testing.T, when spec.G, it spec.S) {
	fakeDiffer := &fakes.FakeDiffer{DiffResult: "some-diff"}
	fakeRefGetter := NewFakeRefGetter()
	kpConfig := config.NewKpConfig("my-cool-repo", corev1.ObjectReference{})

	importDiffer := importpkg.ImportDiffer{
		Differ:             fakeDiffer,
		StoreRefGetter:     fakeRefGetter,
		StackRefGetter:     fakeRefGetter,
		LifecycleRefGetter: fakeRefGetter,
	}
	fakeKeychain := &registryfakes.FakeKeychain{}

	when("DiffClusterStore", func() {
		oldStore := &v1alpha1.ClusterStore{
			ObjectMeta: metav1.ObjectMeta{
				Name: "some-store",
			},
			Spec: v1alpha1.ClusterStoreSpec{
				Sources: []v1alpha1.StoreImage{
					{Image: "some-old-buildpackage"},
					{Image: "some-same-buildpackage"},
					{Image: "some-extra-buildpackage"},
				},
			},
		}
		newStore := importpkg.ClusterStore{
			Name: "some-store",
			Sources: []importpkg.Source{
				{Image: "some-new-buildpackage"},
				{Image: "some-extra-buildpackage"},
			},
		}

		it("returns a diff of only new store images", func() {
			diff, err := importDiffer.DiffClusterStore(fakeKeychain, kpConfig, oldStore, newStore)
			require.NoError(t, err)
			require.Equal(t, "some-diff", diff)
			diffArg0, diffArg1 := fakeDiffer.Args()
			expectedArg0 := "Name: some-store\nSources:"
			expectedArg1 := importpkg.ClusterStore{Name: "some-store", Sources: []importpkg.Source{{Image: "some-new-buildpackage"}}}
			require.Equal(t, expectedArg0, diffArg0)
			require.Equal(t, expectedArg1, diffArg1)
		})

		it("diffs with empty string when old cluster store does not exist", func() {
			diff, err := importDiffer.DiffClusterStore(fakeKeychain, kpConfig, nil, newStore)
			require.NoError(t, err)
			require.Equal(t, "some-diff", diff)
			diffArg0, _ := fakeDiffer.Args()
			require.Equal(t, "", diffArg0)
		})

		it("returns no diff with no new buildpackages", func() {
			oldStore.Spec.Sources = []v1alpha1.StoreImage{
				{Image: "some-new-buildpackage"},
				{Image: "some-extra-buildpackage"},
			}

			importDiffer.Differ = commands.Differ{}
			diff, err := importDiffer.DiffClusterStore(fakeKeychain, kpConfig, oldStore, newStore)
			require.NoError(t, err)
			require.Equal(t, "", diff)
		})
	})

	when("DiffClusterStack", func() {
		it("returns a diff of old and new cluster stack", func() {
			oldStack := &v1alpha1.ClusterStack{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-stack",
				},
				Spec: v1alpha1.ClusterStackSpec{
					Id: "some-id",
					BuildImage: v1alpha1.ClusterStackSpecImage{
						Image: "some-build-image",
					},
					RunImage: v1alpha1.ClusterStackSpecImage{
						Image: "some-run-image",
					},
				},
			}
			newStack := importpkg.ClusterStack{
				Name:       "some-stack",
				BuildImage: importpkg.Source{Image: "some-new-build-image"},
				RunImage:   importpkg.Source{Image: "some-new-run-image"},
			}

			diff, err := importDiffer.DiffClusterStack(fakeKeychain, kpConfig, oldStack, newStack)
			require.NoError(t, err)
			require.Equal(t, "some-diff", diff)
			diffArg0, diffArg1 := fakeDiffer.Args()
			expectedArg0 := importpkg.ClusterStack{
				Name:       "some-stack",
				BuildImage: importpkg.Source{Image: "some-build-image"},
				RunImage:   importpkg.Source{Image: "some-run-image"},
			}
			require.Equal(t, expectedArg0, diffArg0)
			require.Equal(t, newStack, diffArg1)
		})

		it("diffs against nil when old cluster stack does not exist", func() {
			newStack := importpkg.ClusterStack{}

			diff, err := importDiffer.DiffClusterStack(fakeKeychain, kpConfig, nil, newStack)
			require.NoError(t, err)
			require.Equal(t, "some-diff", diff)
			diffArg0, _ := fakeDiffer.Args()
			require.Equal(t, nil, diffArg0)
		})
	})

	when("DiffClusterBuilder", func() {
		it("returns a diff of old and new cluster builder", func() {
			oldBuilder := &v1alpha1.ClusterBuilder{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-builder",
				},
				Spec: v1alpha1.ClusterBuilderSpec{
					BuilderSpec: v1alpha1.BuilderSpec{
						Store: corev1.ObjectReference{
							Name: "some-store",
						},
						Stack: corev1.ObjectReference{
							Name: "some-stack",
						},
						Order: []v1alpha1.OrderEntry{{Group: []v1alpha1.BuildpackRef{{BuildpackInfo: v1alpha1.BuildpackInfo{Id: "some-buildpack"}}}}},
					},
				},
			}
			newBuilder := importpkg.ClusterBuilder{
				Name:         "some-builder",
				ClusterStore: "some-new-store",
				ClusterStack: "some-new-stack",
				Order:        []v1alpha1.OrderEntry{{Group: []v1alpha1.BuildpackRef{{BuildpackInfo: v1alpha1.BuildpackInfo{Id: "some-new-buildpack"}}}}},
			}

			diff, err := importDiffer.DiffClusterBuilder(oldBuilder, newBuilder)
			require.NoError(t, err)
			require.Equal(t, "some-diff", diff)
			diffArg0, diffArg1 := fakeDiffer.Args()
			expectedArg0 := importpkg.ClusterBuilder{
				Name:         "some-builder",
				ClusterStore: "some-store",
				ClusterStack: "some-stack",
				Order:        []v1alpha1.OrderEntry{{Group: []v1alpha1.BuildpackRef{{BuildpackInfo: v1alpha1.BuildpackInfo{Id: "some-buildpack"}}}}},
			}
			require.Equal(t, expectedArg0, diffArg0)
			require.Equal(t, newBuilder, diffArg1)
		})

		it("diffs against nil when old cluster builder does not exist", func() {
			newBuilder := importpkg.ClusterBuilder{}

			diff, err := importDiffer.DiffClusterBuilder(nil, newBuilder)
			require.NoError(t, err)
			require.Equal(t, "some-diff", diff)
			diffArg0, _ := fakeDiffer.Args()
			require.Equal(t, nil, diffArg0)
		})
	})

	when("DiffLifecycle", func() {
		it("diffs lifecycle", func() {
			oldLifecycle := "my-cool-repo/lifecycle@sha256:some-digest"
			newLifecycle := "someregistry/lifecycle@sha256:some-new-digest"
			diff, err := importDiffer.DiffLifecycle(fakeKeychain, kpConfig, oldLifecycle, newLifecycle)
			require.NoError(t, err)
			require.Equal(t, "some-diff", diff)

			diffArg0, diffArg1 := fakeDiffer.Args()
			require.Equal(t, oldLifecycle, diffArg0)
			require.Equal(t, "my-cool-repo/lifecycle@sha256:some-new-digest", diffArg1)
		})
	})
}

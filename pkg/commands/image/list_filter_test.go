// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilter(t *testing.T) {
	spec.Run(t, "TestFilter", testFilter)
}

func testFilter(t *testing.T, when spec.G, it spec.S) {
	images := &v1alpha2.ImageList{
		Items: []v1alpha2.Image{
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-image-1",
					Namespace: "some-namespace",
				},
				Spec: v1alpha2.ImageSpec{
					Builder: corev1.ObjectReference{
						Kind: v1alpha2.BuilderKind,
						Name: "some-builder",
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-image-2",
					Namespace: "some-namespace",
				},
				Spec: v1alpha2.ImageSpec{
					Builder: corev1.ObjectReference{
						Kind: v1alpha2.ClusterBuilderKind,
						Name: "some-cluster-builder",
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-image-3",
					Namespace: "some-namespace",
				},
				Status: v1alpha2.ImageStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-image-4",
					Namespace: "some-namespace",
				},
				Status: v1alpha2.ImageStatus{
					LatestBuildReason: "COMMIT,BUILDPACK",
				},
			},
		},
	}

	when("the builder filter is specified", func() {
		it("filters images", func() {
			imgs, err := filterImageList(images, []string{"builder=some-builder"})
			require.NoError(t, err)

			require.Len(t, imgs.Items, 1)
			require.Equal(t, "test-image-1", imgs.Items[0].ObjectMeta.Name)
		})
	})

	when("the clusterbuilder filter is specified", func() {
		it("filters images", func() {
			imgs, err := filterImageList(images, []string{"clusterbuilder=some-cluster-builder"})
			require.NoError(t, err)

			require.Len(t, imgs.Items, 1)
			require.Equal(t, "test-image-2", imgs.Items[0].ObjectMeta.Name)
		})
	})

	when("the status filter is specified", func() {
		it("filters images", func() {
			imgs, err := filterImageList(images, []string{"ready=true,some-other-status"})
			require.NoError(t, err)

			require.Len(t, imgs.Items, 1)
			require.Equal(t, "test-image-3", imgs.Items[0].ObjectMeta.Name)
		})
	})

	when("the latest-reason filter is specified", func() {
		it("filters images", func() {
			imgs, err := filterImageList(images, []string{"latest-reason=commit,some-other-build-reason"})
			require.NoError(t, err)

			require.Len(t, imgs.Items, 1)
			require.Equal(t, "test-image-4", imgs.Items[0].ObjectMeta.Name)
		})
	})

	imagesWithSameBuilder := &v1alpha2.ImageList{
		Items: []v1alpha2.Image{
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-image-1",
					Namespace: "some-namespace",
				},
				Spec: v1alpha2.ImageSpec{
					Builder: corev1.ObjectReference{
						Kind: v1alpha2.BuilderKind,
						Name: "some-builder",
					},
				},
				Status: v1alpha2.ImageStatus{
					LatestBuildReason: "COMMIT",
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "some-ignored-test-image",
					Namespace: "some-namespace",
				},
				Spec: v1alpha2.ImageSpec{
					Builder: corev1.ObjectReference{
						Kind: v1alpha2.BuilderKind,
						Name: "some-builder",
					},
				},
				Status: v1alpha2.ImageStatus{
					LatestBuildReason: "some-other-build-reason",
				},
			},
		},
	}

	when("multiple filters are specified", func() {
		it("filters images matching all criteria", func() {
			imgs, err := filterImageList(imagesWithSameBuilder, []string{"builder=some-builder", "latest-reason=commit"})
			require.NoError(t, err)

			require.Len(t, imgs.Items, 1)
			require.Equal(t, "test-image-1", imgs.Items[0].ObjectMeta.Name)
		})
	})

	when("an invalid filter is specified", func() {
		it("returns a helpful error message", func() {
			_, err := filterImageList(imagesWithSameBuilder, []string{"some-invalid-filter=some-value"})
			require.Error(t, err, "invalid filter argument \"some-invalid-filter=some-value\"")
		})
	})

}

// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/sclevine/spec"
)

func TestFilter(t *testing.T) {
	spec.Run(t, "TestFilter", testFilter)
}

func testFilter(t *testing.T, when spec.G, it spec.S) {
	images := &v1alpha1.ImageList{
		Items: []v1alpha1.Image{
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-image-1",
					Namespace: "some-namespace",
				},
				Spec: v1alpha1.ImageSpec{
					Builder: corev1.ObjectReference{
						Kind:            v1alpha1.BuilderKind,
						Name:            "some-builder",
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-image-2",
					Namespace: "some-namespace",
				},
				Spec: v1alpha1.ImageSpec{
					Builder: corev1.ObjectReference{
						Kind:            v1alpha1.ClusterBuilderKind,
						Name:            "some-cluster-builder",
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-image-3",
					Namespace: "some-namespace",
				},
				Status: v1alpha1.ImageStatus{
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
				Status: v1alpha1.ImageStatus{
					LatestBuildReason: "COMMIT,BUILDPACK",
				},
			},
		},
	}

	when("the builder filter is specified", func() {
		it("filters images", func() {
			imgs := filterImageList(images, []string{"builder=some-builder"})
			require.Len(t, imgs.Items, 1)
			require.Equal(t, "test-image-1", imgs.Items[0].ObjectMeta.Name)
		})
	})

	when("the clusterbuilder filter is specified", func() {
		it("filters images", func() {
			imgs := filterImageList(images, []string{"clusterbuilder=some-cluster-builder"})
			require.Len(t, imgs.Items, 1)
			require.Equal(t, "test-image-2", imgs.Items[0].ObjectMeta.Name)
		})
	})

	when("the status filter is specified", func() {
		it("filters images", func() {
			imgs := filterImageList(images, []string{"ready=true,some-other-status"})
			require.Len(t, imgs.Items, 1)
			require.Equal(t, "test-image-3", imgs.Items[0].ObjectMeta.Name)
		})
	})

	when("the latest-reason filter is specified", func() {
		it("filters images", func() {
			imgs := filterImageList(images, []string{"latest-reason=commit,some-other-build-reason"})
			require.Len(t, imgs.Items, 1)
			require.Equal(t, "test-image-4", imgs.Items[0].ObjectMeta.Name)
		})
	})

	imagesWithSameBuilder := &v1alpha1.ImageList{
		Items: []v1alpha1.Image{
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-image-1",
					Namespace: "some-namespace",
				},
				Spec: v1alpha1.ImageSpec{
					Builder: corev1.ObjectReference{
						Kind:            v1alpha1.BuilderKind,
						Name:            "some-builder",
					},
				},
				Status: v1alpha1.ImageStatus{
					LatestBuildReason: "COMMIT",
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "some-ignored-test-image",
					Namespace: "some-namespace",
				},
				Spec: v1alpha1.ImageSpec{
					Builder: corev1.ObjectReference{
						Kind:            v1alpha1.BuilderKind,
						Name:            "some-builder",
					},
				},
				Status: v1alpha1.ImageStatus{
					LatestBuildReason: "some-other-build-reason",
				},
			},
		},
	}

	when("multiple filters are specified", func() {
		it("filters images matching all criteria", func() {
			imgs := filterImageList(imagesWithSameBuilder, []string{"builder=some-builder", "latest-reason=commit"})
			require.Len(t, imgs.Items, 1)
			require.Equal(t, "test-image-1", imgs.Items[0].ObjectMeta.Name)
		})
	})

}

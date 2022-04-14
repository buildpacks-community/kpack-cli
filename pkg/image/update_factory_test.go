// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"io/ioutil"
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/image"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
)

func TestPatchFactory(t *testing.T) {
	spec.Run(t, "TestPatchFactory", testPatchFactory)
}

func testPatchFactory(t *testing.T, when spec.G, it spec.S) {
	var (
		factory = image.Factory{
			SourceUploader: fakes.NewFakeSourceUploader(ioutil.Discard, true),
		}

		img = &v1alpha2.Image{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-image",
				Namespace: "some-namespace",
			},
			Spec: v1alpha2.ImageSpec{
				Tag: "some-tag",
				Builder: corev1.ObjectReference{
					Kind: v1alpha2.ClusterBuilderKind,
					Name: "some-ccb",
				},
				ServiceAccountName: "some-service-account",
				Source: corev1alpha1.SourceConfig{
					Blob: &corev1alpha1.Blob{
						URL: "some-blob-url",
					},
					SubPath: "some-sub-path",
				},
				Build: &v1alpha2.ImageBuild{
					Env: []corev1.EnvVar{
						{
							Name:  "foo",
							Value: "",
						},
					},
				},
				AdditionalTags: []string{"some-other-tag"},
			},
		}

		expectedImg *v1alpha2.Image
	)

	it.Before(func() {
		expectedImg = img.DeepCopy()
	})

	it("defaults the git revision to main", func() {
		factory.GitRepo = "some-repo"

		expectedImg.Spec.Source = corev1alpha1.SourceConfig{
			Git: &corev1alpha1.Git{
				URL:      "some-repo",
				Revision: "main",
			},
			SubPath: "some-sub-path",
		}

		updatedImage, err := factory.UpdateImage(img)
		require.NoError(t, err)
		require.Equal(t, expectedImg, updatedImage)
	})

	it("adds and removes service accounts", func() {
		factory.AdditionalTags = []string{"some-new-tag"}
		factory.DeleteAdditionalTags = []string{"some-other-tag"}

		expectedImg.Spec.AdditionalTags = []string{"some-new-tag"}

		updatedImage, err := factory.UpdateImage(img)
		require.NoError(t, err)
		require.Equal(t, expectedImg, updatedImage)
	})

	when("too many source types are provided", func() {
		it("returns an error message", func() {
			factory.GitRepo = "some-git-repo"
			factory.Blob = "some-blob"
			factory.LocalPath = "some-local-path"
			_, err := factory.UpdateImage(img)
			require.EqualError(t, err, "image source must be one of git, blob, or local-path")
		})
	})

	when("git revision is provided with non-git source types", func() {
		it("returns an error message", func() {
			factory.Blob = "some-blob"
			factory.GitRevision = "some-revision"
			_, err := factory.UpdateImage(img)
			require.EqualError(t, err, "git-revision is incompatible with blob and local path image sources")
		})
	})

	when("git revision is provided with an existing non-git source types", func() {
		it("returns an error message", func() {
			factory.GitRevision = "some-revision"
			_, err := factory.UpdateImage(img)
			require.EqualError(t, err, "git-revision is incompatible with existing image source")
		})
	})

	when("both builder and cluster builder are provided", func() {
		it("returns an error message", func() {
			factory.Builder = "some-builder"
			factory.ClusterBuilder = "some-cluster-builder"
			_, err := factory.UpdateImage(img)
			require.EqualError(t, err, "must provide one of builder or cluster-builder")
		})
	})

	when("delete-env and env have the same key", func() {
		it("returns an error message", func() {
			factory.DeleteEnv = []string{"foo"}
			factory.Env = []string{"foo=bar"}
			_, err := factory.UpdateImage(img)
			require.EqualError(t, err, "duplicate delete-env and env-var parameter 'foo'")
		})
	})

	when("delete-env does not exist in the current image", func() {
		it("returns an error message", func() {
			factory.DeleteEnv = []string{"bar"}
			_, err := factory.UpdateImage(img)
			require.EqualError(t, err, "delete-env parameter 'bar' not found in existing image configuration")
		})
	})

	when("an AdditionalTag already exists", func() {
		it("does not re-add it", func() {
			factory.AdditionalTags = []string{"some-other-tag"}
			updatedImage, err := factory.UpdateImage(img)
			require.NoError(t, err)
			require.Equal(t, img, updatedImage)
		})
	})

	when("DeleteAdditionalTags and AdditionalTags have the same value", func() {
		it("returns an error message", func() {
			factory.DeleteAdditionalTags = []string{"some-other-tag"}
			factory.AdditionalTags = []string{"some-other-tag"}
			_, err := factory.UpdateImage(img)
			require.EqualError(t, err, "duplicate delete-additional-tag and additional-tag parameter 'some-other-tag'")
		})
	})

	when("delete-additional-tag does not exist in the current image", func() {
		it("returns an error message", func() {
			factory.DeleteAdditionalTags = []string{"bar"}
			_, err := factory.UpdateImage(img)
			require.EqualError(t, err, "delete-additional-tag parameter 'bar' not found in existing image additional tags")
		})
	})

	when("the image.spec.build is nil", func() {
		it("does not panic", func() {
			img.Spec.Build = nil
			_, err := factory.UpdateImage(img)
			require.NoError(t, err)
		})
	})

	when("an env var has an equal sign in the value", func() {
		it("handles the env var", func() {
			factory.Env = append(factory.Env, `BP_MAVEN_BUILD_ARGUMENTS="-Dmaven.test.skip=true -Pk8s package"`)

			expectedImg.Spec.Build.Env = append(expectedImg.Spec.Build.Env, corev1.EnvVar{
				Name:  "BP_MAVEN_BUILD_ARGUMENTS",
				Value: `"-Dmaven.test.skip=true -Pk8s package"`,
			})

			updatedImg, err := factory.UpdateImage(img)
			require.NoError(t, err)
			require.Equal(t, expectedImg, updatedImg)
		})
	})

	when("patching cache size", func() {
		it("can set a new cache size", func() {
			factory.CacheSize = "3G"

			s := resource.MustParse("3G")
			expectedImg.Spec.Cache = &v1alpha2.ImageCacheConfig{
				Volume: &v1alpha2.ImagePersistentVolumeCache{
					Size: &s,
				},
			}

			updatedImg, err := factory.UpdateImage(img)
			require.NoError(t, err)
			require.Equal(t, expectedImg, updatedImg)
		})

		it("errors if cache size is decreased", func() {
			cacheSize := resource.MustParse("2G")
			img.Spec.Cache = &v1alpha2.ImageCacheConfig{
				Volume: &v1alpha2.ImagePersistentVolumeCache{
					Size: &cacheSize,
				},
			}

			factory.CacheSize = "1G"
			_, err := factory.UpdateImage(img)
			require.EqualError(t, err, "cache size cannot be decreased, current: 2G, requested: 1G")
		})

		it("errors if cache size is invalid", func() {
			factory.CacheSize = "invalid"
			_, err := factory.UpdateImage(img)
			require.EqualError(t, err, "invalid cache size, must be valid quantity ex. 2G")
		})
	})
}

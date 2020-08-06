// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"testing"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/image"
	srcfakes "github.com/pivotal/build-service-cli/pkg/source/fakes"
)

func TestPatchFactory(t *testing.T) {
	spec.Run(t, "TestPatchFactory", testPatchFactory)
}

func testPatchFactory(t *testing.T, when spec.G, it spec.S) {
	uploader := &srcfakes.SourceUploader{
		ImageRef: "",
	}

	factory := image.PatchFactory{
		SourceUploader: uploader,
	}

	img := &v1alpha1.Image{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-image",
			Namespace: "some-namespace",
		},
		Spec: v1alpha1.ImageSpec{
			Tag: "some-tag",
			Builder: corev1.ObjectReference{
				Kind: v1alpha1.ClusterBuilderKind,
				Name: "some-ccb",
			},
			ServiceAccount: "some-service-account",
			Source: v1alpha1.SourceConfig{
				Blob: &v1alpha1.Blob{
					URL: "some-blob-url",
				},
				SubPath: "some-sub-path",
			},
			Build: &v1alpha1.ImageBuild{
				Env: []corev1.EnvVar{
					{
						Name:  "foo",
						Value: "",
					},
				},
			},
		},
	}

	when("too many source types are provided", func() {
		it("returns an error message", func() {
			factory.GitRepo = "some-git-repo"
			factory.Blob = "some-blob"
			factory.LocalPath = "some-local-path"
			_, err := factory.MakePatch(img)
			require.EqualError(t, err, "image source must be one of git, blob, or local-path")
		})
	})

	when("git revision is provided with non-git source types", func() {
		it("returns an error message", func() {
			factory.Blob = "some-blob"
			factory.GitRevision = "some-revision"
			_, err := factory.MakePatch(img)
			require.EqualError(t, err, "git-revision is incompatible with blob and local path image sources")
		})
	})

	when("git revision is provided with an existing non-git source types", func() {
		it("returns an error message", func() {
			factory.GitRevision = "some-revision"
			_, err := factory.MakePatch(img)
			require.EqualError(t, err, "git-revision is incompatible with existing image source")
		})
	})

	when("both builder and cluster builder are provided", func() {
		it("returns an error message", func() {
			factory.Builder = "some-builder"
			factory.ClusterBuilder = "some-cluster-builder"
			_, err := factory.MakePatch(img)
			require.EqualError(t, err, "must provide one of builder or cluster-builder")
		})
	})

	when("delete-env and env have the same key", func() {
		it("returns an error message", func() {
			factory.DeleteEnv = []string{"foo"}
			factory.Env = []string{"foo=bar"}
			_, err := factory.MakePatch(img)
			require.EqualError(t, err, "duplicate delete-env and env-var parameter 'foo'")
		})
	})

	when("delete-env does not exist in the current image", func() {
		it("returns an error message", func() {
			factory.DeleteEnv = []string{"bar"}
			_, err := factory.MakePatch(img)
			require.EqualError(t, err, "delete-env parameter 'bar' not found in existing image configuration")
		})
	})

	when("the image.spec.build is nil", func() {
		it("does not panic", func() {
			img.Spec.Build = nil
			_, err := factory.MakePatch(img)
			require.NoError(t, err)
		})
	})

	when("an env var has an equal sign in the value", func() {
		it("handles the env var", func() {
			factory.Env = append(factory.Env, `BP_MAVEN_BUILD_ARGUMENTS="-Dmaven.test.skip=true -Pk8s package"`)
			patch, err := factory.MakePatch(img)
			require.NoError(t, err)
			require.Equal(t, `{"spec":{"build":{"env":[{"name":"foo"},{"name":"BP_MAVEN_BUILD_ARGUMENTS","value":"\"-Dmaven.test.skip=true -Pk8s package\""}]}}}`, string(patch))
		})
	})
}

// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder_test

import (
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/fake"
	buildv1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/buildpacks-community/kpack-cli/pkg/builder"
	"github.com/buildpacks-community/kpack-cli/pkg/testhelpers"
)

func TestReadOrderFromImage(t *testing.T) {
	spec.Run(t, "TestReadOrderFromImage", testReadOrderFromImage)
}

func testReadOrderFromImage(t *testing.T, when spec.G, it spec.S) {
	var (
		fakeFetcher *testhelpers.FakeFetcher
		keychain    authn.Keychain
		imageRef    = "some-registry.io/builder:tag"
	)

	it.Before(func() {
		fakeFetcher = &testhelpers.FakeFetcher{}
		keychain = authn.DefaultKeychain
	})

	when("the image contains a valid order label", func() {
		it("extracts the order successfully", func() {
			orderJSON := `[
				{
					"group": [
						{
							"id": "org.cloudfoundry.nodejs",
							"version": "1.2.3"
						},
						{
							"id": "org.cloudfoundry.npm",
							"version": "4.5.6",
							"optional": true
						}
					]
				},
				{
					"group": [
						{
							"id": "org.cloudfoundry.go"
						}
					]
				}
			]`

			fakeImage := &fake.FakeImage{}
			fakeImage.ConfigFileReturns(&v1.ConfigFile{
				Config: v1.Config{
					Labels: map[string]string{
						"io.buildpacks.buildpack.order": orderJSON,
					},
				},
			}, nil)

			fakeFetcher.SetImage(imageRef, fakeImage)

			order, err := builder.ReadOrderFromImage(keychain, fakeFetcher, imageRef)
			require.NoError(t, err)

			expectedOrder := []buildv1alpha2.BuilderOrderEntry{
				{
					Group: []buildv1alpha2.BuilderBuildpackRef{
						{
							BuildpackRef: corev1alpha1.BuildpackRef{
								BuildpackInfo: corev1alpha1.BuildpackInfo{
									Id:      "org.cloudfoundry.nodejs",
									Version: "1.2.3",
								},
								Optional: false,
							},
						},
						{
							BuildpackRef: corev1alpha1.BuildpackRef{
								BuildpackInfo: corev1alpha1.BuildpackInfo{
									Id:      "org.cloudfoundry.npm",
									Version: "4.5.6",
								},
								Optional: true,
							},
						},
					},
				},
				{
					Group: []buildv1alpha2.BuilderBuildpackRef{
						{
							BuildpackRef: corev1alpha1.BuildpackRef{
								BuildpackInfo: corev1alpha1.BuildpackInfo{
									Id:      "org.cloudfoundry.go",
									Version: "",
								},
								Optional: false,
							},
						},
					},
				},
			}

			assert.Equal(t, expectedOrder, order)
		})
	})

	when("the image does not contain the order label", func() {
		it("returns an error", func() {
			fakeImage := &fake.FakeImage{}
			fakeImage.ConfigFileReturns(&v1.ConfigFile{
				Config: v1.Config{
					Labels: map[string]string{},
				},
			}, nil)

			fakeFetcher.SetImage(imageRef, fakeImage)

			_, err := builder.ReadOrderFromImage(keychain, fakeFetcher, imageRef)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "does not contain the io.buildpacks.buildpack.order label")
		})
	})

	when("the order label contains invalid JSON", func() {
		it("returns an error", func() {
			fakeImage := &fake.FakeImage{}
			fakeImage.ConfigFileReturns(&v1.ConfigFile{
				Config: v1.Config{
					Labels: map[string]string{
						"io.buildpacks.buildpack.order": "not valid json",
					},
				},
			}, nil)

			fakeFetcher.SetImage(imageRef, fakeImage)

			_, err := builder.ReadOrderFromImage(keychain, fakeFetcher, imageRef)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to parse io.buildpacks.buildpack.order label")
		})
	})

	when("the image cannot be fetched", func() {
		it("returns an error", func() {
			fakeFetcher.SetError("some-error")

			_, err := builder.ReadOrderFromImage(keychain, fakeFetcher, imageRef)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to fetch image")
		})
	})
}

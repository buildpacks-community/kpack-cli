// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package registry_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/pivotal/kpack/pkg/registry/registryfakes"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
)

func TestRelocateStackImages(t *testing.T) {
	spec.Run(t, "Test Relocation of Stack Images", testRelocateStackImages)
}

func testRelocateStackImages(t *testing.T, when spec.G, it spec.S) {
	var (
		fakeKeychain = &registryfakes.FakeKeychain{}
	)

	when("#Fetch", func() {
		when("remote", func() {
			it("it should fetch the image with the digest", func() {
				fetcher := registry.DefaultFetcher{}

				image, err := fetcher.Fetch(fakeKeychain, "cloudfoundry/run:tiny-cnb")
				require.NoError(t, err)
				assert.NotNil(t, image)
			})
		})
	})

	when("#Relocate", func() {
		it("should correctly relocate image to the dest registry", func() {
			dstImageName := "dest-repo/an-image"
			additionalTags := 0
			dstRegistryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodHead {
					http.Error(w, "NotFound", http.StatusNotFound)
					return
				}
				switch path := r.URL.Path; {
				case path == "/v2/":
					w.WriteHeader(http.StatusOK)
				case path == "/v2/"+dstImageName+"/blobs/uploads/":
					http.Error(w, "Mounted", http.StatusCreated)
				case path == "/v2/"+dstImageName+"/manifests/latest":
					http.Error(w, "Created", http.StatusCreated)
				case regexp.MustCompile(fmt.Sprintf("/v2/%s/manifests/\\d{14}", dstImageName)).Match([]byte(path)):
					additionalTags++
					http.Error(w, "Created", http.StatusCreated)
				default:
					t.Fatalf("Unexpected path: %v", r.URL.Path)
				}
			}))
			defer dstRegistryServer.Close()

			uri, err := url.Parse(dstRegistryServer.URL)
			require.NoError(t, err)
			dst := fmt.Sprintf("%s/%s@sha256:f55aa0bd26b801374773c103bed4479865d0e37435b848cb39d164ccb2c3ba51", uri.Host, dstImageName)

			srcImage, err := random.Image(int64(100), int64(5))
			require.NoError(t, err)
			srcImageDigest, err := srcImage.Digest()
			require.NoError(t, err)

			output := &bytes.Buffer{}
			relocator := registry.NewDefaultRelocator(output, registry.DefaultTLSConfig())
			relocatedRef, err := relocator.Relocate(fakeKeychain, srcImage, dst)
			require.NoError(t, err)

			require.Equal(t, 1, strings.Count(relocatedRef, "sha256:"))
			relocatedHex := relocatedRef[len(relocatedRef)-64:]
			require.Equal(t, srcImageDigest.Hex, relocatedHex)
			require.Equal(t, 1, additionalTags)

			require.Equal(t, output.String(), fmt.Sprintf("\tUploading '%s'", relocatedRef))
		})

		it("should error on invalid destination", func() {
			srcImage, err := random.Image(int64(100), int64(5))
			require.NoError(t, err)

			relocator := registry.NewDefaultRelocator(ioutil.Discard, registry.DefaultTLSConfig())
			_, err = relocator.Relocate(fakeKeychain, srcImage, "notuser/notimage:tag")
			require.Error(t, err)
		})
	})
}

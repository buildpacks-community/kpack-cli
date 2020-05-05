package image_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pivotal/build-service-cli/pkg/image"
)

func TestRelocateStackImages(t *testing.T) {
	spec.Run(t, "Test Relocation of Stack Images", testRelocateStackImages)
}

func testRelocateStackImages(t *testing.T, when spec.G, it spec.S) {
	when("#Fetch", func() {
		when("remote", func() {
			it("it should fetch the image with the digest", func() {
				fetcher := image.Fetcher{}

				image, err := fetcher.Fetch("cloudfoundry/run:tiny-cnb")
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

			relocator := image.Relocator{}
			relocatedRef, err := relocator.Relocate(srcImage, dst)
			require.NoError(t, err)
			require.Equal(t, 1, strings.Count(relocatedRef, "sha256:"))
			relocatedHex := relocatedRef[len(relocatedRef)-64:]
			require.Equal(t, srcImageDigest.Hex, relocatedHex)
			require.Equal(t, 1, additionalTags)
		})

		it("should error on invalid destination", func() {
			srcImage, err := random.Image(int64(100), int64(5))
			require.NoError(t, err)
			relocator := image.Relocator{}
			_, err = relocator.Relocate(srcImage, "notuser/notimage:tag")
			require.Error(t, err)
		})
	})
}

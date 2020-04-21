package registry

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"

	"github.com/pivotal/build-service-cli/pkg/archive"
)

type SourceUploader struct {
}

func (s *SourceUploader) Upload(ref, path string) (string, error) {
	if archive.IsZip(path) {
		unzipDest, err := ioutil.TempDir("", "zipdir")
		if err != nil {
			return "", err
		}
		defer os.RemoveAll(unzipDest)

		err = archive.ExtractZip(path, unzipDest)
		if err != nil {
			return "", err
		}

		path = unzipDest
	}

	tarFile, err := archive.CreateTar(path)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tarFile)

	image, err := random.Image(0, 0)
	if err != nil {
		return "", err
	}

	layer, err := tarball.LayerFromFile(tarFile)
	if err != nil {
		return "", err
	}

	image, err = mutate.AppendLayers(image, layer)
	if err != nil {
		return "", errors.Wrap(err, "adding layer")
	}

	fullRef, err := name.ParseReference(ref + ":" + fmt.Sprint(time.Now().UnixNano()))
	if err != nil {
		return "", err
	}

	err = remote.Write(fullRef, image, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		if transportError, ok := err.(*transport.Error); ok {
			if transportError.StatusCode == 401 {
				return "", errors.Errorf("invalid credentials, ensure registry credentials for '%s' are available locally", fullRef.Context().Registry)
			}
		}
		return "", err
	}

	hash, err := image.Digest()
	if err != nil {
		return "", err
	}

	return ref + "@" + hash.String(), nil
}

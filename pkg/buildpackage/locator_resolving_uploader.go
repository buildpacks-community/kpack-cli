package buildpackage

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
)

const (
	buildpackageMetadataLabel = "io.buildpacks.buildpackage.metadata"
)

type Relocator interface {
	Relocate(image v1.Image, dest string) (string, error)
}

type Fetcher interface {
	Fetch(src string) (v1.Image, error)
}

type LocatorResolvingUpdater struct {
	Relocator Relocator
	Fetcher   Fetcher
}

func (u *LocatorResolvingUpdater) Upload(repository, buildPackage string) (string, error) {
	tempDir, err := ioutil.TempDir("", "cnb-upload")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	image, err := u.read(buildPackage, tempDir)
	if err != nil {
		return "", err
	}

	type buildpackageMetadata struct {
		Id string `json:"id"`
	}

	metadata := buildpackageMetadata{}
	err = imagehelpers.GetLabel(image, buildpackageMetadataLabel, &metadata)
	if err != nil {
		return "", err
	}

	return u.Relocator.Relocate(image, path.Join(repository, strings.ReplaceAll(metadata.Id, "/", "_")))
}

func (u *LocatorResolvingUpdater) read(buildPackage, tempDir string) (v1.Image, error) {
	if isLocalCnb(buildPackage) {
		cnb, err := readCNB(buildPackage, tempDir)
		return cnb, errors.Wrapf(err, "invalid local buildpackage %s", buildPackage)
	}
	return u.Fetcher.Fetch(buildPackage)
}

func isLocalCnb(buildPackage string) bool {
	_, err := os.Stat(buildPackage)
	return err == nil
}

func readCNB(buildPackage, tempDir string) (v1.Image, error) {

	cnbFile, err := os.Open(buildPackage)
	if err != nil {
		return nil, err
	}

	err = extractTar(cnbFile, tempDir)
	if err != nil {
		return nil, err
	}

	index, err := layout.ImageIndexFromPath(tempDir)
	if err != nil {
		return nil, err
	}

	manifest, err := index.IndexManifest()
	if err != nil {
		return nil, err
	}

	image, err := index.Image(manifest.Manifests[0].Digest)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func extractTar(reader io.Reader, dir string) error {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		filePath := filepath.Join(dir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err := os.MkdirAll(filePath, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
		}
	}
	return nil
}

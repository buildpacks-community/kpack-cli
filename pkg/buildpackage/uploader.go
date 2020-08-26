// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package buildpackage

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/pkg/errors"

	"github.com/pivotal/build-service-cli/pkg/archive"
)

const (
	buildpackageMetadataLabel = "io.buildpacks.buildpackage.metadata"
)

type Relocator interface {
	Relocate(writer io.Writer, image v1.Image, dest string) (string, error)
}

type Fetcher interface {
	Fetch(src string) (v1.Image, error)
}

type Uploader struct {
	Relocator Relocator
	Fetcher   Fetcher
}

func (u *Uploader) UploadBuildpackage(writer io.Writer, repository, buildPackage string) (string, error) {
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

	return u.Relocator.Relocate(writer, image, path.Join(repository, strings.ReplaceAll(metadata.Id, "/", "_")))
}

func (u *Uploader) read(buildPackage, tempDir string) (v1.Image, error) {
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

	err = archive.ReadTar(cnbFile, tempDir)
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

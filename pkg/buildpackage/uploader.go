// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package buildpackage

import (
	"io/ioutil"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/pkg/errors"

	"github.com/buildpacks-community/kpack-cli/pkg/archive"
)

type Relocator interface {
	Relocate(keychain authn.Keychain, image v1.Image, dest string) (string, error)
}

type Fetcher interface {
	Fetch(keychain authn.Keychain, image string) (v1.Image, error)
}

type Uploader struct {
	Relocator Relocator
	Fetcher   Fetcher
}

func (u *Uploader) UploadBuildpackage(keychain authn.Keychain, buildPackage, repository string) (string, error) {
	tempDir, err := ioutil.TempDir("", "cnb-upload")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	image, err := u.read(keychain, buildPackage, tempDir)
	if err != nil {
		return "", err
	}

	return u.Relocator.Relocate(keychain, image, repository)
}

func (u *Uploader) read(keychain authn.Keychain, buildPackage, tempDir string) (v1.Image, error) {
	if isLocalCnb(buildPackage) {
		cnb, err := readCNB(buildPackage, tempDir)
		return cnb, errors.Wrapf(err, "invalid local buildpackage %s", buildPackage)
	}
	return u.Fetcher.Fetch(keychain, buildPackage)
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

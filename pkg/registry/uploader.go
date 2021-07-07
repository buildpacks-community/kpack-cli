// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/kpack-cli/pkg/archive"
)

type SourceUploader interface {
	Upload(keychain authn.Keychain, dstImgRefStr, srcPath string) (string, error)
}

type DefaultSourceUploader struct {
	Relocator Relocator
}

func (d DefaultSourceUploader) Upload(keychain authn.Keychain, dstImgRefStr, srcPath string) (string, error) {
	image, err := writeLocalPathToImage(srcPath)
	if err != nil {
		return "", err
	}

	return d.Relocator.Relocate(keychain, image, dstImgRefStr)
}

func writeLocalPathToImage(path string) (v1.Image, error) {
	var srcTarPath string
	var err error

	srcTarPath, err = readPathToTar(path)
	defer os.Remove(srcTarPath)

	if err != nil {
		return nil, err
	}

	image, err := random.Image(0, 0)
	if err != nil {
		return image, err
	}

	layer, err := tarball.LayerFromFile(srcTarPath)
	if err != nil {
		return image, err
	}

	image, err = mutate.AppendLayers(image, layer)
	if err != nil {
		return image, err
	}

	return image, nil
}

func readPathToTar(path string) (string, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if !fi.IsDir() && archive.IsZip(path) {
		return archive.ZipToTar(path)
	} else if !fi.IsDir() {
		return "", errors.New("local path must be a directory or zip")
	}

	return archive.CreateTar(path)
}

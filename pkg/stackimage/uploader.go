// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package stackimage

import (
	"fmt"
	"io"
	"path/filepath"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pkg/errors"

	"github.com/pivotal/build-service-cli/pkg/registry"
)

const (
	IdLabel        = "io.buildpacks.stack.id"
	RunImageName   = "run"
	BuildImageName = "build"
)

type Relocator interface {
	Relocate(image v1.Image, dest string, writer io.Writer, tlsCfg registry.TLSConfig) (string, error)
}

type Fetcher interface {
	Fetch(src string, tlsCfg registry.TLSConfig) (v1.Image, error)
}

type Uploader struct {
	Relocator Relocator
	Fetcher   Fetcher
}

func (u *Uploader) UploadStackImages(buildImageTag, runImageTag, dest string, tlsCfg registry.TLSConfig, writer io.Writer) (string, string, error) {
	buildImage, err := u.Fetcher.Fetch(buildImageTag, tlsCfg)
	if err != nil {
		return "", "", err
	}

	runImage, err := u.Fetcher.Fetch(runImageTag, tlsCfg)
	if err != nil {
		return "", "", err
	}

	relocatedBuildImageRef, err := u.Relocator.Relocate(buildImage, filepath.Join(dest, BuildImageName), writer, tlsCfg)
	if err != nil {
		return "", "", err
	}

	relocatedRunImageRef, err := u.Relocator.Relocate(runImage, filepath.Join(dest, RunImageName), writer, tlsCfg)
	if err != nil {
		return "", "", err
	}

	return relocatedBuildImageRef, relocatedRunImageRef, nil
}

func (u *Uploader) ValidateStackIDs(buildImageTag, runImageTag string, tlsCfg registry.TLSConfig) (string, error) {
	buildImage, err := u.Fetcher.Fetch(buildImageTag, tlsCfg)
	if err != nil {
		return "", err
	}

	buildStackId, err := getStackId(buildImage)
	if err != nil {
		return "", err
	}

	runImage, err := u.Fetcher.Fetch(runImageTag, tlsCfg)
	if err != nil {
		return "", err
	}

	runStackId, err := getStackId(runImage)
	if err != nil {
		return "", err
	}

	if buildStackId != runStackId {
		return "", errors.Errorf("build stack '%s' does not match run stack '%s'", buildStackId, runStackId)
	}

	return buildStackId, nil
}

func (u *Uploader) UploadedBuildImageRef(imageTag, dest string, tlsCfg registry.TLSConfig) (string, error) {
	image, err := u.Fetcher.Fetch(imageTag, tlsCfg)
	if err != nil {
		return "", err
	}

	digest, err := image.Digest()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s@%s", filepath.Join(dest, BuildImageName), digest.String()), nil
}

func (u *Uploader) UploadedRunImageRef(imageTag, dest string, tlsCfg registry.TLSConfig) (string, error) {
	image, err := u.Fetcher.Fetch(imageTag, tlsCfg)
	if err != nil {
		return "", err
	}

	digest, err := image.Digest()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s@%s", filepath.Join(dest, RunImageName), digest.String()), nil
}

func getStackId(img v1.Image) (string, error) {
	config, err := img.ConfigFile()
	if err != nil {
		return "", err
	}

	labels := config.Config.Labels

	id, ok := labels[IdLabel]
	if !ok {
		return "", errors.New("invalid stack image")
	}

	return id, nil
}

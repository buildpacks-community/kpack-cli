// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package stackimage

import (
	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pkg/errors"
)

const (
	IdLabel = "io.buildpacks.stack.id"
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

func (u *Uploader) UploadStackImages(keychain authn.Keychain, buildImageTag, runImageTag, dest string) (string, string, error) {
	buildImage, err := u.Fetcher.Fetch(keychain, buildImageTag)
	if err != nil {
		return "", "", err
	}

	runImage, err := u.Fetcher.Fetch(keychain, runImageTag)
	if err != nil {
		return "", "", err
	}

	relocatedBuildImageRef, err := u.Relocator.Relocate(keychain, buildImage, dest)
	if err != nil {
		return "", "", err
	}

	relocatedRunImageRef, err := u.Relocator.Relocate(keychain, runImage, dest)
	if err != nil {
		return "", "", err
	}

	return relocatedBuildImageRef, relocatedRunImageRef, nil
}

func (u *Uploader) ValidateStackIDs(keychain authn.Keychain, buildImageTag, runImageTag string) (string, error) {
	buildImage, err := u.Fetcher.Fetch(keychain, buildImageTag)
	if err != nil {
		return "", err
	}

	buildStackId, err := getStackId(buildImage)
	if err != nil {
		return "", err
	}

	runImage, err := u.Fetcher.Fetch(keychain, runImageTag)
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

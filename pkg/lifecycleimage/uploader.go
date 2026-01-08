// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package lifecycleimage

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
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

const (
	lifecycleVersionLabel = "io.buildpacks.lifecycle.version"
	lifecycleApisLabel    = "io.buildpacks.lifecycle.apis"
)

func (u *Uploader) UploadLifecycleImage(keychain authn.Keychain, imageTag, dest string) (string, error) {
	image, err := u.Fetcher.Fetch(keychain, imageTag)
	if err != nil {
		return "", err
	}

	return u.Relocator.Relocate(keychain, image, dest)
}

func (u *Uploader) ValidateLifecycleImage(keychain authn.Keychain, imageTag string) error {
	buildImage, err := u.Fetcher.Fetch(keychain, imageTag)
	if err != nil {
		return err
	}

	hasVersionLabel, err := imagehelpers.HasLabel(buildImage, lifecycleVersionLabel)
	if err != nil {
		return fmt.Errorf("could not get label %s: %w", lifecycleVersionLabel, err)
	}
	if !hasVersionLabel {
		return fmt.Errorf("missing label %s", lifecycleVersionLabel)
	}

	hasApisLabel, err := imagehelpers.HasLabel(buildImage, lifecycleApisLabel)
	if err != nil {
		return fmt.Errorf("could not get label %s: %w", lifecycleApisLabel, err)
	}

	if !hasApisLabel {
		return fmt.Errorf("missing label %s", lifecycleApisLabel)
	}

	return nil
}

// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"io"

	"github.com/pkg/errors"

	"github.com/pivotal/build-service-cli/pkg/registry"
)

type FakeStackUploader struct {
	Images  map[string]string
	StackID string
}

func (f FakeStackUploader) UploadStackImages(buildImageTag, runImageTag, dest string, tlsCfg registry.TLSConfig, writer io.Writer) (string, string, error) {
	uploadedBuildImage, ok := f.Images[buildImageTag]
	if !ok {
		return "", "", errors.Errorf("could not upload build image %s", buildImageTag)
	}

	uploadedRunImage, ok := f.Images[runImageTag]
	if !ok {
		return "", "", errors.Errorf("could not upload run image %s", buildImageTag)
	}
	return uploadedBuildImage, uploadedRunImage, nil
}

func (f FakeStackUploader) UploadedBuildImageRef(imageTag, dest string, tlsCfg registry.TLSConfig) (string, error) {
	uploadedImage, ok := f.Images[imageTag]
	if !ok {
		return "", errors.Errorf("could not get ref %s", imageTag)
	}
	return uploadedImage, nil
}

func (f FakeStackUploader) UploadedRunImageRef(imageTag, dest string, tlsCfg registry.TLSConfig) (string, error) {
	uploadedImage, ok := f.Images[imageTag]
	if !ok {
		return "", errors.Errorf("could not get ref %s", imageTag)
	}
	return uploadedImage, nil
}

func (f *FakeStackUploader) ValidateStackIDs(buildImageTag, runImageTag string, tlsCfg registry.TLSConfig) (string, error) {
	return f.StackID, nil
}

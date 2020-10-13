// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"io"

	"github.com/pkg/errors"

	"github.com/pivotal/build-service-cli/pkg/registry"
)

type FakeBuildpackageUploader map[string]string

func (f FakeBuildpackageUploader) UploadBuildpackage(w io.Writer, _ string, buildPackage string, _ registry.TLSConfig) (string, error) {
	uploadedImage, ok := f[buildPackage]
	if !ok {
		return "", errors.Errorf("could not upload %s", buildPackage)
	}
	return uploadedImage, nil
}

func (f FakeBuildpackageUploader) RelocatedBuildpackage(repository, buildPackage string, tlsCfg registry.TLSConfig) (string, error) {
	uploadedImage, ok := f[buildPackage]
	if !ok {
		return "", errors.Errorf("could not get ref %s", buildPackage)
	}
	return uploadedImage, nil
}

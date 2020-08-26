// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"io"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
)

type Relocator struct {
}

func (*Relocator) Relocate(writer io.Writer, image v1.Image, dest string) (string, error) {
	ref, err := name.ParseReference(dest, name.WeakValidation)
	if err != nil {
		return "", errors.WithStack(err)
	}

	refName := fmt.Sprintf("%s/%s", ref.Context().RegistryStr(), ref.Context().RepositoryStr())
	ref, err = name.ParseReference(refName, name.WeakValidation)
	if err != nil {
		return "", errors.WithStack(err)
	}

	digest, err := image.Digest()
	if err != nil {
		return "", errors.WithStack(err)
	}

	size, err := imageSize(image)
	if err != nil {
		return "", errors.WithStack(err)
	}

	writer.Write([]byte(fmt.Sprintf("\tUploading '%s@%s'", ref, digest)))
	spinner := newUploadSpinner(writer, size)
	defer spinner.Stop()
	go spinner.Write()

	err = remote.Write(ref, image, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", newImageAccessError(refName, err)
	}

	return fmt.Sprintf("%s@%s", refName, digest.String()), remote.Tag(ref.Context().Tag(timestampTag()), image, remote.WithAuthFromKeychain(authn.DefaultKeychain))
}

func timestampTag() string {
	now := time.Now()
	return fmt.Sprintf("%s%02d%02d%02d", now.Format("20060102"), now.Hour(), now.Minute(), now.Second())
}

func imageSize(image v1.Image) (int64, error) {
	size, err := image.Size()
	if err != nil {
		return 0, err
	}

	layers, err := image.Layers()
	if err != nil {
		return 0, err
	}

	for _, layer := range layers {
		layerSize, err := layer.Size()
		if err != nil {
			return 0, err
		}

		size += layerSize
	}
	return size, nil
}

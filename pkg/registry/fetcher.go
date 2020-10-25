// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

type Fetcher interface {
	Fetch(src string, tlsCfg TLSConfig) (v1.Image, error)
}

type DefaultFetcher struct{}

func (d DefaultFetcher) Fetch(src string, tlsCfg TLSConfig) (v1.Image, error) {
	if d.isLocal(src) {
		return tarball.ImageFromPath(src, nil)
	} else {
		imageRef, err := name.ParseReference(src, name.WeakValidation)
		if err != nil {
			return nil, err
		}

		t, err := tlsCfg.Transport()
		if err != nil {
			return nil, err
		}

		img, err := remote.Image(imageRef, remote.WithAuthFromKeychain(authn.DefaultKeychain), remote.WithTransport(t))
		if err != nil {
			return nil, newImageAccessError(imageRef.String(), err)
		}
		return img, nil
	}
}

func (d DefaultFetcher) isLocal(src string) bool {
	_, err := os.Stat(src)
	return err == nil
}

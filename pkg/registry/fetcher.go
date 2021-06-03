// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"os"
	"runtime"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

type Fetcher interface {
	Fetch(keychain authn.Keychain, src string) (v1.Image, error)
}

type DefaultFetcher struct {
	tlsCfg TLSConfig
}

func NewDefaultFetcher(tlsCfg TLSConfig) DefaultFetcher {
	return DefaultFetcher{tlsCfg: tlsCfg}
}

func (d DefaultFetcher) Fetch(keychain authn.Keychain, src string) (v1.Image, error) {
	if d.isLocal(src) {
		return tarball.ImageFromPath(src, nil)
	} else {
		imageRef, err := name.ParseReference(src, name.WeakValidation)
		if err != nil {
			return nil, err
		}

		// Do not verify with custom CA on windows when reading from registry
		// https://github.com/golang/go/issues/16736
		if runtime.GOOS == "windows" {
			d.tlsCfg.CaCertPath = ""
		}

		t, err := d.tlsCfg.Transport()
		if err != nil {
			return nil, err
		}

		img, err := remote.Image(imageRef, remote.WithAuthFromKeychain(keychain), remote.WithTransport(t))
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

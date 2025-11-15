// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type FakeFetcher struct {
	images map[string]v1.Image
	err    string
}

func NewFakeFetcher() *FakeFetcher {
	return &FakeFetcher{
		images: make(map[string]v1.Image),
	}
}

func (f *FakeFetcher) SetImage(imageRef string, image v1.Image) {
	if f.images == nil {
		f.images = make(map[string]v1.Image)
	}
	f.images[imageRef] = image
}

func (f *FakeFetcher) SetError(err string) {
	f.err = err
}

func (f *FakeFetcher) Fetch(keychain authn.Keychain, src string) (v1.Image, error) {
	if f.err != "" {
		return nil, fmt.Errorf("%s", f.err)
	}

	img, ok := f.images[src]
	if !ok {
		return nil, fmt.Errorf("image not found: %s", src)
	}

	return img, nil
}

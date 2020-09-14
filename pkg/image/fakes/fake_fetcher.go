// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"errors"
	"github.com/pivotal/build-service-cli/pkg/registry"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type Fetcher struct {
	images map[string]v1.Image
}

func (f Fetcher) Fetch(src string, _ registry.TLSConfig) (v1.Image, error) {
	image, ok := f.images[src]
	if !ok {
		return nil, errors.New("image not found")
	}
	return image, nil
}

func (f *Fetcher) AddImage(identifier string, image v1.Image) {
	if f.images == nil {
		f.images = make(map[string]v1.Image)
	}
	f.images[identifier] = image
}

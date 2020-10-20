// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"errors"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/pivotal/build-service-cli/pkg/registry"
)

type Fetcher struct {
	images    map[string]v1.Image
	callCount int
}

func (f *Fetcher) Fetch(src string, _ registry.TLSConfig) (v1.Image, error) {
	f.callCount++
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

func (f *Fetcher) CallCount() int {
	return f.callCount
}

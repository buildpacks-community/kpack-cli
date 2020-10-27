// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pkg/errors"

	"github.com/pivotal/build-service-cli/pkg/registry"
)

const (
	stackLabel    = "io.buildpacks.stack.id"
	metadataLabel = "io.buildpacks.buildpackage.metadata"
)

type Fetcher struct {
	images    map[string]v1.Image
	callCount int
}

type ImageInfo struct {
	Ref    string
	Digest string
}

type StackInfo struct {
	StackID  string
	BuildImg ImageInfo
	RunImg   ImageInfo
}

type BuildpackImgInfo struct {
	Id string
	ImageInfo
}

func NewStackImagesFetcher(i ...StackInfo) *Fetcher {
	fetcher := Fetcher{}
	fetcher.AddStackImages(i...)
	return &fetcher
}

func NewBuildpackImagesFetcher(i ...BuildpackImgInfo) *Fetcher {
	fetcher := Fetcher{}
	fetcher.AddBuildpackImages(i...)
	return &fetcher
}

func (f *Fetcher) Fetch(src string, _ registry.TLSConfig) (v1.Image, error) {
	f.callCount++
	image, ok := f.images[src]
	if !ok {
		return nil, errors.Errorf("image not found: %q", src)
	}
	return image, nil
}

func (f *Fetcher) CallCount() int {
	return f.callCount
}

func (f *Fetcher) AddImage(identifier string, image v1.Image) {
	f.getImages()[identifier] = image
}

func (f *Fetcher) AddBuildpackImages(infos ...BuildpackImgInfo) {
	images := f.getImages()
	for _, i := range infos {
		metadata := fmt.Sprintf("{\"id\":%q}", i.Id)
		images[i.Ref] = NewFakeLabeledImage(metadataLabel, metadata, i.Digest)
	}
}

func (f *Fetcher) AddStackImages(infos ...StackInfo) {
	images := f.getImages()
	for _, i := range infos {
		images[i.BuildImg.Ref] = NewFakeLabeledImage(stackLabel, i.StackID, i.BuildImg.Digest)
		images[i.RunImg.Ref] = NewFakeLabeledImage(stackLabel, i.StackID, i.RunImg.Digest)
	}
}

func (f *Fetcher) getImages() map[string]v1.Image {
	if f.images == nil {
		f.images = make(map[string]v1.Image)
	}
	return f.images
}

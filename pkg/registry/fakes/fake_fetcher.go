// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pkg/errors"
)

const (
	stackLabel                = "io.buildpacks.stack.id"
	buildpackageMetadataLabel = "io.buildpacks.buildpackage.metadata"
	lifecycleMetadataLabel    = "io.buildpacks.lifecycle.metadata"
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
	Id      string
	Version string
	ImageInfo
}

type LifecycleInfo struct {
	Metadata string
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

func NewLifecycleImageFetcher(i ...LifecycleInfo) *Fetcher {
	fetcher := Fetcher{}
	fetcher.AddLifecycleImages(i...)
	return &fetcher
}

func (f *Fetcher) Fetch(_ authn.Keychain, src string) (v1.Image, error) {
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

		if i.Version == "" {
			i.Version = "0.0.1"
		}

		metadata := fmt.Sprintf("{\"id\":%q, \"version\":%q}", i.Id, i.Version)
		images[i.Ref] = NewFakeLabeledImage(buildpackageMetadataLabel, metadata, i.Digest)
	}
}

func (f *Fetcher) AddStackImages(infos ...StackInfo) {
	images := f.getImages()
	for _, i := range infos {
		images[i.BuildImg.Ref] = NewFakeLabeledImage(stackLabel, i.StackID, i.BuildImg.Digest)
		images[i.RunImg.Ref] = NewFakeLabeledImage(stackLabel, i.StackID, i.RunImg.Digest)
	}
}

func (f *Fetcher) AddLifecycleImages(infos ...LifecycleInfo) {
	images := f.getImages()
	for _, i := range infos {
		images[i.Ref] = NewFakeLabeledImage(lifecycleMetadataLabel, i.Metadata, i.Digest)
	}
}

func (f *Fetcher) getImages() map[string]v1.Image {
	if f.images == nil {
		f.images = make(map[string]v1.Image)
	}
	return f.images
}

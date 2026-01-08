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
	lifecycleVersionLabel     = "io.buildpacks.lifecycle.version"
	lifecycleApisLabel        = "io.buildpacks.lifecycle.apis"
)

type Fetcher struct {
	images    map[string]v1.Image
	callCount int
	err       error
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

type LifecycleInfo struct {
	Metadata string
	Version  string
	Apis     string
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
	if f.err != nil {
		return nil, f.err
	}
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
		labels := map[string]string{
			lifecycleVersionLabel: i.Version,
			lifecycleApisLabel:    i.Apis,
		}
		if i.Metadata != "" {
			labels[lifecycleMetadataLabel] = i.Metadata
		}
		images[i.Ref] = NewFakeMultiLabeledImage(labels, i.Digest)
	}
}

func (f *Fetcher) SetError(err error) {
	f.err = err
}

func (f *Fetcher) getImages() map[string]v1.Image {
	if f.images == nil {
		f.images = make(map[string]v1.Image)
	}
	return f.images
}

package fakes

import (
	"errors"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type Fetcher struct {
	images map[string]v1.Image
}

func (f Fetcher) Fetch(src string) (v1.Image, error) {
	image, ok := f.images[src]
	if !ok {
		return nil, errors.New("image not found")
	}
	return image, nil
}

func (f *Fetcher) AddImage(image v1.Image, identifier string) {
	if f.images == nil {
		f.images = make(map[string]v1.Image)
	}
	f.images[identifier] = image
}

package fakes

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pkg/errors"
)

type FakeRelocator struct {
}

func (r FakeRelocator) Relocate(image v1.Image, dest string) (string, error) {
	digest, err := image.Digest()
	if err != nil {
		return "", err
	}
	sha := digest.String()

	destRef, err := name.ParseReference(dest)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s@%s", destRef.Context().RegistryStr(), destRef.Context().RepositoryStr(), sha), nil
}

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

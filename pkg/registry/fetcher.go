package registry

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type Fetcher struct {
}

func (*Fetcher) Fetch(src string) (v1.Image, error) {
	imageRef, err := name.ParseReference(src, name.WeakValidation)
	if err != nil {
		return nil, err
	}
	img, err := remote.Image(imageRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return nil, newImageAccessError(imageRef.String(), err)
	}
	return img, nil
}

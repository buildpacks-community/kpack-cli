package image

import (
	"fmt"
	"os"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

type Relocator interface {
	Relocate(image v1.Image, dest string) (string, error)
}

type Fetcher interface {
	Fetch(src string) (v1.Image, error)
}

type Uploader struct {
	Fetcher   Fetcher
	Relocator Relocator
}

func (u *Uploader) Upload(repository, name, image string) (string, v1.Image, error) {
	img, err := u.read(image)
	if err != nil {
		return "", nil, err
	}

	ref, err := u.Relocator.Relocate(img, fmt.Sprintf("%s/%s", repository, name))
	if err != nil {
		return "", nil, err
	}

	return ref, img, nil
}

func (u *Uploader) read(name string) (v1.Image, error) {
	if u.isLocalImage(name) {
		return tarball.ImageFromPath(name, nil)
	} else {
		return u.Fetcher.Fetch(name)
	}
}

func (u *Uploader) isLocalImage(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

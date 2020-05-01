package stack

import (
	"fmt"
	"os"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

type ImageUploader interface {
	Upload(img v1.Image, repository, name string) (string, error)
	Read(name string) (v1.Image, error)
}

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

func (u *Uploader) Upload(img v1.Image, repository, name string) (string, error) {
	ref, err := u.Relocator.Relocate(img, fmt.Sprintf("%s/%s", repository, name))
	if err != nil {
		return "", err
	}

	return ref, nil
}

func (u *Uploader) Read(name string) (v1.Image, error) {
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

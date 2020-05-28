package fakes

import "github.com/pkg/errors"

type FakeBuildpackageUploader map[string]string

func (f FakeBuildpackageUploader) Upload(_ string, buildpackage string) (string, error) {
	uploadedImage, ok := f[buildpackage]
	if !ok {
		return "", errors.Errorf("could not upload %s", buildpackage)
	}
	return uploadedImage, nil
}

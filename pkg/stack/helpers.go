package stack

import (
	"errors"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

const (
	IdLabel                     = "io.buildpacks.stack.id"
	RunImageName                = "run"
	BuildImageName              = "build"
	DefaultRepositoryAnnotation = "buildservice.pivotal.io/defaultRepository"
)

func GetStackId(img v1.Image) (string, error) {
	config, err := img.ConfigFile()
	if err != nil {
		return "", err
	}

	labels := config.Config.Labels

	id, ok := labels[IdLabel]
	if !ok {
		return "", errors.New("invalid stack image")
	}

	return id, nil
}

func GetDigest(ref string) (string, error) {
	s := strings.Split(ref, "@")
	if len(s) != 2 {
		return "", errors.New("failed to get image digest")
	}
	return s[1], nil
}

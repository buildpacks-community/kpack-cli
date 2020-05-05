package stack

import (
	"errors"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

const (
	IdLabel = "io.buildpacks.stack.id"
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

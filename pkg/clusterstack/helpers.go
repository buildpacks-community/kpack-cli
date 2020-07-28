// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"errors"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

const (
	IdLabel        = "io.buildpacks.stack.id"
	RunImageName   = "run"
	BuildImageName = "build"
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

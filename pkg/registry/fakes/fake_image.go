// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type FakeImage struct {
	labels map[string]string
	digest v1.Hash
}

func NewFakeImage(digest string) FakeImage {
	return FakeImage{
		digest: v1.Hash{
			Algorithm: "sha256",
			Hex:       digest,
		},
	}
}

func NewFakeLabeledImage(label, labelValue, digest string) FakeImage {
	return FakeImage{
		labels: map[string]string{
			label: labelValue,
		},
		digest: v1.Hash{
			Algorithm: "sha256",
			Hex:       digest,
		},
	}
}

func (f FakeImage) Layers() ([]v1.Layer, error) {
	return []v1.Layer{}, nil
}

func (f FakeImage) MediaType() (types.MediaType, error) {
	return "", nil
}

func (f FakeImage) Size() (int64, error) {
	return 0, nil
}

func (f FakeImage) ConfigName() (v1.Hash, error) {
	return v1.Hash{}, nil
}

func (f FakeImage) ConfigFile() (*v1.ConfigFile, error) {
	configFile := &v1.ConfigFile{}
	if f.labels != nil {
		configFile.Config = v1.Config{
			Labels: f.labels,
		}
	}
	return configFile, nil
}

func (f FakeImage) RawConfigFile() ([]byte, error) {
	return []byte{}, nil
}

func (f FakeImage) Digest() (v1.Hash, error) {
	return f.digest, nil
}

func (f FakeImage) Manifest() (*v1.Manifest, error) {
	return &v1.Manifest{}, nil
}

func (f FakeImage) RawManifest() ([]byte, error) {
	return []byte{}, nil
}

func (f FakeImage) LayerByDigest(v1.Hash) (v1.Layer, error) {
	return nil, nil
}

func (f FakeImage) LayerByDiffID(v1.Hash) (v1.Layer, error) {
	return nil, nil
}

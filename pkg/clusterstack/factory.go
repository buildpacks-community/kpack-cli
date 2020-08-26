// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"io"
	"path"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

type ImageFetcher interface {
	Fetch(src string) (v1.Image, error)
}

type ImageRelocator interface {
	Relocate(writer io.Writer, image v1.Image, dest string) (string, error)
}

type Factory struct {
	Printer       *commands.Logger
	Fetcher       ImageFetcher
	Relocator     ImageRelocator
	Repository    string
	BuildImageRef string
	RunImageRef   string
}

func (f *Factory) MakeStack(name string) (*v1alpha1.ClusterStack, error) {
	if err := f.validate(); err != nil {
		return nil, err
	}

	buildImage, err := f.Fetcher.Fetch(f.BuildImageRef)
	if err != nil {
		return nil, err
	}

	buildStackId, err := GetStackId(buildImage)
	if err != nil {
		return nil, err
	}

	runImage, err := f.Fetcher.Fetch(f.RunImageRef)
	if err != nil {
		return nil, err
	}

	runStackId, err := GetStackId(runImage)
	if err != nil {
		return nil, err
	}

	if buildStackId != runStackId {
		return nil, errors.Errorf("build stack '%s' does not match run stack '%s'", buildStackId, runStackId)
	}

	relocatedBuildImageRef, err := f.Relocator.Relocate(f.Printer, buildImage, path.Join(f.Repository, BuildImageName))
	if err != nil {
		return nil, err
	}

	relocatedRunImageRef, err := f.Relocator.Relocate(f.Printer, runImage, path.Join(f.Repository, RunImageName))
	if err != nil {
		return nil, err
	}

	return &v1alpha1.ClusterStack{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ClusterStackKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: v1alpha1.ClusterStackSpec{
			Id: buildStackId,
			BuildImage: v1alpha1.ClusterStackSpecImage{
				Image: relocatedBuildImageRef,
			},
			RunImage: v1alpha1.ClusterStackSpecImage{
				Image: relocatedRunImageRef,
			},
		},
	}, nil
}

func (f *Factory) validate() error {
	_, err := name.ParseReference(f.Repository, name.WeakValidation)
	return err
}

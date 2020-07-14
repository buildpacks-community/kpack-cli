// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package stack

import (
	"path"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultRepositoryAnnotation = "buildservice.pivotal.io/defaultRepository"

type ImageFetcher interface {
	Fetch(src string) (v1.Image, error)
}

type ImageRelocator interface {
	Relocate(image v1.Image, dest string) (string, error)
}

type Factory struct {
	Fetcher           ImageFetcher
	Relocator         ImageRelocator
	DefaultRepository string
	BuildImageRef     string
	RunImageRef       string
}

func (f *Factory) MakeStack(name string) (*expv1alpha1.Stack, error) {
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

	relocatedBuildImageRef, err := f.Relocator.Relocate(buildImage, path.Join(f.DefaultRepository, BuildImageName))
	if err != nil {
		return nil, err
	}

	relocatedRunImageRef, err := f.Relocator.Relocate(runImage, path.Join(f.DefaultRepository, RunImageName))
	if err != nil {
		return nil, err
	}

	return &expv1alpha1.Stack{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				defaultRepositoryAnnotation: f.DefaultRepository,
			},
		},
		Spec: expv1alpha1.StackSpec{
			Id: buildStackId,
			BuildImage: expv1alpha1.StackSpecImage{
				Image: relocatedBuildImageRef,
			},
			RunImage: expv1alpha1.StackSpecImage{
				Image: relocatedRunImageRef,
			},
		},
	}, nil
}

func (f *Factory) validate() error {
	_, err := name.ParseReference(f.DefaultRepository, name.WeakValidation)
	return err
}

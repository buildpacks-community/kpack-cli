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

	"github.com/pivotal/build-service-cli/pkg/registry"
)

type ImageFetcher interface {
	Fetch(src string, tlsCfg registry.TLSConfig) (v1.Image, error)
}

type ImageRelocator interface {
	Relocate(image v1.Image, dest string, writer io.Writer, tlsCfg registry.TLSConfig) (string, error)
}

type Printer interface {
	Printlnf(format string, args ...interface{}) error
	Writer() io.Writer
}

type Factory struct {
	Printer       Printer
	Fetcher       ImageFetcher
	Relocator     ImageRelocator
	TLSConfig     registry.TLSConfig
	Repository    string
	BuildImageRef string
	RunImageRef   string
}

func (f *Factory) MakeStack(name string) (*v1alpha1.ClusterStack, error) {
	if err := f.validate(); err != nil {
		return nil, err
	}

	relocatedBuildImageRef, relocatedRunImageRef, stackId, err := f.relocateStack()
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
			Id: stackId,
			BuildImage: v1alpha1.ClusterStackSpecImage{
				Image: relocatedBuildImageRef,
			},
			RunImage: v1alpha1.ClusterStackSpecImage{
				Image: relocatedRunImageRef,
			},
		},
	}, nil
}

func (f *Factory) UpdateStack(stack *v1alpha1.ClusterStack) (bool, error) {
	if err := f.validate(); err != nil {
		return false, err
	}

	relocatedBuildImageRef, relocatedRunImageRef, stackId, err := f.relocateStack()
	if err != nil {
		return false, err
	}

	if wasUpdated, err := wasUpdated(stack, relocatedBuildImageRef, relocatedRunImageRef, stackId); err != nil {
		return false, err
	} else if !wasUpdated {
		return false, f.Printer.Printlnf("Build and Run images already exist in stack")
	}
	return true, nil
}

func (f *Factory) relocateStack() (string, string, string, error) {
	buildImage, err := f.Fetcher.Fetch(f.BuildImageRef, f.TLSConfig)
	if err != nil {
		return "", "", "", err
	}

	buildStackId, err := GetStackId(buildImage)
	if err != nil {
		return "", "", "", err
	}

	runImage, err := f.Fetcher.Fetch(f.RunImageRef, f.TLSConfig)
	if err != nil {
		return "", "", "", err
	}

	runStackId, err := GetStackId(runImage)
	if err != nil {
		return "", "", "", err
	}

	if buildStackId != runStackId {
		return "", "", "", errors.Errorf("build stack '%s' does not match run stack '%s'", buildStackId, runStackId)
	}

	if err = f.Printer.Printlnf("Uploading to '%s'...", f.Repository); err != nil {
		return "", "", "", err
	}

	relocatedBuildImageRef, err := f.Relocator.Relocate(buildImage, path.Join(f.Repository, BuildImageName), f.Printer.Writer(), f.TLSConfig)
	if err != nil {
		return "", "", "", err
	}

	relocatedRunImageRef, err := f.Relocator.Relocate(runImage, path.Join(f.Repository, RunImageName), f.Printer.Writer(), f.TLSConfig)
	if err != nil {
		return "", "", "", err
	}

	return relocatedBuildImageRef, relocatedRunImageRef, buildStackId, nil
}

func (f *Factory) validate() error {
	_, err := name.ParseReference(f.Repository, name.WeakValidation)
	return err
}

func wasUpdated(stack *v1alpha1.ClusterStack, buildImageRef, runImageRef, stackId string) (bool, error) {
	oldBuildDigest, err := GetDigest(stack.Status.BuildImage.LatestImage)
	if err != nil {
		return false, err
	}

	newBuildDigest, err := GetDigest(buildImageRef)
	if err != nil {
		return false, err
	}

	oldRunDigest, err := GetDigest(stack.Status.RunImage.LatestImage)
	if err != nil {
		return false, err
	}

	newRunDigest, err := GetDigest(runImageRef)
	if err != nil {
		return false, err
	}

	if oldBuildDigest != newBuildDigest && oldRunDigest != newRunDigest {
		stack.Spec.BuildImage.Image = buildImageRef
		stack.Spec.RunImage.Image = runImageRef
		stack.Spec.Id = stackId
		return true, nil
	}

	return false, nil
}

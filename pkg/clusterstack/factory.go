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
)

type ImageFetcher interface {
	Fetch(src string) (v1.Image, error)
}

type ImageRelocator interface {
	Relocate(writer io.Writer, image v1.Image, dest string) (string, error)
}

type Factory struct {
	Printer       IPrinter
	Fetcher       ImageFetcher
	Relocator     ImageRelocator
	Repository    string
	BuildImageRef string
	RunImageRef   string
}

type IPrinter interface {
	Printlnf(format string, a ...interface{}) error
	Writer() io.Writer
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
		f.Printer.Printlnf("Build and Run images already exist in stack\nClusterStack Unchanged")
		return false, nil
	}
	return true, nil
}

func (f *Factory) relocateStack() (string, string, string, error) {
	buildImage, err := f.Fetcher.Fetch(f.BuildImageRef)
	if err != nil {
		return "", "", "", err
	}

	buildStackId, err := GetStackId(buildImage)
	if err != nil {
		return "", "", "", err
	}

	runImage, err := f.Fetcher.Fetch(f.RunImageRef)
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

	f.Printer.Printlnf("Uploading to '%s'...", f.Repository)

	relocatedBuildImageRef, err := f.Relocator.Relocate(f.Printer.Writer(), buildImage, path.Join(f.Repository, BuildImageName))
	if err != nil {
		return "", "", "", err
	}

	relocatedRunImageRef, err := f.Relocator.Relocate(f.Printer.Writer(), runImage, path.Join(f.Repository, RunImageName))
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

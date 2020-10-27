// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"io"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/registry"
)

type Uploader interface {
	ValidateStackIDs(buildImageTag, runImageTag string, tlsCfg registry.TLSConfig) (string, error)
	UploadStackImages(buildImageTag, runImageTag, dest string, tlsCfg registry.TLSConfig, writer io.Writer) (string, string, error)
	UploadedBuildImageRef(imageTag, dest string, tlsCfg registry.TLSConfig) (string, error)
	UploadedRunImageRef(imageTag, dest string, tlsCfg registry.TLSConfig) (string, error)
}

type Printer interface {
	Printlnf(format string, args ...interface{}) error
	PrintStatus(format string, args ...interface{}) error
	Writer() io.Writer
}

type Factory struct {
	Uploader   Uploader
	Printer    Printer
	TLSConfig  registry.TLSConfig
	Repository string
}

func (f *Factory) MakeStack(name, buildImageTag, runImageTag string) (*v1alpha1.ClusterStack, error) {
	stackID, err := f.validate(buildImageTag, runImageTag)
	if err != nil {
		return nil, err
	}

	if err := f.Printer.PrintStatus("Uploading to '%s'...", f.Repository); err != nil {
		return nil, err
	}

	relocatedBuildImageRef, relocatedRunImageRef, err := f.Uploader.UploadStackImages(buildImageTag, runImageTag, f.Repository, f.TLSConfig, f.Printer.Writer())
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
			Id: stackID,
			BuildImage: v1alpha1.ClusterStackSpecImage{
				Image: relocatedBuildImageRef,
			},
			RunImage: v1alpha1.ClusterStackSpecImage{
				Image: relocatedRunImageRef,
			},
		},
	}, nil
}

func (f *Factory) UpdateStack(stack *v1alpha1.ClusterStack, buildImageTag, runImageTag string) (bool, error) {
	stackID, err := f.validate(buildImageTag, runImageTag)
	if err != nil {
		return false, err
	}

	if err := f.Printer.PrintStatus("Uploading to '%s'...", f.Repository); err != nil {
		return false, err
	}

	relocatedBuildImageRef, relocatedRunImageRef, err := f.Uploader.UploadStackImages(buildImageTag, runImageTag, f.Repository, f.TLSConfig, f.Printer.Writer())
	if err != nil {
		return false, err
	}

	if wasUpdated, err := wasUpdated(stack, relocatedBuildImageRef, relocatedRunImageRef, stackID); err != nil {
		return false, err
	} else if !wasUpdated {
		return false, f.Printer.Printlnf("Build and Run images already exist in stack")
	}
	return true, nil
}

func (f *Factory) RelocatedBuildImage(tag string) (string, error) {
	return f.Uploader.UploadedBuildImageRef(tag, f.Repository, f.TLSConfig)
}

func (f *Factory) RelocatedRunImage(tag string) (string, error) {
	return f.Uploader.UploadedRunImageRef(tag, f.Repository, f.TLSConfig)
}

func (f *Factory) validate(buildTag, runTag string) (string, error) {
	_, err := name.ParseReference(f.Repository, name.WeakValidation)
	if err != nil {
		return "", err
	}

	return f.Uploader.ValidateStackIDs(buildTag, runTag, f.TLSConfig)
}

func wasUpdated(stack *v1alpha1.ClusterStack, buildImageRef, runImageRef, stackId string) (bool, error) {
	oldBuildDigest, err := getDigest(stack.Status.BuildImage.LatestImage)
	if err != nil {
		return false, err
	}

	newBuildDigest, err := getDigest(buildImageRef)
	if err != nil {
		return false, err
	}

	oldRunDigest, err := getDigest(stack.Status.RunImage.LatestImage)
	if err != nil {
		return false, err
	}

	newRunDigest, err := getDigest(runImageRef)
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

func getDigest(ref string) (string, error) {
	s := strings.Split(ref, "@")
	if len(s) != 2 {
		return "", errors.Errorf("failed to get image digest from reference %q", ref)
	}
	return s[1], nil
}

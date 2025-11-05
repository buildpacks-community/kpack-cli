// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstack

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/buildpacks-community/kpack-cli/pkg/config"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
	"github.com/buildpacks-community/kpack-cli/pkg/stackimage"
)

type Uploader interface {
	UploadStackImages(keychain authn.Keychain, buildImageTag, runImageTag, dest string) (string, string, error)
	ValidateStackIDs(keychain authn.Keychain, buildImageTag, runImageTag string) (string, error)
}

type Printer interface {
	Printlnf(format string, args ...interface{}) error
	PrintStatus(format string, args ...interface{}) error
}

type Factory struct {
	Uploader Uploader
	Printer  Printer
}

func NewFactory(printer Printer, relocator registry.Relocator, fetcher registry.Fetcher) *Factory {
	return &Factory{
		Uploader: &stackimage.Uploader{
			Fetcher:   fetcher,
			Relocator: relocator,
		},
		Printer: printer,
	}
}

func (f *Factory) MakeStack(keychain authn.Keychain, name, buildImageTag, runImageTag string, kpConfig config.KpConfig) (*v1alpha2.ClusterStack, error) {
	stackID, err := f.validate(keychain, buildImageTag, runImageTag)
	if err != nil {
		return nil, err
	}

	defaultRepo, err := kpConfig.DefaultRepository()
	if err != nil {
		return nil, err
	}

	if err := f.Printer.PrintStatus("Uploading to '%s'...", defaultRepo); err != nil {
		return nil, err
	}

	relocatedBuildImageRef, relocatedRunImageRef, err := f.Uploader.UploadStackImages(keychain, buildImageTag, runImageTag, defaultRepo)
	if err != nil {
		return nil, err
	}

	sa := kpConfig.ServiceAccount()

	return &v1alpha2.ClusterStack{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha2.ClusterStackKind,
			APIVersion: "kpack.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: v1alpha2.ClusterStackSpec{
			Id: stackID,
			BuildImage: v1alpha2.ClusterStackSpecImage{
				Image: relocatedBuildImageRef,
			},
			RunImage: v1alpha2.ClusterStackSpecImage{
				Image: relocatedRunImageRef,
			},
			ServiceAccountRef: &sa,
		},
	}, nil
}

func (f *Factory) UpdateStack(keychain authn.Keychain, stack *v1alpha2.ClusterStack, buildImageTag, runImageTag string, kpConfig config.KpConfig) (*v1alpha2.ClusterStack, error) {
	stackID, err := f.validate(keychain, buildImageTag, runImageTag)
	if err != nil {
		return nil, err
	}

	relocatedBuildImageRef, relocatedRunImageRef, err := f.uploadStackImages(keychain, buildImageTag, runImageTag, kpConfig)
	if err != nil {
		return nil, err
	}

	return updatedStack(stack, relocatedBuildImageRef, relocatedRunImageRef, stackID), nil
}

func (f *Factory) uploadStackImages(keychain authn.Keychain, buildImageTag, runImageTag string, kpConfig config.KpConfig) (string, string, error) {
	defaultRepo, err := kpConfig.DefaultRepository()
	if err != nil {
		return "", "", err
	}

	if err := f.Printer.PrintStatus("Uploading to '%s'...", defaultRepo); err != nil {
		return "", "", err
	}

	return f.Uploader.UploadStackImages(keychain, buildImageTag, runImageTag, defaultRepo)
}

func (f *Factory) validate(keychain authn.Keychain, buildTag, runTag string) (string, error) {
	return f.Uploader.ValidateStackIDs(keychain, buildTag, runTag)
}

func updatedStack(stack *v1alpha2.ClusterStack, buildImageRef, runImageRef, stackId string) *v1alpha2.ClusterStack {
	newStack := stack.DeepCopy()

	newStack.Spec.BuildImage.Image = buildImageRef
	newStack.Spec.RunImage.Image = runImageRef
	newStack.Spec.Id = stackId

	return newStack
}

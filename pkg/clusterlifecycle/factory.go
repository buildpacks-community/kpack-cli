// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterlifecycle

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/buildpacks-community/kpack-cli/pkg/config"
	"github.com/buildpacks-community/kpack-cli/pkg/lifecycleimage"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
)

type Uploader interface {
	UploadLifecycleImage(keychain authn.Keychain, imageTag, dest string) (string, error)
	ValidateLifecycleImage(keychain authn.Keychain, imageTag string) error
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
		Uploader: &lifecycleimage.Uploader{
			Fetcher:   fetcher,
			Relocator: relocator,
		},
		Printer: printer,
	}
}

func (f *Factory) MakeLifecycle(keychain authn.Keychain, name, imageTag string, kpConfig config.KpConfig) (*v1alpha2.ClusterLifecycle, error) {
	err := f.validate(keychain, imageTag)
	if err != nil {
		return nil, fmt.Errorf("invalid lifecycle image: %w", err)
	}

	defaultRepo, err := kpConfig.DefaultRepository()
	if err != nil {
		return nil, err
	}

	if err := f.Printer.PrintStatus("Uploading to '%s'...", defaultRepo); err != nil {
		return nil, err
	}

	relocatedImageRef, err := f.Uploader.UploadLifecycleImage(keychain, imageTag, defaultRepo)
	if err != nil {
		return nil, err
	}

	sa := kpConfig.ServiceAccount()

	return &v1alpha2.ClusterLifecycle{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha2.ClusterLifecycleKind,
			APIVersion: "kpack.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: v1alpha2.ClusterLifecycleSpec{
			ImageSource: v1alpha1.ImageSource{
				Image: relocatedImageRef,
			},
			ServiceAccountRef: &sa,
		},
	}, nil
}

func (f *Factory) UpdateLifecycle(keychain authn.Keychain, lifecycle *v1alpha2.ClusterLifecycle, imageTag string, kpConfig config.KpConfig) (*v1alpha2.ClusterLifecycle, error) {
	err := f.validate(keychain, imageTag)
	if err != nil {
		return nil, fmt.Errorf("invalid lifecycle image: %w", err)
	}

	defaultRepo, err := kpConfig.DefaultRepository()
	if err != nil {
		return nil, err
	}

	if err := f.Printer.PrintStatus("Uploading to '%s'...", defaultRepo); err != nil {
		return nil, err
	}

	relocatedImageRef, err := f.Uploader.UploadLifecycleImage(keychain, imageTag, defaultRepo)
	if err != nil {
		return nil, err
	}

	newLifecycle := lifecycle.DeepCopy()
	newLifecycle.Spec.Image = relocatedImageRef
	return newLifecycle, nil
}

func (f *Factory) validate(keychain authn.Keychain, imageTag string) error {
	return f.Uploader.ValidateLifecycleImage(keychain, imageTag)
}

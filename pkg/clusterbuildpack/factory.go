// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuildpack

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/buildpacks-community/kpack-cli/pkg/buildpackage"
	"github.com/buildpacks-community/kpack-cli/pkg/config"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
)

type BuildpackageUploader interface {
	UploadBuildpackage(keychain authn.Keychain, buildPackage, repository string) (string, error)
	ValidateBuildpackImage(keychain authn.Keychain, imageTag string) error
}

type Printer interface {
	Printlnf(format string, args ...interface{}) error
	PrintStatus(format string, args ...interface{}) error
}

type Factory struct {
	Uploader BuildpackageUploader
	Printer  Printer
}

func NewFactory(printer Printer, relocator registry.Relocator, fetcher registry.Fetcher) *Factory {
	return &Factory{
		Uploader: &buildpackage.Uploader{
			Fetcher:   fetcher,
			Relocator: relocator,
		},
		Printer: printer,
	}
}

func (f *Factory) MakeBuildpack(keychain authn.Keychain, name, imageTag string, kpConfig config.KpConfig) (*v1alpha2.ClusterBuildpack, error) {
	err := f.validate(keychain, imageTag)
	if err != nil {
		return nil, fmt.Errorf("invalid buildpack image: %w", err)
	}

	defaultRepo, err := kpConfig.DefaultRepository()
	if err != nil {
		return nil, err
	}

	if err := f.Printer.PrintStatus("Uploading to '%s'...", defaultRepo); err != nil {
		return nil, err
	}

	relocatedImageRef, err := f.Uploader.UploadBuildpackage(keychain, imageTag, defaultRepo)
	if err != nil {
		return nil, err
	}

	sa := kpConfig.ServiceAccount()

	buildpack := &v1alpha2.ClusterBuildpack{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha2.ClusterBuildpackKind,
			APIVersion: "kpack.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: v1alpha2.ClusterBuildpackSpec{
			ImageSource: corev1alpha1.ImageSource{
				Image: relocatedImageRef,
			},
			ServiceAccountRef: &sa,
		},
	}

	return buildpack, k8s.SetLastAppliedCfg(buildpack)
}

func (f *Factory) UpdateBuildpack(keychain authn.Keychain, buildpack *v1alpha2.ClusterBuildpack, imageTag string, kpConfig config.KpConfig) (*v1alpha2.ClusterBuildpack, error) {
	err := f.validate(keychain, imageTag)
	if err != nil {
		return nil, fmt.Errorf("invalid buildpack image: %w", err)
	}

	defaultRepo, err := kpConfig.DefaultRepository()
	if err != nil {
		return nil, err
	}

	if err := f.Printer.PrintStatus("Uploading to '%s'...", defaultRepo); err != nil {
		return nil, err
	}

	relocatedImageRef, err := f.Uploader.UploadBuildpackage(keychain, imageTag, defaultRepo)
	if err != nil {
		return nil, err
	}

	newBuildpack := buildpack.DeepCopy()
	newBuildpack.Spec.ImageSource.Image = relocatedImageRef
	return newBuildpack, nil
}

func (f *Factory) validate(keychain authn.Keychain, imageTag string) error {
	return f.Uploader.ValidateBuildpackImage(keychain, imageTag)
}

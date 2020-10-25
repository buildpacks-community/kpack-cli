// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"io"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/registry"
)

type BuildpackageUploader interface {
	UploadBuildpackage(buildPackage, repository string, tlsCfg registry.TLSConfig, writer io.Writer) (string, error)
	UploadedBuildpackageRef(buildPackage, repository string, tlsCfg registry.TLSConfig) (string, error)
}

type Printer interface {
	Printlnf(format string, args ...interface{}) error
	Writer() io.Writer
}

type Factory struct {
	Uploader   BuildpackageUploader
	TLSConfig  registry.TLSConfig
	Repository string
	Printer    Printer
}

func (f *Factory) MakeStore(name string, buildpackages ...string) (*v1alpha1.ClusterStore, error) {
	if err := f.validate(buildpackages); err != nil {
		return nil, err
	}

	newStore := &v1alpha1.ClusterStore{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ClusterStoreKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: v1alpha1.ClusterStoreSpec{},
	}

	for _, buildpackage := range buildpackages {
		uploadedBp, err := f.Uploader.UploadBuildpackage(buildpackage, f.Repository, f.TLSConfig, f.Printer.Writer())
		if err != nil {
			return nil, err
		}

		newStore.Spec.Sources = append(newStore.Spec.Sources, v1alpha1.StoreImage{
			Image: uploadedBp,
		})
	}

	return newStore, k8s.SetLastAppliedCfg(newStore)
}

func (f *Factory) AddToStore(store *v1alpha1.ClusterStore, buildpackages ...string) (*v1alpha1.ClusterStore, bool, error) {
	storeUpdated := false
	for _, buildpackage := range buildpackages {
		uploadedBp, err := f.Uploader.UploadBuildpackage(buildpackage, f.Repository, f.TLSConfig, f.Printer.Writer())
		if err != nil {
			return nil, false, err
		}

		if storeContains(store, uploadedBp) {
			if err = f.Printer.Printlnf("\tBuildpackage already exists in the store"); err != nil {
				return store, false, err
			}
			continue
		}

		store.Spec.Sources = append(store.Spec.Sources, v1alpha1.StoreImage{
			Image: uploadedBp,
		})

		if err = f.Printer.Printlnf("\tAdded Buildpackage"); err != nil {
			return nil, false, err
		}

		storeUpdated = true
	}

	return store, storeUpdated, nil
}

func (f *Factory) RelocatedBuildpackage(buildPackage string) (string, error) {
	return f.Uploader.UploadedBuildpackageRef(buildPackage, f.Repository, f.TLSConfig)
}

func (f *Factory) validate(buildpackages []string) error {
	if len(buildpackages) < 1 {
		return errors.New("At least one buildpackage must be provided")
	}

	_, err := name.ParseReference(f.Repository, name.WeakValidation)
	return err
}

func storeContains(store *v1alpha1.ClusterStore, buildpackage string) bool {
	digest := strings.Split(buildpackage, "@")[1]

	for _, image := range store.Spec.Sources {
		parts := strings.Split(image.Image, "@")
		if len(parts) != 2 {
			continue
		}

		if parts[1] == digest {
			return true
		}
	}
	return false
}

// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/kpack-cli/pkg/buildpackage"
	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
)

type BuildpackageUploader interface {
	UploadBuildpackage(keychain authn.Keychain, buildPackage, repository string) (string, error)
	UploadedBuildpackageRef(keychain authn.Keychain, buildPackage, repository string) (string, error)
}

type Printer interface {
	Printlnf(format string, args ...interface{}) error
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

func (f *Factory) MakeStore(keychain authn.Keychain, name string, kpConfig config.KpConfig, buildpackages ...string) (*v1alpha2.ClusterStore, error) {
	if err := f.validate(buildpackages); err != nil {
		return nil, err
	}

	newStore := &v1alpha2.ClusterStore{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha2.ClusterStoreKind,
			APIVersion: "kpack.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: v1alpha2.ClusterStoreSpec{},
	}

	defaultRepo, err := kpConfig.DefaultRepository()
	if err != nil {
		return nil, err
	}

	for _, buildpackage := range buildpackages {
		uploadedBp, err := f.Uploader.UploadBuildpackage(keychain, buildpackage, defaultRepo)
		if err != nil {
			return nil, err
		}

		newStore.Spec.Sources = append(newStore.Spec.Sources, corev1alpha1.StoreImage{
			Image: uploadedBp,
		})
	}

	return newStore, k8s.SetLastAppliedCfg(newStore)
}

func (f *Factory) AddToStore(keychain authn.Keychain, store *v1alpha2.ClusterStore, kpConfig config.KpConfig, buildpackages ...string) (*v1alpha2.ClusterStore, bool, error) {
	storeUpdated := false

	defaultRepo, err := kpConfig.DefaultRepository()
	if err != nil {
		return nil, false, err
	}

	for _, buildpackage := range buildpackages {
		uploadedBp, err := f.Uploader.UploadBuildpackage(keychain, buildpackage, defaultRepo)
		if err != nil {
			return nil, false, err
		}

		if storeContains(store, uploadedBp) {
			if err = f.Printer.Printlnf("\tBuildpackage already exists in the store"); err != nil {
				return store, false, err
			}
			continue
		}

		store.Spec.Sources = append(store.Spec.Sources, corev1alpha1.StoreImage{
			Image: uploadedBp,
		})

		if err = f.Printer.Printlnf("\tAdded Buildpackage"); err != nil {
			return nil, false, err
		}

		storeUpdated = true
	}

	return store, storeUpdated, nil
}

func (f *Factory) RelocatedBuildpackage(keychain authn.Keychain, kpConfig config.KpConfig, buildPackage string) (string, error) {
	defaultRepo, err := kpConfig.DefaultRepository()
	if err != nil {
		return "", err
	}

	return f.Uploader.UploadedBuildpackageRef(keychain, buildPackage, defaultRepo)
}

func (f *Factory) validate(buildpackages []string) error {
	if len(buildpackages) < 1 {
		return errors.New("At least one buildpackage must be provided")
	}

	return nil
}

func storeContains(store *v1alpha2.ClusterStore, buildpackage string) bool {
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

// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/buildpacks-community/kpack-cli/pkg/buildpackage"
	"github.com/buildpacks-community/kpack-cli/pkg/config"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
)

type BuildpackageUploader interface {
	UploadBuildpackage(keychain authn.Keychain, buildPackage, repository string) (string, error)
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

	sa := kpConfig.ServiceAccount()
	newStore := &v1alpha2.ClusterStore{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha2.ClusterStoreKind,
			APIVersion: "kpack.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: v1alpha2.ClusterStoreSpec{
			ServiceAccountRef: &sa,
		},
	}

	defaultRepo, err := kpConfig.DefaultRepository()
	if err != nil {
		return nil, err
	}

	for _, bp := range buildpackages {
		uploadedBp, err := f.Uploader.UploadBuildpackage(keychain, bp, defaultRepo)
		if err != nil {
			return nil, err
		}

		newStore.Spec.Sources = append(newStore.Spec.Sources, corev1alpha1.ImageSource{
			Image: uploadedBp,
		})
	}

	return newStore, k8s.SetLastAppliedCfg(newStore)
}

func (f *Factory) AddToStore(keychain authn.Keychain, store *v1alpha2.ClusterStore, kpConfig config.KpConfig, buildpackages ...string) (*v1alpha2.ClusterStore, error) {
	updatedStore := store.DeepCopy()

	defaultRepo, err := kpConfig.DefaultRepository()
	if err != nil {
		return nil, err
	}

	for _, bp := range buildpackages {
		uploadedBp, err := f.Uploader.UploadBuildpackage(keychain, bp, defaultRepo)
		if err != nil {
			return nil, err
		}

		if storeContains(updatedStore, uploadedBp) {
			if err = f.Printer.Printlnf("\tBuildpackage already exists in the store"); err != nil {
				return nil, err
			}
			continue
		}

		updatedStore.Spec.Sources = append(updatedStore.Spec.Sources, corev1alpha1.ImageSource{
			Image: uploadedBp,
		})

		if err = f.Printer.Printlnf("\tAdded Buildpackage"); err != nil {
			return nil, err
		}
	}

	return updatedStore, nil
}

func (f *Factory) RemoveFromStore(store *v1alpha2.ClusterStore, buildpackages ...string) (*v1alpha2.ClusterStore, error) {
	newStore := store.DeepCopy()

	bpToStoreImage := map[string]corev1alpha1.ImageSource{}
	for _, bp := range buildpackages {
		if storeImage, ok := getStoreImage(store, bp); !ok {
			return nil, errors.Errorf("Buildpackage '%s' does not exist in the ClusterStore", bp)
		} else {
			bpToStoreImage[bp] = storeImage
		}
	}

	for _, bp := range buildpackages {
		f.Printer.Printlnf("Removing buildpackage %s", bp)

		for i, img := range newStore.Spec.Sources {
			if img.Image == bpToStoreImage[bp].Image {
				newStore.Spec.Sources = append(newStore.Spec.Sources[:i], newStore.Spec.Sources[i+1:]...)
				break
			}
		}
	}

	return newStore, nil
}

func getStoreImage(store *v1alpha2.ClusterStore, buildpackage string) (corev1alpha1.ImageSource, bool) {
	for _, bp := range store.Status.Buildpacks {
		if fmt.Sprintf("%s@%s", bp.Id, bp.Version) == buildpackage {
			return bp.StoreImage, true
		}
	}
	return corev1alpha1.ImageSource{}, false
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

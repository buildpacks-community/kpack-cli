// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"encoding/json"
	"github.com/pivotal/build-service-cli/pkg/registry"
	"io"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

const (
	KubectlLastAppliedConfig = "kubectl.kubernetes.io/last-applied-configuration"
)

type BuildpackageUploader interface {
	UploadBuildpackage(writer io.Writer, repository, buildPackage string, tlsCfg registry.TLSConfig) (string, error)
}

type Factory struct {
	Uploader   BuildpackageUploader
	TLSConfig  registry.TLSConfig
	Repository string
	Printer    *commands.Logger
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
		uploadedBp, err := f.Uploader.UploadBuildpackage(f.Printer, f.Repository, buildpackage, f.TLSConfig)
		if err != nil {
			return nil, err
		}

		newStore.Spec.Sources = append(newStore.Spec.Sources, v1alpha1.StoreImage{
			Image: uploadedBp,
		})
	}

	marshal, err := json.Marshal(newStore)
	if err != nil {
		return nil, err
	}

	newStore.Annotations[KubectlLastAppliedConfig] = string(marshal)

	return newStore, nil
}

func (f *Factory) AddToStore(store *v1alpha1.ClusterStore, repository string, buildpackages ...string) (*v1alpha1.ClusterStore, bool, error) {
	storeUpdated := false
	for _, buildpackage := range buildpackages {
		uploadedBp, err := f.Uploader.UploadBuildpackage(f.Printer, repository, buildpackage, f.TLSConfig)
		if err != nil {
			return nil, false, err
		}

		if storeContains(store, uploadedBp) {
			f.Printer.Printf("\tBuildpackage already exists in the store")
			continue
		}

		store.Spec.Sources = append(store.Spec.Sources, v1alpha1.StoreImage{
			Image: uploadedBp,
		})
		f.Printer.Printf("\tAdded Buildpackage")
		storeUpdated = true
	}

	return store, storeUpdated, nil
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

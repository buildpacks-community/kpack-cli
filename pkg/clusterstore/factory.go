// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"encoding/json"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
)

const (
	DefaultRepositoryAnnotation = "buildservice.pivotal.io/defaultRepository"
	KubectlLastAppliedConfig    = "kubectl.kubernetes.io/last-applied-configuration"
)

type BuildpackageUploader interface {
	Upload(repository, buildPackage string) (string, error)
}

type Factory struct {
	Uploader          BuildpackageUploader
	DefaultRepository string
	Printer           *commands.Logger
}

func (f *Factory) MakeStore(name string, buildpackages ...string) (*v1alpha1.ClusterStore, error) {
	if err := f.validate(buildpackages); err != nil {
		return nil, err
	}

	newStore := &v1alpha1.ClusterStore{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ClusterStoreKind,
			APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				DefaultRepositoryAnnotation: f.DefaultRepository,
			},
		},
		Spec: v1alpha1.ClusterStoreSpec{},
	}

	f.Printer.Printf("Uploading to '%s'...", f.DefaultRepository)

	for _, buildpackage := range buildpackages {
		uploadedBp, err := f.Uploader.Upload(f.DefaultRepository, buildpackage)
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
	f.Printer.Printf("Uploading to '%s'...", repository)

	var uploaded []string
	for _, buildpackage := range buildpackages {
		uploadedBp, err := f.Uploader.Upload(repository, buildpackage)
		if err != nil {
			return nil, false, err
		}
		uploaded = append(uploaded, uploadedBp)
	}

	storeUpdated := false
	for _, uploadedBp := range uploaded {
		if storeContains(store, uploadedBp) {
			f.Printer.Printf("Buildpackage '%s' already exists in the store", uploadedBp)
			continue
		}

		store.Spec.Sources = append(store.Spec.Sources, v1alpha1.StoreImage{
			Image: uploadedBp,
		})
		storeUpdated = true
		f.Printer.Printf("Added Buildpackage '%s'", uploadedBp)
	}

	return store, storeUpdated, nil
}

func (f *Factory) validate(buildpackages []string) error {
	if len(buildpackages) < 1 {
		return errors.New("At least one buildpackage must be provided")
	}

	_, err := name.ParseReference(f.DefaultRepository, name.WeakValidation)

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

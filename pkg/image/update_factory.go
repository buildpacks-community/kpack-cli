// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

func (f *Factory) UpdateImage(img *v1alpha2.Image) (*v1alpha2.Image, error) {
	if img.Spec.Build == nil {
		img.Spec.Build = &v1alpha2.ImageBuild{}
	}

	err := f.validateUpdate(img)
	if err != nil {
		return nil, err
	}

	updatedImage := img.DeepCopy()

	err = f.setSource(updatedImage)
	if err != nil {
		return nil, err
	}

	f.setAdditionalTags(updatedImage)

	err = f.setCacheSize(updatedImage)
	if err != nil {
		return nil, err
	}

	err = f.setBuild(updatedImage)
	if err != nil {
		return nil, err
	}

	f.setBuilder(updatedImage)

	f.setServiceAccount(updatedImage)

	return updatedImage, nil
}

func (f *Factory) validateUpdate(img *v1alpha2.Image) error {
	if err := f.validateSourceUpdate(img); err != nil {
		return err
	}

	if err := f.validateBuilder(); err != nil {
		return err
	}

	if err := f.validateEnvVars(img); err != nil {
		return err
	}

	return f.validateAdditionalTags(img)
}

func (f *Factory) validateSourceUpdate(img *v1alpha2.Image) error {
	sourceSet := paramSet{}
	sourceSet.add("git", f.GitRepo)
	sourceSet.add("blob", f.Blob)
	sourceSet.add("local-path", f.LocalPath)

	if len(sourceSet) > 1 {
		return errors.New("image source must be one of git, blob, or local-path")
	}

	if (sourceSet.contains("blob") || sourceSet.contains("local-path")) && f.GitRevision != "" {
		return errors.New("git-revision is incompatible with blob and local path image sources")
	}

	if len(sourceSet) == 0 && img.Spec.Source.Git == nil && f.GitRevision != "" {
		return errors.New("git-revision is incompatible with existing image source")
	}

	return nil
}

func (f *Factory) validateEnvVars(img *v1alpha2.Image) error {
	envVars, err := f.makeEnvVars()
	if err != nil {
		return err
	}

	for _, varName := range f.DeleteEnv {
		found := false

		for _, envVar := range img.Spec.Build.Env {
			if envVar.Name == varName {
				found = true
				break
			}
		}

		if !found {
			return errors.Errorf("delete-env parameter '%s' not found in existing image configuration", varName)
		}

		found = false

		for _, envVar := range envVars {
			if envVar.Name == varName {
				found = true
				break
			}
		}

		if found {
			return errors.Errorf("duplicate delete-env and env-var parameter '%s'", varName)
		}
	}
	return nil
}

func (f *Factory) validateAdditionalTags(img *v1alpha2.Image) error {
	for _, deleteTag := range f.DeleteAdditionalTags {
		found := false

		for _, existingTag := range img.Spec.AdditionalTags {
			if existingTag == deleteTag {
				found = true
				break
			}
		}

		if !found {
			return errors.Errorf("delete-additional-tag parameter '%s' not found in existing image additional tags", deleteTag)
		}

		found = false

		for _, newAdditionalTag := range f.AdditionalTags {
			if newAdditionalTag == deleteTag {
				found = true
				break
			}
		}

		if found {
			return errors.Errorf("duplicate delete-additional-tag and additional-tag parameter '%s'", deleteTag)
		}
	}
	return nil
}

func (f *Factory) setSource(image *v1alpha2.Image) error {
	if f.SubPath != nil {
		image.Spec.Source.SubPath = *f.SubPath
	}

	if f.GitRepo != "" || f.GitRevision != "" {
		if f.GitRepo != "" {
			image.Spec.Source.Blob = nil
			image.Spec.Source.Registry = nil
			image.Spec.Source.Git = &corev1alpha1.Git{
				URL:      f.GitRepo,
				Revision: defaultRevision,
			}
		}

		if f.GitRevision != "" {
			image.Spec.Source.Git.Revision = f.GitRevision
		}
	} else if f.Blob != "" {
		image.Spec.Source.Git = nil
		image.Spec.Source.Registry = nil
		image.Spec.Source.Blob = &corev1alpha1.Blob{URL: f.Blob}
	} else if f.LocalPath != "" {
		ref, err := name.ParseReference(image.Spec.Tag)
		if err != nil {
			return err
		}

		sourceRef, err := f.SourceUploader.Upload(authn.DefaultKeychain, ref.Context().Name()+"-source", f.LocalPath)
		if err != nil {
			return err
		}

		image.Spec.Source.Git = nil
		image.Spec.Source.Blob = nil
		image.Spec.Source.Registry = &corev1alpha1.Registry{Image: sourceRef}
	}

	return nil
}

func (f *Factory) setCacheSize(image *v1alpha2.Image) error {
	if f.CacheSize == "" {
		return nil
	}

	c, err := f.getCacheSize()
	if err != nil {
		return err
	}

	if image.Spec.Cache == nil {
		image.Spec.Cache = &v1alpha2.ImageCacheConfig{
			Volume: &v1alpha2.ImagePersistentVolumeCache{
				Size: c,
			},
		}
	} else if image.Spec.Cache.Volume == nil {
		image.Spec.Cache.Volume = &v1alpha2.ImagePersistentVolumeCache{
			Size: c,
		}
	}
	if c.Cmp(*image.Spec.Cache.Volume.Size) < 0 {
		return errors.Errorf("cache size cannot be decreased, current: %v, requested: %v", image.Spec.Cache.Volume.Size, c)
	} else {
		image.Spec.Cache.Volume.Size = c
	}

	return nil
}

func (f *Factory) setAdditionalTags(image *v1alpha2.Image) {
	for _, additionalTagToDelete := range f.DeleteAdditionalTags {
		for i, at := range image.Spec.AdditionalTags {
			if at == additionalTagToDelete {
				image.Spec.AdditionalTags = append(image.Spec.AdditionalTags[:i], image.Spec.AdditionalTags[i+1:]...)
				break
			}
		}
	}

	for _, additionalTag := range f.AdditionalTags {
		if tagExists(additionalTag, image.Spec.AdditionalTags) {
			continue
		}

		image.Spec.AdditionalTags = append(image.Spec.AdditionalTags, additionalTag)
	}
}

func tagExists(newTag string, existingTags []string) bool {
	for _, e := range existingTags {
		if e == newTag {
			return true
		}
	}
	return false
}

func (f *Factory) setBuild(image *v1alpha2.Image) error {
	for _, envToDelete := range f.DeleteEnv {
		for i, e := range image.Spec.Build.Env {
			if e.Name == envToDelete {
				image.Spec.Build.Env = append(image.Spec.Build.Env[:i], image.Spec.Build.Env[i+1:]...)
				break
			}
		}
	}

	envsToSave, err := f.makeEnvVars()
	if err != nil {
		return err
	}

	for _, env := range envsToSave {
		updated := false

		for i, e := range image.Spec.Build.Env {
			if e.Name == env.Name {
				image.Spec.Build.Env[i].Value = env.Value
				updated = true
				break
			}
		}

		if !updated {
			image.Spec.Build.Env = append(image.Spec.Build.Env, corev1.EnvVar{Name: env.Name, Value: env.Value})
		}
	}

	return nil
}

func (f *Factory) setBuilder(image *v1alpha2.Image) {
	if f.Builder != "" {
		image.Spec.Builder = corev1.ObjectReference{
			Kind:      v1alpha2.BuilderKind,
			Namespace: image.Namespace,
			Name:      f.Builder,
		}
	} else if f.ClusterBuilder != "" {
		image.Spec.Builder = corev1.ObjectReference{
			Kind: v1alpha2.ClusterBuilderKind,
			Name: f.ClusterBuilder,
		}
	}
}

func (f *Factory) setServiceAccount(image *v1alpha2.Image) {
	if f.ServiceAccount != "" {
		image.Spec.ServiceAccountName = f.ServiceAccount
	}
}

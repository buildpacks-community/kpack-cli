// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	v1alpha12 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/pivotal/build-service-cli/pkg/k8s"
)

type PatchFactory struct {
	SourceUploader SourceUploader
	GitRepo        string
	GitRevision    string
	Blob           string
	LocalPath      string
	SubPath        *string
	Builder        string
	ClusterBuilder string
	Env            []string
	DeleteEnv      []string
}

func (f *PatchFactory) MakePatch(img *v1alpha1.Image) ([]byte, error) {
	if img.Spec.Build == nil {
		img.Spec.Build = &v1alpha1.ImageBuild{}
	}

	err := f.validate(img)
	if err != nil {
		return nil, err
	}

	patchedImage := img.DeepCopy()

	err = f.setSource(patchedImage)
	if err != nil {
		return nil, err
	}

	err = f.setBuild(patchedImage)
	if err != nil {
		return nil, err
	}

	f.setBuilder(patchedImage)

	return k8s.CreatePatch(img, patchedImage)
}

func (f *PatchFactory) validate(img *v1alpha1.Image) error {
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

	builderSet := paramSet{}
	builderSet.add("builder", f.Builder)
	builderSet.add("cluster-builder", f.ClusterBuilder)

	if len(builderSet) > 1 {
		return errors.New("must provide one of builder or cluster-builder")
	}

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

func (f *PatchFactory) setSource(image *v1alpha1.Image) error {
	if f.SubPath != nil {
		image.Spec.Source.SubPath = *f.SubPath
	}

	if f.GitRepo != "" || f.GitRevision != "" {
		if f.GitRepo != "" {
			image.Spec.Source.Blob = nil
			image.Spec.Source.Registry = nil
			image.Spec.Source.Git = &v1alpha1.Git{
				URL:      f.GitRepo,
				Revision: "master",
			}
		}

		if f.GitRevision != "" {
			image.Spec.Source.Git.Revision = f.GitRevision
		}
	} else if f.Blob != "" {
		image.Spec.Source.Git = nil
		image.Spec.Source.Registry = nil
		image.Spec.Source.Blob = &v1alpha1.Blob{URL: f.Blob}
	} else if f.LocalPath != "" {
		ref, err := name.ParseReference(image.Spec.Tag)
		if err != nil {
			return err
		}

		sourceRef, err := f.SourceUploader.Upload(ref.Context().Name()+"-source", f.LocalPath)
		if err != nil {
			return err
		}

		image.Spec.Source.Git = nil
		image.Spec.Source.Blob = nil
		image.Spec.Source.Registry = &v1alpha1.Registry{Image: sourceRef}
	}

	return nil
}

func (f *PatchFactory) setBuild(image *v1alpha1.Image) error {
	for _, envToDelete := range f.DeleteEnv {
		for i, e := range image.Spec.Build.Env {
			if e.Name == envToDelete {
				image.Spec.Build.Env = append(image.Spec.Build.Env[:i], image.Spec.Build.Env[i+1:]...)
				break
			}
		}
	}

	envsToUpsert, err := f.makeEnvVars()
	if err != nil {
		return err
	}

	for _, env := range envsToUpsert {
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

func (f *PatchFactory) setBuilder(image *v1alpha1.Image) {
	if f.Builder != "" {
		image.Spec.Builder = corev1.ObjectReference{
			Kind:      v1alpha12.BuilderKind,
			Namespace: image.Namespace,
			Name:      f.Builder,
		}
	} else if f.ClusterBuilder != "" {
		image.Spec.Builder = corev1.ObjectReference{
			Kind: v1alpha12.ClusterBuilderKind,
			Name: f.ClusterBuilder,
		}
	}
}

func (f *PatchFactory) makeEnvVars() ([]corev1.EnvVar, error) {
	var envVars []corev1.EnvVar
	for _, e := range f.Env {
		idx := strings.Index(e, "=")
		if idx == -1 {
			return nil, errors.Errorf("env vars are improperly formatted")
		}
		envVars = append(envVars, corev1.EnvVar{
			Name:  e[:idx],
			Value: e[idx+1:],
		})
	}
	return envVars, nil
}

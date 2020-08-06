// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"sort"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SourceUploader interface {
	Upload(ref, path string) (string, error)
}

type Factory struct {
	SourceUploader SourceUploader
	GitRepo        string
	GitRevision    string
	Blob           string
	LocalPath      string
	SubPath        string
	Builder        string
	ClusterBuilder string
	Env            []string
}

func (f *Factory) MakeImage(name, namespace, tag string) (*v1alpha1.Image, error) {
	err := f.validate()
	if err != nil {
		return nil, err
	}

	source, err := f.makeSource(tag)
	if err != nil {
		return nil, err
	}

	envVars, err := f.makeEnvVars()
	if err != nil {
		return nil, err
	}

	builder := f.makeBuilder(namespace)

	return &v1alpha1.Image{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Image",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ImageSpec{
			Tag:            tag,
			Builder:        builder,
			ServiceAccount: "default",
			Source:         source,
			Build: &v1alpha1.ImageBuild{
				Env: envVars,
			},
		},
	}, nil
}

func (f *Factory) validate() error {
	sourceSet := paramSet{}
	sourceSet.add("git", f.GitRepo)
	sourceSet.add("blob", f.Blob)
	sourceSet.add("local-path", f.LocalPath)

	if len(sourceSet) != 1 {
		return errors.New("image source must be one of git, blob, or local-path")
	}

	if sourceSet.contains("git") && f.GitRevision == "" {
		return errors.New("missing parameter git-revision")
	}

	builderSet := paramSet{}
	builderSet.add("builder", f.Builder)
	builderSet.add("cluster-builder", f.ClusterBuilder)

	if len(builderSet) > 1 {
		return errors.New("must provide one of builder or cluster-builder")
	}

	return nil
}

func (f *Factory) makeEnvVars() ([]corev1.EnvVar, error) {
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

func (f *Factory) makeSource(tag string) (v1alpha1.SourceConfig, error) {
	if f.GitRepo != "" {
		return v1alpha1.SourceConfig{
			Git: &v1alpha1.Git{
				URL:      f.GitRepo,
				Revision: f.GitRevision,
			},
			SubPath: f.SubPath,
		}, nil
	} else if f.Blob != "" {
		return v1alpha1.SourceConfig{
			Blob: &v1alpha1.Blob{
				URL: f.Blob,
			},
			SubPath: f.SubPath,
		}, nil
	} else {
		ref, err := name.ParseReference(tag)
		if err != nil {
			return v1alpha1.SourceConfig{}, err
		}

		sourceRef, err := f.SourceUploader.Upload(ref.Context().Name()+"-source", f.LocalPath)
		if err != nil {
			return v1alpha1.SourceConfig{}, err
		}

		return v1alpha1.SourceConfig{
			Registry: &v1alpha1.Registry{
				Image: sourceRef,
			},
			SubPath: f.SubPath,
		}, nil
	}
}

func (f *Factory) makeBuilder(namespace string) corev1.ObjectReference {
	if f.Builder != "" {
		return corev1.ObjectReference{
			Kind:      v1alpha1.BuilderKind,
			Namespace: namespace,
			Name:      f.Builder,
		}
	} else if f.ClusterBuilder != "" {
		return corev1.ObjectReference{
			Kind: v1alpha1.ClusterBuilderKind,
			Name: f.ClusterBuilder,
		}
	} else {
		return corev1.ObjectReference{
			Kind: v1alpha1.ClusterBuilderKind,
			Name: "default",
		}
	}
}

type paramSet map[string]interface{}

func (p paramSet) add(key string, value string) {
	if value != "" {
		p[key] = nil
	}
}

func (p paramSet) contains(key string) bool {
	_, ok := p[key]
	return ok
}

func (p paramSet) getExtraParamsError(keys ...string) error {
	for _, k := range keys {
		delete(p, k)
	}
	var v []string
	for k := range p {
		v = append(v, k)
	}
	sort.Strings(v)
	return errors.Errorf("extraneous parameters: %s", strings.Join(v, ", "))
}

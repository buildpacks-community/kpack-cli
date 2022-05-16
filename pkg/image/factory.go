// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"io"
	"sort"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultRevision = "main"
)

var (
	keychain = authn.DefaultKeychain
)

type SourceUploader interface {
	Upload(keychain authn.Keychain, ref, path string) (string, error)
}

type Printer interface {
	Printlnf(format string, args ...interface{}) error
	PrintStatus(format string, args ...interface{}) error
	Writer() io.Writer
}

type Factory struct {
	SourceUploader       SourceUploader
	AdditionalTags       []string
	GitRepo              string
	GitRevision          string
	Blob                 string
	LocalPath            string
	SubPath              *string
	Builder              string
	ClusterBuilder       string
	Env                  []string
	CacheSize            string
	DeleteEnv            []string
	DeleteAdditionalTags []string
	Printer              Printer
	ServiceAccount       string
}

func (f *Factory) MakeImage(name, namespace, tag string) (*v1alpha2.Image, error) {
	err := f.validateCreate(tag)
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

	cacheSize, err := f.makeCacheSize()
	if err != nil {
		return nil, err
	}

	builder := f.makeBuilder(namespace)

	if f.ServiceAccount == "" {
		f.ServiceAccount = "default"
	}

	return &v1alpha2.Image{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Image",
			APIVersion: v1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha2.ImageSpec{
			Tag:                tag,
			AdditionalTags:     f.AdditionalTags,
			Builder:            builder,
			ServiceAccountName: f.ServiceAccount,
			Source:             source,
			Build: &v1alpha2.ImageBuild{
				Env: envVars,
			},
			Cache: &v1alpha2.ImageCacheConfig{
				Volume: &v1alpha2.ImagePersistentVolumeCache{
					Size: cacheSize,
				},
			},
		},
	}, nil
}

func (f *Factory) validateCreate(tag string) error {
	if err := f.validateTagsSameRegistry(tag); err != nil {
		return err
	}

	if err := f.validateSourceCreate(); err != nil {
		return err
	}

	if err := f.validateBuilder(); err != nil {
		return err
	}

	return nil
}

func (f *Factory) validateSourceCreate() error {
	sourceSet := paramSet{}
	sourceSet.add("git", f.GitRepo)
	sourceSet.add("blob", f.Blob)
	sourceSet.add("local-path", f.LocalPath)

	if len(sourceSet) != 1 {
		return errors.New("image source must be one of git, blob, or local-path")
	}

	if (sourceSet.contains("blob") || sourceSet.contains("local-path")) && f.GitRevision != "" {
		return errors.New("git-revision is incompatible with blob and local path image sources")
	}

	return nil
}

func (f *Factory) validateBuilder() error {
	builderSet := paramSet{}
	builderSet.add("builder", f.Builder)
	builderSet.add("cluster-builder", f.ClusterBuilder)

	if len(builderSet) > 1 {
		return errors.New("must provide one of builder or cluster-builder")
	}
	return nil
}

func (f *Factory) validateTagsSameRegistry(tag string) error {
	mainTag, err := name.NewTag(tag, name.WeakValidation)
	if err == nil {
		for _, t := range f.AdditionalTags {
			addT, err := name.NewTag(t, name.WeakValidation)
			if err == nil {
				if addT.RegistryStr() != mainTag.RegistryStr() {
					return errors.Errorf("all additional tags must have the same registry as tag. expected: %s, got: %s", mainTag.RegistryStr(), addT.RegistryStr())
				}
			}
		}
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

func (f *Factory) makeCacheSize() (*resource.Quantity, error) {
	if f.CacheSize == "" {
		return nil, nil
	}

	return f.getCacheSize()
}

func (f *Factory) getCacheSize() (*resource.Quantity, error) {
	c, err := resource.ParseQuantity(f.CacheSize)
	if err != nil {
		return nil, errors.New("invalid cache size, must be valid quantity ex. 2G")
	}

	if c.Sign() <= 0 {
		return nil, errors.New("cache size must be greater than 0")
	}

	return &c, nil
}

func (f *Factory) makeSource(tag string) (corev1alpha1.SourceConfig, error) {
	subPath := ""
	if f.SubPath != nil {
		subPath = *f.SubPath
	}
	if f.GitRepo != "" {
		s := corev1alpha1.SourceConfig{
			Git: &corev1alpha1.Git{
				URL:      f.GitRepo,
				Revision: defaultRevision,
			},
			SubPath: subPath,
		}
		if f.GitRevision != "" {
			s.Git.Revision = f.GitRevision
		}
		return s, nil
	} else if f.Blob != "" {
		return corev1alpha1.SourceConfig{
			Blob: &corev1alpha1.Blob{
				URL: f.Blob,
			},
			SubPath: subPath,
		}, nil
	} else {
		ref, err := name.ParseReference(tag)
		if err != nil {
			return corev1alpha1.SourceConfig{}, err
		}

		imgRepo := ref.Context().Name() + "-source"
		if err = f.Printer.PrintStatus("Uploading to '%s'...", imgRepo); err != nil {
			return corev1alpha1.SourceConfig{}, err
		}

		sourceRef, err := f.SourceUploader.Upload(keychain, imgRepo, f.LocalPath)
		if err != nil {
			return corev1alpha1.SourceConfig{}, err
		}

		return corev1alpha1.SourceConfig{
			Registry: &corev1alpha1.Registry{
				Image: sourceRef,
			},
			SubPath: subPath,
		}, nil
	}
}

func (f *Factory) makeBuilder(namespace string) corev1.ObjectReference {
	if f.Builder != "" {
		return corev1.ObjectReference{
			Kind:      v1alpha2.BuilderKind,
			Namespace: namespace,
			Name:      f.Builder,
		}
	} else if f.ClusterBuilder != "" {
		return corev1.ObjectReference{
			Kind: v1alpha2.ClusterBuilderKind,
			Name: f.ClusterBuilder,
		}
	} else {
		return corev1.ObjectReference{
			Kind: v1alpha2.ClusterBuilderKind,
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

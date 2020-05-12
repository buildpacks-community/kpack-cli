package image

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/mattbaird/jsonpatch"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	v1alpha12 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

type PatchFactory struct {
	SourceUploader SourceUploader
	GitRepo        string
	GitRevision    string
	Blob           string
	LocalPath      string
	SubPath        string
	Builder        string
	ClusterBuilder string
	Env            []string
	DeleteEnv      []string
}

func (f *PatchFactory) MakePatch(img *v1alpha1.Image) ([]byte, error) {
	err := f.validatePatch(img)
	if err != nil {
		return nil, err
	}

	patchedImage := img.DeepCopy()

	patchedImage.Spec.Source, err = f.patchSource(patchedImage)
	if err != nil {
		return nil, err
	}
	patchedImage.Spec.Build, err = f.patchBuild(patchedImage)
	if err != nil {
		return nil, err
	}
	patchedImage.Spec.Builder = f.patchBuilder(patchedImage)

	patchedImageBytes, err := json.Marshal(patchedImage)
	if err != nil {
		return nil, err
	}
	imageBytes, err := json.Marshal(img)
	if err != nil {
		return nil, err
	}

	jsonPatch, err := jsonpatch.CreatePatch(imageBytes, patchedImageBytes)
	if err != nil {
		return nil, err
	}
	return json.Marshal(jsonPatch)

}

func (f *PatchFactory) validatePatch(img *v1alpha1.Image) error {
	sourceSet := paramSet{}
	sourceSet.add("git", f.GitRepo)
	sourceSet.add("blob", f.Blob)
	sourceSet.add("local-path", f.LocalPath)

	if len(sourceSet) > 1 {
		return errors.New("image source must be one of git, blob, or local-path")
	}

	if len(sourceSet) == 1 && !sourceSet.contains("git") && f.GitRevision != "" {
		return errors.New("parameter git-revision is incompatible with blob and local path sources")
	}

	if len(sourceSet) == 0 && f.GitRevision != "" && img.Spec.Source.Git.URL == "" {
		return errors.New("parameter git-revision is incompatible with blob and local path sources")
	}

	if len(sourceSet) == 0 && f.GitRevision != "" {
		f.GitRepo = img.Spec.Source.Git.URL
	}

	builderSet := paramSet{}
	builderSet.add("builder", f.Builder)
	builderSet.add("cluster-builder", f.ClusterBuilder)

	if len(builderSet) > 1 {
		return errors.New("must provide one of builder or cluster-builder")
	}

	if len(f.DeleteEnv) > 0 {
		for _, varName := range f.DeleteEnv {
			found := false
			for _, envVar := range img.Spec.Build.Env {
				if envVar.Name == varName {
					found = true
					break
				}
			}
			if !found {
				return errors.New(fmt.Sprintf("env var to delete %s not set on image configuration", varName))
			}
		}
	}

	return nil
}

func (f *PatchFactory) sourceUpdated() bool {
	if f.GitRepo == "" && f.GitRevision == "" && f.Blob == "" && f.LocalPath == "" {
		return false
	}
	return true
}

func (f *PatchFactory) builderUpdated() bool {
	if f.ClusterBuilder == "" && f.Builder == "" {
		return false
	}
	return true
}

func (f *PatchFactory) buildUpdated() bool {
	if len(f.Env) == 0 && len(f.DeleteEnv) == 0 {
		return false
	}
	return true
}

func (f *PatchFactory) patchSource(image *v1alpha1.Image) (v1alpha1.SourceConfig, error) {
	if !f.sourceUpdated() {
		return image.Spec.Source, nil
	}

	if f.SubPath == "" {
		f.SubPath = image.Spec.Source.SubPath
	}

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
		ref, err := name.ParseReference(image.Spec.Tag)
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

func (f *PatchFactory) makeEnvVars() ([]corev1.EnvVar, error) {
	var envVars []corev1.EnvVar
	for _, e := range f.Env {
		item := strings.Split(e, "=")
		if len(item) != 2 {
			return nil, errors.Errorf("env vars are improperly formatted")
		}
		envVars = append(envVars, corev1.EnvVar{
			Name:  item[0],
			Value: item[1],
		})
	}
	return envVars, nil
}

func (f *PatchFactory) patchBuild(image *v1alpha1.Image) (*v1alpha1.ImageBuild, error) {
	if !f.buildUpdated() {
		return image.Spec.Build, nil
	}
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
		return nil, err
	}
found:
	for _, env := range envsToUpsert {
		for i, e := range image.Spec.Build.Env {
			if e.Name == env.Name {
				image.Spec.Build.Env[i].Value = env.Value
				continue found
			}
		}
		image.Spec.Build.Env = append(image.Spec.Build.Env, corev1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return image.Spec.Build, nil
}

func (f *PatchFactory) patchBuilder(image *v1alpha1.Image) corev1.ObjectReference {
	if !f.builderUpdated() {
		return image.Spec.Builder
	}

	if f.Builder != "" {
		return corev1.ObjectReference{
			Kind:      v1alpha12.CustomBuilderKind,
			Namespace: image.Namespace,
			Name:      f.Builder,
		}
	} else {
		return corev1.ObjectReference{
			Kind: v1alpha12.CustomClusterBuilderKind,
			Name: f.ClusterBuilder,
		}
	}

}

// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package secret

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DockerhubUrl  = "https://index.docker.io/v1/"
	GcrUrl        = "gcr.io"
	GcrUser       = "_json_key"
	GitAnnotation = "kpack.io/git"
)

type CredentialFetcher interface {
	FetchPassword(envVar, prompt string) (string, error)
}

type Factory struct {
	CredentialFetcher     CredentialFetcher
	DockerhubId           string
	Registry              string
	RegistryUser          string
	GcrServiceAccountFile string
	GitUrl                string
	GitSshKeyFile         string
	GitUser               string
}

func (f *Factory) MakeSecret(name, namespace string) (*corev1.Secret, string, error) {
	if err := f.validate(); err != nil {
		return nil, "", err
	}

	kind, err := f.getSecretKind()
	if err != nil {
		return nil, "", err
	}

	switch kind {
	case dockerHubKind:
		return f.makeDockerhubSecret(name, namespace)
	case gcrKind:
		return f.makeGcrSecret(name, namespace)
	case registryKind:
		return f.makeRegistrySecret(name, namespace)
	case gitSshKind:
		return f.makeGitSshSecret(name, namespace)
	case gitBasicAuthKind:
		return f.makeGitBasicAuthSecret(name, namespace)
	}

	return nil, "", errors.Errorf("incorrect flags provided")
}

func (f *Factory) validate() error {
	set := paramSet{}
	set.add("dockerhub", f.DockerhubId)
	set.add("registry", f.Registry)
	set.add("gcr", f.GcrServiceAccountFile)
	set.add("git", f.GitUrl)

	if len(set) != 1 {
		return errors.Errorf("secret must be one of dockerhub, gcr, registry, or git")
	}

	set.add("registry-user", f.RegistryUser)
	set.add("git-user", f.GitUser)
	set.add("git-ssh-key", f.GitSshKeyFile)

	if set.contains("dockerhub") && len(set) != 1 {
		return set.getExtraParamsError("dockerhub")
	}

	if set.contains("gcr") && len(set) != 1 {
		return set.getExtraParamsError("gcr")
	}

	if set.contains("registry") {
		if !set.contains("registry-user") {
			return errors.Errorf("missing parameter registry-user")
		} else if len(set) != 2 {
			return set.getExtraParamsError("registry", "registry-user")
		}
	}

	if set.contains("git") {
		if !set.contains("git-user") && !set.contains("git-ssh-key") {
			return errors.Errorf("missing parameter git-user or git-ssh-key")
		} else if set.contains("git-user") && set.contains("git-ssh-key") {
			return errors.Errorf("must provide one of git-user or git-ssh-key")
		} else if len(set) != 2 {
			return set.getExtraParamsError("git", "git-user", "git-ssh-key")
		}
	}

	if f.GitUser != "" && !(strings.HasPrefix(f.GitUrl, "http://") || strings.HasPrefix(f.GitUrl, "https://")) {
		return errors.Errorf("must provide a valid git url for basic auth (ex. https://github.com)")
	}

	if f.GitSshKeyFile != "" && !strings.HasPrefix(f.GitUrl, "git@") {
		return errors.Errorf("must provide a valid git url for SSH (ex. git@github.com)")
	}

	return nil
}

func (f *Factory) getSecretKind() (secretKind, error) {
	if f.DockerhubId != "" {
		return dockerHubKind, nil
	} else if f.Registry != "" && f.RegistryUser != "" {
		return registryKind, nil
	} else if f.GcrServiceAccountFile != "" {
		return gcrKind, nil
	} else if f.GitUrl != "" && f.GitSshKeyFile != "" {
		return gitSshKind, nil
	} else if f.GitUrl != "" && f.GitUser != "" {
		return gitBasicAuthKind, nil
	}
	return "", errors.Errorf("received secret with unknown type")
}

func (f *Factory) makeDockerhubSecret(name, namespace string) (*corev1.Secret, string, error) {
	password, err := f.CredentialFetcher.FetchPassword("DOCKER_PASSWORD", "dockerhub password: ")
	if err != nil {
		return nil, "", err
	}

	configJson := DockerConfigJson{Auths: DockerCredentials{
		DockerhubUrl: authn.AuthConfig{
			Username: f.DockerhubId,
			Password: password,
		},
	}}
	dockerCfgJson, err := json.Marshal(configJson)
	if err != nil {
		return nil, "", err
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: dockerCfgJson,
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}, DockerhubUrl, nil
}

func (f *Factory) makeGcrSecret(name string, namespace string) (*corev1.Secret, string, error) {
	password, err := ioutil.ReadFile(f.GcrServiceAccountFile)
	if err != nil {
		return nil, "", err
	}

	configJson := DockerConfigJson{Auths: DockerCredentials{
		GcrUrl: authn.AuthConfig{
			Username: GcrUser,
			Password: string(password),
		},
	}}
	dockerCfgJson, err := json.Marshal(configJson)
	if err != nil {
		return nil, "", err
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: dockerCfgJson,
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}, GcrUrl, nil
}

func (f *Factory) makeRegistrySecret(name string, namespace string) (*corev1.Secret, string, error) {
	password, err := f.CredentialFetcher.FetchPassword("REGISTRY_PASSWORD", "registry password: ")
	if err != nil {
		return nil, "", err
	}

	configJson := DockerConfigJson{Auths: DockerCredentials{
		f.Registry: authn.AuthConfig{
			Username: f.RegistryUser,
			Password: password,
		},
	}}
	dockerCfgJson, err := json.Marshal(configJson)
	if err != nil {
		return nil, "", err
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: dockerCfgJson,
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}, f.Registry, nil
}

func (f *Factory) makeGitSshSecret(name string, namespace string) (*corev1.Secret, string, error) {
	password, err := ioutil.ReadFile(f.GitSshKeyFile)
	if err != nil {
		return nil, "", err
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				GitAnnotation: f.GitUrl,
			},
		},
		Data: map[string][]byte{
			corev1.SSHAuthPrivateKey: []byte(password),
		},
		Type: corev1.SecretTypeSSHAuth,
	}, f.GitUrl, nil
}

func (f *Factory) makeGitBasicAuthSecret(name string, namespace string) (*corev1.Secret, string, error) {
	password, err := f.CredentialFetcher.FetchPassword("GIT_PASSWORD", "git password: ")
	if err != nil {
		return nil, "", err
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				GitAnnotation: f.GitUrl,
			},
		},
		Data: map[string][]byte{
			corev1.BasicAuthUsernameKey: []byte(f.GitUser),
			corev1.BasicAuthPasswordKey: []byte(password),
		},
		Type: corev1.SecretTypeBasicAuth,
	}, f.GitUrl, nil
}

type secretKind string

const (
	dockerHubKind    secretKind = "dockerhub"
	gcrKind                     = "gcr"
	registryKind                = "registry"
	gitSshKind                  = "git ssh"
	gitBasicAuthKind            = "git basic auth"
)

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

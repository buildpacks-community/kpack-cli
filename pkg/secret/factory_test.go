// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package secret_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/pivotal/build-service-cli/pkg/secret"
)

func TestSecretFactory(t *testing.T) {
	spec.Run(t, "TestSecretFactory", testSecretFactory)
}

func testSecretFactory(t *testing.T, when spec.G, it spec.S) {
	var factory *secret.Factory

	it.Before(func() {
		factory = &secret.Factory{CredentialFetcher: fakeCredentialFetcher{pw: "foo"}}
	})

	it("can make a registry secret", func() {
		factory.Registry = "registry.io"
		factory.RegistryUser = "some-reg-user"
		s, _, err := factory.MakeSecret("test-name", "test-namespace")
		require.NoError(t, err)
		require.Equal(t, "test-name", s.Name)
		require.Equal(t, "test-namespace", s.Namespace)
		require.Equal(t, `{"auths":{"registry.io":{"username":"some-reg-user","password":"foo"}}}`, string(s.Data[".dockerconfigjson"]))
	})

	when("no params are set", func() {
		it("returns an error message", func() {
			_, _, err := factory.MakeSecret("test-name", "test-namespace")
			require.EqualError(t, err, "secret must be one of dockerhub, gcr, registry, or git")
		})
	})

	when("too many params are set", func() {
		it("returns an error message", func() {
			factory.DockerhubId = "some-dockerhub-id"
			factory.GcrServiceAccountFile = "some-gcr-service-account"
			_, _, err := factory.MakeSecret("test-name", "test-namespace")
			require.EqualError(t, err, "secret must be one of dockerhub, gcr, registry, or git")
		})
	})

	when("sub params are mixed with dockerhub", func() {
		it("returns an error message", func() {
			factory.DockerhubId = "some-dockerhub-id"
			factory.RegistryUser = "some-reg-user"
			factory.GitUser = "some-git-user"
			_, _, err := factory.MakeSecret("test-name", "test-namespace")
			require.EqualError(t, err, "extraneous parameters: git-user, registry-user")
		})
	})

	when("sub params are mixed with gcr", func() {
		it("returns an error message", func() {
			factory.GcrServiceAccountFile = "some-gcr-service-account-file"
			factory.RegistryUser = "some-reg-user"
			factory.GitSshKeyFile = "some-git-ssh-key-file"
			_, _, err := factory.MakeSecret("test-name", "test-namespace")
			require.EqualError(t, err, "extraneous parameters: git-ssh-key, registry-user")
		})
	})

	when("registry is missing registry user", func() {
		it("returns an error message", func() {
			factory.Registry = "some-dockerhub-id"
			_, _, err := factory.MakeSecret("test-name", "test-namespace")
			require.EqualError(t, err, "missing parameter registry-user")
		})
	})

	when("registry uses full path", func() {
		it("uses only the registry domain", func() {
			factory.Registry = "registry.io/my-repo"
			factory.RegistryUser = "some-reg-user"
			s, _, err := factory.MakeSecret("test-name", "test-namespace")
			require.NoError(t, err)
			require.Equal(t, `{"auths":{"registry.io":{"username":"some-reg-user","password":"foo"}}}`, string(s.Data[".dockerconfigjson"]))
		})
	})

	when("sub params are mixed with registry", func() {
		it("returns an error message", func() {
			factory.Registry = "some-registry"
			factory.RegistryUser = "some-reg-user"
			factory.GitUser = "some-git-user"
			_, _, err := factory.MakeSecret("test-name", "test-namespace")
			require.EqualError(t, err, "extraneous parameters: git-user")
		})
	})

	when("sub params are mixed with git", func() {
		it("returns an error message", func() {
			it("returns an error message", func() {
				factory.GitUrl = "some-git"
				factory.RegistryUser = "some-reg-user"
				factory.GitUser = "some-git-user"
				_, _, err := factory.MakeSecret("test-name", "test-namespace")
				require.EqualError(t, err, "extraneous parameters: registry-user")
			})
		})
	})

	when("neither git basic auth nor git ssh are provided", func() {
		it("returns an error message", func() {
			it("returns an error message", func() {
				factory.GitUrl = "some-git"
				_, _, err := factory.MakeSecret("test-name", "test-namespace")
				require.EqualError(t, err, "missing parameter git-user or git-ssh-key")
			})
		})
	})

	when("both git basic auth and git ssh are provided", func() {
		it("returns an error message", func() {
			factory.GitUrl = "some-git"
			factory.GitUser = "some-git-user"
			factory.GitSshKeyFile = "some-ssh-key"
			_, _, err := factory.MakeSecret("test-name", "test-namespace")
			require.EqualError(t, err, "must provide one of git-user or git-ssh-key")
		})
	})

	when("using git basic auth", func() {
		it("validates that the git url begins with http:// or https://", func() {
			factory.GitUrl = "some-git"
			factory.GitUser = "some-git-user"
			_, _, err := factory.MakeSecret("test-name", "test-namespace")
			require.EqualError(t, err, "must provide a valid git url for basic auth (ex. https://github.com)")
		})
	})

	when("using git ssh keys", func() {
		it("validates that the git url begins with git@", func() {
			factory.GitUrl = "some-git"
			factory.GitSshKeyFile = "some-ssh-key"
			_, _, err := factory.MakeSecret("test-name", "test-namespace")
			require.EqualError(t, err, "must provide a valid git url for SSH (ex. git@github.com)")
		})
	})
}

type fakeCredentialFetcher struct {
	pw string
}

func (f fakeCredentialFetcher) FetchPassword(envVar, prompt string) (string, error) {
	return f.pw, nil
}

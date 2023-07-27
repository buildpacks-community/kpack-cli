// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package secret_test

import (
	"fmt"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/kpack-cli/pkg/secret"
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
		require.Equal(t, `{"auths":{"registry.io":{"username":"some-reg-user","password":"foo","auth":"c29tZS1yZWctdXNlcjpmb28="}}}`, string(s.Data[".dockerconfigjson"]))
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
			require.Equal(t, `{"auths":{"registry.io":{"username":"some-reg-user","password":"foo","auth":"c29tZS1yZWctdXNlcjpmb28="}}}`, string(s.Data[".dockerconfigjson"]))
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
			factory.GitUrl = "some-git"
			factory.RegistryUser = "some-reg-user"
			factory.GitUser = "some-git-user"
			_, _, err := factory.MakeSecret("test-name", "test-namespace")
			require.EqualError(t, err, "extraneous parameters: registry-user")
		})
	})

	when("neither git basic auth nor git ssh are provided", func() {
		it("returns an error message", func() {
			factory.GitUrl = "some-git"
			_, _, err := factory.MakeSecret("test-name", "test-namespace")
			require.EqualError(t, err, "missing parameter git-user or git-ssh-key")
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
		it("validates that the git url is correct", func() {
			validGitUrls := []string{
				"https://github.com",
				"http://github.com",
				"https://github.enterprise:1234",
				"http://github.enterprise:134",
				"https://domain.com",
				"http://domain.com",
				"github.com",
				"bitbucket.org",
			}
			for _, testUrl := range validGitUrls {
				factory.GitUrl = testUrl
				factory.GitUser = "some-git-user"
				s, _, err := factory.MakeSecret("test-name", "test-namespace")
				require.NotNilf(t, s, "factory.GitUrl = \"%s\" secret should not be nil", factory.GitUrl)
				require.NoError(t, err, fmt.Sprintf("factory.GitUrl = \"%s\" should not have errors", factory.GitUrl))
			}
		})

		it("validates that the git url is correct", func() {
			invalidGitUrls := []string{
				"some-git",
				"https://some-git.com/test",
				"https://some-git.com/",
				"https://domain.com/stash/csm/blabla/project.git",
				"http://github.enterprise:13444456",
			}
			for _, testUrl := range invalidGitUrls {
				factory.GitUrl = testUrl
				factory.GitUser = "some-git-user"
				s, _, err := factory.MakeSecret("test-name", "test-namespace")
				require.Nilf(t, s, "factory.GitUrl = \"%s\" secret should be nil", factory.GitUrl)
				require.EqualError(t, err, "must provide a valid git url without the repository path for basic auth (ex. https://github.com)")
			}
		})
	})

	when("using git ssh keys", func() {
		it("creates a secret when a valid ssh git url is passed", func() {
			validGitSshUrls := []string{
				"git@github.com",
				"user@domain.com",
				"test@github.com",
				"domain.com",
				"ssh://git@github.com:123",
				"ssh://git@github.com",
			}
			for _, testUrl := range validGitSshUrls {
				factory.GitUrl = testUrl
				factory.GitSshKeyFile = "./testdata/some-ssh-key.pem"
				s, _, err := factory.MakeSecret("test-name", "test-namespace")
				require.NotNilf(t, s, "factory.GitUrl = \"%s\" secret should not be nil", factory.GitUrl)
				require.NoError(t, err, fmt.Sprintf("factory.GitUrl = \"%s\" should not have errors", factory.GitUrl))
			}
		})

		it("prints an error when the git url is not valid", func() {
			invalidGitSshUrls := []string{
				"some-git",
				"git@github.com:owner",
				"git@github.com:owner/repo",
				"git@github.com:443/user/repo.git",
				"git@github.com:user/repo.git",
				"git@github.com:buildpacks-community/kpack-cli.git",
				"ssh://git@ssh.github.com:443/YOUR-USERNAME/YOUR-REPOSITORY.git",
				"git@github.com/abc.git",
				"git@github.com:user/repo.git",
				"ssh://git@example.com/path/to/repo.git",
				"ssh://user@bitbucket.org/group/my-repo.git",
				"git@custom-git-server.local:myrepo.git",
				"git@gitlab.com:username/project.git",
				"git@github.com:44335678",
			}
			for _, testUrl := range invalidGitSshUrls {
				factory.GitUrl = testUrl
				factory.GitUrl = "some-git"
				factory.GitSshKeyFile = "./testdata/some-ssh-key.pem"
				s, _, err := factory.MakeSecret("test-name", "test-namespace")
				require.Nilf(t, s, "factory.GitUrl = \"%s\" secret should be nil", factory.GitUrl)
				require.EqualError(t, err, "must provide a valid git url for SSH (ex. git@github.com)")
			}
		})
	})
}

type fakeCredentialFetcher struct {
	pw string
}

func (f fakeCredentialFetcher) FetchPassword(envVar, prompt string) (string, error) {
	return f.pw, nil
}

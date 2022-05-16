// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package secret_test

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	secretcmds "github.com/vmware-tanzu/kpack-cli/pkg/commands/secret"
	"github.com/vmware-tanzu/kpack-cli/pkg/secret"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestSecretCreateCommand(t *testing.T) {
	spec.Run(t, "TestSecretCreateCommand", testSecretCreateCommand)
}

func testSecretCreateCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		defaultNamespace = "some-default-namespace"
		namespace        = "some-namespace"
	)

	fetcher := &fakeCredentialFetcher{
		passwords: map[string]string{},
	}

	factory := &secret.Factory{
		CredentialFetcher: fetcher,
	}

	cmdFunc := func(k8sClient *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeK8sProvider(k8sClient, defaultNamespace)
		return secretcmds.NewCreateCommand(clientSetProvider, factory)
	}

	defaultServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Name:      "default",
			Namespace: defaultNamespace,
		},
	}

	defaultNamespacedServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Name:      "default",
			Namespace: namespace,
		},
	}

	customNamespacedServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Name:      "some-sa",
			Namespace: namespace,
		},
	}

	it("can create a secret and update a service account", func() {

		var (
			registry               = "my-registry.io"
			registryUser           = "my-registry-user"
			registryPassword       = "dummy-password"
			secretName             = "my-registry-cred"
			expectedRegistryConfig = fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\"}}}", registry, registryUser, registryPassword)
		)

		fetcher.passwords["REGISTRY_PASSWORD"] = registryPassword

		expectedRegistrySecret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      secretName,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				corev1.DockerConfigJsonKey: []byte(expectedRegistryConfig),
			},
			Type: corev1.SecretTypeDockerConfigJson,
		}

		testhelpers.CommandTest{
			Objects: []runtime.Object{
				customNamespacedServiceAccount,
			},
			Args: []string{secretName,
				"--registry", registry,
				"--registry-user", registryUser,
				"--service-account", "some-sa",
				"-n", namespace},
			ExpectedOutput: `Secret "my-registry-cred" created
`,
			ExpectCreates: []runtime.Object{
				expectedRegistrySecret,
			},
			ExpectPatches: []string{
				`{"imagePullSecrets":[{"name":"my-registry-cred"}],"metadata":{"annotations":{"kpack.io/managedSecret":"{\"my-registry-cred\":\"my-registry.io\"}"}},"secrets":[{"name":"my-registry-cred"}]}`,
			},
		}.TestK8s(t, cmdFunc)

	})

	when("namespace is provided", func() {
		when("creating a dockerhub secret", func() {
			var (
				dockerhubId          = "my-dockerhub-id"
				dockerPassword       = "dummy-password"
				secretName           = "my-docker-cred"
				expectedDockerConfig = fmt.Sprintf("{\"auths\":{\"https://index.docker.io/v1/\":{\"username\":\"%s\",\"password\":\"%s\"}}}", dockerhubId, dockerPassword)
			)

			fetcher.passwords["DOCKER_PASSWORD"] = dockerPassword

			it("creates a secret with the correct annotations for docker in the provided namespace and updates the default service account", func() {
				expectedDockerSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: namespace,
					},
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte(expectedDockerConfig),
					},
					Type: corev1.SecretTypeDockerConfigJson,
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultNamespacedServiceAccount,
					},
					Args: []string{secretName, "--dockerhub", dockerhubId, "-n", namespace},
					ExpectedOutput: `Secret "my-docker-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectPatches: []string{
						`{"imagePullSecrets":[{"name":"my-docker-cred"}],"metadata":{"annotations":{"kpack.io/managedSecret":"{\"my-docker-cred\":\"https://index.docker.io/v1/\"}"}},"secrets":[{"name":"my-docker-cred"}]}`,
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("creating a generic registry secret", func() {
			var (
				registry               = "my-registry.io"
				registryUser           = "my-registry-user"
				registryPassword       = "dummy-password"
				secretName             = "my-registry-cred"
				expectedRegistryConfig = fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\"}}}", registry, registryUser, registryPassword)
			)

			fetcher.passwords["REGISTRY_PASSWORD"] = registryPassword

			it("creates a secret with the correct annotations for the registry in the provided namespace and updates the default service account", func() {
				expectedDockerSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: namespace,
					},
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte(expectedRegistryConfig),
					},
					Type: corev1.SecretTypeDockerConfigJson,
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultNamespacedServiceAccount,
					},
					Args: []string{secretName, "--registry", registry, "--registry-user", registryUser, "-n", namespace},
					ExpectedOutput: `Secret "my-registry-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectPatches: []string{
						`{"imagePullSecrets":[{"name":"my-registry-cred"}],"metadata":{"annotations":{"kpack.io/managedSecret":"{\"my-registry-cred\":\"my-registry.io\"}"}},"secrets":[{"name":"my-registry-cred"}]}`,
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("creating a gcr registry secret", func() {
			var (
				gcrServiceAccountFile  = "./testdata/gcr-service-account.json"
				secretName             = "my-gcr-cred"
				expectedRegistryConfig = fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"{\"some-key\":\"some-value\"}"}}}`, secret.GcrUrl, secret.GcrUser)
			)

			fetcher.passwords[gcrServiceAccountFile] = `{"some-key":"some-value"}`

			it("creates a secret with the correct annotations for the registry in the provided namespace and updates the default service account", func() {
				expectedDockerSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: namespace,
					},
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte(expectedRegistryConfig),
					},
					Type: corev1.SecretTypeDockerConfigJson,
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultNamespacedServiceAccount,
					},
					Args: []string{secretName, "--gcr", gcrServiceAccountFile, "-n", namespace},
					ExpectedOutput: `Secret "my-gcr-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectPatches: []string{
						`{"imagePullSecrets":[{"name":"my-gcr-cred"}],"metadata":{"annotations":{"kpack.io/managedSecret":"{\"my-gcr-cred\":\"gcr.io\"}"}},"secrets":[{"name":"my-gcr-cred"}]}`,
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("creating a git ssh secret", func() {
			var (
				gitRepo    = "git@github.com"
				gitSshFile = "./testdata/git-ssh.pem"
				secretName = "my-git-ssh-cred"
			)

			fetcher.passwords[gitSshFile] = "some git ssh key"

			it("creates a secret with the correct annotations for git ssh in the provided namespace and updates the default service account", func() {
				expectedGitSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: namespace,
						Annotations: map[string]string{
							secret.GitAnnotation: gitRepo,
						},
					},
					Data: map[string][]byte{
						corev1.SSHAuthPrivateKey: []byte("some git ssh key"),
					},
					Type: corev1.SecretTypeSSHAuth,
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultNamespacedServiceAccount,
					},
					Args: []string{secretName, "--git-url", gitRepo, "--git-ssh-key", gitSshFile, "-n", namespace},
					ExpectedOutput: `Secret "my-git-ssh-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedGitSecret,
					},
					ExpectPatches: []string{
						`{"metadata":{"annotations":{"kpack.io/managedSecret":"{\"my-git-ssh-cred\":\"git@github.com\"}"}},"secrets":[{"name":"my-git-ssh-cred"}]}`,
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("creating a git basic auth secret", func() {
			var (
				gitRepo     = "https://github.com"
				gitUser     = "my-git-user"
				gitPassword = "my-git-password"
				secretName  = "my-git-basic-cred"
			)

			fetcher.passwords["GIT_PASSWORD"] = gitPassword

			it("creates a secret with the correct annotations for git basic auth in the provided namespace and updates the default service account", func() {
				expectedGitSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: namespace,
						Annotations: map[string]string{
							secret.GitAnnotation: gitRepo,
						},
					},
					Data: map[string][]byte{
						corev1.BasicAuthUsernameKey: []byte(gitUser),
						corev1.BasicAuthPasswordKey: []byte(gitPassword),
					},
					Type: corev1.SecretTypeBasicAuth,
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultNamespacedServiceAccount,
					},
					Args: []string{secretName, "--git-url", gitRepo, "--git-user", gitUser, "-n", namespace},
					ExpectedOutput: `Secret "my-git-basic-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedGitSecret,
					},
					ExpectPatches: []string{
						`{"metadata":{"annotations":{"kpack.io/managedSecret":"{\"my-git-basic-cred\":\"https://github.com\"}"}},"secrets":[{"name":"my-git-basic-cred"}]}`,
					},
				}.TestK8s(t, cmdFunc)
			})
		})
	})

	when("namespace is not provided", func() {
		when("creating a dockerhub secret", func() {
			var (
				dockerhubId          = "my-dockerhub-id"
				dockerPassword       = "dummy-password"
				secretName           = "my-docker-cred"
				expectedDockerConfig = fmt.Sprintf("{\"auths\":{\"https://index.docker.io/v1/\":{\"username\":\"%s\",\"password\":\"%s\"}}}", dockerhubId, dockerPassword)
			)

			fetcher.passwords["DOCKER_PASSWORD"] = dockerPassword

			it("creates a secret with the correct annotations for docker in the default namespace and updates the default service account", func() {
				expectedDockerSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: defaultNamespace,
					},
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte(expectedDockerConfig),
					},
					Type: corev1.SecretTypeDockerConfigJson,
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--dockerhub", dockerhubId},
					ExpectedOutput: `Secret "my-docker-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectPatches: []string{
						`{"imagePullSecrets":[{"name":"my-docker-cred"}],"metadata":{"annotations":{"kpack.io/managedSecret":"{\"my-docker-cred\":\"https://index.docker.io/v1/\"}"}},"secrets":[{"name":"my-docker-cred"}]}`,
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("creating a generic registry secret", func() {
			var (
				registry               = "my-registry.io"
				registryUser           = "my-registry-user"
				registryPassword       = "dummy-password"
				secretName             = "my-registry-cred"
				expectedRegistryConfig = fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\"}}}", registry, registryUser, registryPassword)
			)

			fetcher.passwords["REGISTRY_PASSWORD"] = registryPassword

			it("creates a secret with the correct annotations for the registry in the default namespace and updates the default service account", func() {
				expectedDockerSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: defaultNamespace,
					},
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte(expectedRegistryConfig),
					},
					Type: corev1.SecretTypeDockerConfigJson,
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--registry", registry, "--registry-user", registryUser},
					ExpectedOutput: `Secret "my-registry-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectPatches: []string{
						`{"imagePullSecrets":[{"name":"my-registry-cred"}],"metadata":{"annotations":{"kpack.io/managedSecret":"{\"my-registry-cred\":\"my-registry.io\"}"}},"secrets":[{"name":"my-registry-cred"}]}`,
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("creating a gcr registry secret", func() {
			var (
				gcrServiceAccountFile  = "./testdata/gcr-service-account.json"
				secretName             = "my-gcr-cred"
				expectedRegistryConfig = fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"{\"some-key\":\"some-value\"}"}}}`, secret.GcrUrl, secret.GcrUser)
			)

			fetcher.passwords[gcrServiceAccountFile] = `{"some-key":"some-value"}`

			it("creates a secret with the correct annotations for gcr in the default namespace and updates the default service account", func() {
				expectedDockerSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: defaultNamespace,
					},
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte(expectedRegistryConfig),
					},
					Type: corev1.SecretTypeDockerConfigJson,
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--gcr", gcrServiceAccountFile},
					ExpectedOutput: `Secret "my-gcr-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectPatches: []string{
						`{"imagePullSecrets":[{"name":"my-gcr-cred"}],"metadata":{"annotations":{"kpack.io/managedSecret":"{\"my-gcr-cred\":\"gcr.io\"}"}},"secrets":[{"name":"my-gcr-cred"}]}`,
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("creating a git ssh secret", func() {
			var (
				gitRepo    = "git@github.com"
				gitSshFile = "./testdata/git-ssh.pem"
				secretName = "my-git-ssh-cred"
			)

			fetcher.passwords[gitSshFile] = "some git ssh key"

			it("creates a secret with the correct annotations for git ssh in the default namespace and updates the default service account", func() {
				expectedGitSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							secret.GitAnnotation: gitRepo,
						},
					},
					Data: map[string][]byte{
						corev1.SSHAuthPrivateKey: []byte("some git ssh key"),
					},
					Type: corev1.SecretTypeSSHAuth,
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--git-url", gitRepo, "--git-ssh-key", gitSshFile},
					ExpectedOutput: `Secret "my-git-ssh-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedGitSecret,
					},
					ExpectPatches: []string{
						`{"metadata":{"annotations":{"kpack.io/managedSecret":"{\"my-git-ssh-cred\":\"git@github.com\"}"}},"secrets":[{"name":"my-git-ssh-cred"}]}`,
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("creating a git basic auth secret", func() {
			var (
				gitRepo     = "https://github.com"
				gitUser     = "my-git-user"
				gitPassword = "my-git-password"
				secretName  = "my-git-basic-cred"
			)

			fetcher.passwords["GIT_PASSWORD"] = gitPassword

			it("creates a secret with the correct annotations for git basic auth in the default namespace and updates the default service account", func() {
				expectedGitSecret := &corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      secretName,
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							secret.GitAnnotation: gitRepo,
						},
					},
					Data: map[string][]byte{
						corev1.BasicAuthUsernameKey: []byte(gitUser),
						corev1.BasicAuthPasswordKey: []byte(gitPassword),
					},
					Type: corev1.SecretTypeBasicAuth,
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--git-url", gitRepo, "--git-user", gitUser},
					ExpectedOutput: `Secret "my-git-basic-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedGitSecret,
					},
					ExpectPatches: []string{
						`{"metadata":{"annotations":{"kpack.io/managedSecret":"{\"my-git-basic-cred\":\"https://github.com\"}"}},"secrets":[{"name":"my-git-basic-cred"}]}`,
					},
				}.TestK8s(t, cmdFunc)
			})
		})
	})

	when("output flag is used", func() {
		var (
			dockerhubId          = "my-dockerhub-id"
			dockerPassword       = "dummy-password"
			secretName           = "my-docker-cred"
			expectedDockerConfig = fmt.Sprintf("{\"auths\":{\"https://index.docker.io/v1/\":{\"username\":\"%s\",\"password\":\"%s\"}}}", dockerhubId, dockerPassword)
		)

		expectedDockerSecret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      secretName,
				Namespace: defaultNamespace,
			},
			Data: map[string][]byte{
				corev1.DockerConfigJsonKey: []byte(expectedDockerConfig),
			},
			Type: corev1.SecretTypeDockerConfigJson,
		}

		fetcher.passwords["DOCKER_PASSWORD"] = dockerPassword

		it("can output in yaml format", func() {
			const resourceYAML = `apiVersion: v1
data:
  .dockerconfigjson: eyJhdXRocyI6eyJodHRwczovL2luZGV4LmRvY2tlci5pby92MS8iOnsidXNlcm5hbWUiOiJteS1kb2NrZXJodWItaWQiLCJwYXNzd29yZCI6ImR1bW15LXBhc3N3b3JkIn19fQ==
kind: Secret
metadata:
  creationTimestamp: null
  name: my-docker-cred
  namespace: some-default-namespace
type: kubernetes.io/dockerconfigjson
---
apiVersion: v1
imagePullSecrets:
- name: my-docker-cred
kind: ServiceAccount
metadata:
  annotations:
    kpack.io/managedSecret: '{"my-docker-cred":"https://index.docker.io/v1/"}'
  creationTimestamp: null
  name: default
  namespace: some-default-namespace
secrets:
- name: my-docker-cred
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					defaultServiceAccount,
				},
				Args: []string{
					secretName,
					"--dockerhub", dockerhubId,
					"--output", "yaml",
				},
				ExpectedOutput: resourceYAML,
				ExpectCreates: []runtime.Object{
					expectedDockerSecret,
				},
				ExpectPatches: []string{
					`{"imagePullSecrets":[{"name":"my-docker-cred"}],"metadata":{"annotations":{"kpack.io/managedSecret":"{\"my-docker-cred\":\"https://index.docker.io/v1/\"}"}},"secrets":[{"name":"my-docker-cred"}]}`,
				},
			}.TestK8s(t, cmdFunc)
		})

		it("can output in json format", func() {
			const resourceJSON = `{
    "kind": "Secret",
    "apiVersion": "v1",
    "metadata": {
        "name": "my-docker-cred",
        "namespace": "some-default-namespace",
        "creationTimestamp": null
    },
    "data": {
        ".dockerconfigjson": "eyJhdXRocyI6eyJodHRwczovL2luZGV4LmRvY2tlci5pby92MS8iOnsidXNlcm5hbWUiOiJteS1kb2NrZXJodWItaWQiLCJwYXNzd29yZCI6ImR1bW15LXBhc3N3b3JkIn19fQ=="
    },
    "type": "kubernetes.io/dockerconfigjson"
}
{
    "kind": "ServiceAccount",
    "apiVersion": "v1",
    "metadata": {
        "name": "default",
        "namespace": "some-default-namespace",
        "creationTimestamp": null,
        "annotations": {
            "kpack.io/managedSecret": "{\"my-docker-cred\":\"https://index.docker.io/v1/\"}"
        }
    },
    "secrets": [
        {
            "name": "my-docker-cred"
        }
    ],
    "imagePullSecrets": [
        {
            "name": "my-docker-cred"
        }
    ]
}
`

			testhelpers.CommandTest{
				Objects: []runtime.Object{
					defaultServiceAccount,
				},
				Args: []string{
					secretName,
					"--dockerhub", dockerhubId,
					"--output", "json",
				},
				ExpectedOutput: resourceJSON,
				ExpectCreates: []runtime.Object{
					expectedDockerSecret,
				},
				ExpectPatches: []string{
					`{"imagePullSecrets":[{"name":"my-docker-cred"}],"metadata":{"annotations":{"kpack.io/managedSecret":"{\"my-docker-cred\":\"https://index.docker.io/v1/\"}"}},"secrets":[{"name":"my-docker-cred"}]}`,
				},
			}.TestK8s(t, cmdFunc)
		})
	})

	when("dry-run flag is used", func() {
		fetcher.passwords["DOCKER_PASSWORD"] = "dummy-password"

		it("does not create the secret and prints result with dry run indicated", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					defaultServiceAccount,
				},
				Args: []string{
					"my-docker-cred",
					"--dockerhub", "my-dockerhub-id",
					"--dry-run",
				},
				ExpectedOutput: `Secret "my-docker-cred" created (dry run)
`,
			}.TestK8s(t, cmdFunc)
		})

		when("output flag is used", func() {
			it("does not create the secret and prints resource output", func() {
				const resourceYAML = `apiVersion: v1
data:
  .dockerconfigjson: eyJhdXRocyI6eyJodHRwczovL2luZGV4LmRvY2tlci5pby92MS8iOnsidXNlcm5hbWUiOiJteS1kb2NrZXJodWItaWQiLCJwYXNzd29yZCI6ImR1bW15LXBhc3N3b3JkIn19fQ==
kind: Secret
metadata:
  creationTimestamp: null
  name: my-docker-cred
  namespace: some-default-namespace
type: kubernetes.io/dockerconfigjson
---
apiVersion: v1
imagePullSecrets:
- name: my-docker-cred
kind: ServiceAccount
metadata:
  annotations:
    kpack.io/managedSecret: '{"my-docker-cred":"https://index.docker.io/v1/"}'
  creationTimestamp: null
  name: default
  namespace: some-default-namespace
secrets:
- name: my-docker-cred
`

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{
						"my-docker-cred",
						"--dockerhub", "my-dockerhub-id",
						"--output", "yaml",
						"--dry-run",
					},
					ExpectedOutput: resourceYAML,
				}.TestK8s(t, cmdFunc)
			})
		})
	})
}

type fakeCredentialFetcher struct {
	passwords map[string]string
}

func (f *fakeCredentialFetcher) FetchPassword(envVar, _ string) (string, error) {
	if password, ok := f.passwords[envVar]; ok {
		return password, nil
	}
	return "", errors.Errorf("secret for %s not found", envVar)
}

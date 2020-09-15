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
	clientgotesting "k8s.io/client-go/testing"

	secretcmds "github.com/pivotal/build-service-cli/pkg/commands/secret"
	"github.com/pivotal/build-service-cli/pkg/secret"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
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

	when("namespace is provided", func() {
		when("creating a dockerhub secret", func() {
			var (
				dockerhubId          = "my-dockerhub-id"
				dockerPassword       = "dummy-password"
				secretName           = "my-docker-cred"
				expectedDockerConfig = fmt.Sprintf("{\"auths\":{\"https://index.docker.io/v1/\":{\"username\":\"%s\",\"password\":\"%s\"}}}", dockerhubId, dockerPassword)
			)

			fetcher.passwords["DOCKER_PASSWORD"] = dockerPassword

			it("creates a secret with the correct annotations for docker in the provided namespace and updates the service account", func() {
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

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: namespace,
						Annotations: map[string]string{
							secretcmds.ManagedSecretAnnotationKey: fmt.Sprintf(`{"%s":"%s"}`, secretName, secret.DockerhubUrl),
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultNamespacedServiceAccount,
					},
					Args: []string{secretName, "--dockerhub", dockerhubId, "-n", namespace},
					ExpectedOutput: `"my-docker-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("creating a generic registry secret", func() {
			var (
				registry               = "my-registry.io/my-repo"
				registryUser           = "my-registry-user"
				registryPassword       = "dummy-password"
				secretName             = "my-registry-cred"
				expectedRegistryConfig = fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\"}}}", registry, registryUser, registryPassword)
			)

			fetcher.passwords["REGISTRY_PASSWORD"] = registryPassword

			it("creates a secret with the correct annotations for the registry in the provided namespace and updates the service account", func() {
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

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: namespace,
						Annotations: map[string]string{
							secretcmds.ManagedSecretAnnotationKey: fmt.Sprintf(`{"%s":"%s"}`, secretName, registry),
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultNamespacedServiceAccount,
					},
					Args: []string{secretName, "--registry", registry, "--registry-user", registryUser, "-n", namespace},
					ExpectedOutput: `"my-registry-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
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

			it("creates a secret with the correct annotations for the registry in the provided namespace and updates the service account", func() {
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

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: namespace,
						Annotations: map[string]string{
							secretcmds.ManagedSecretAnnotationKey: fmt.Sprintf(`{"%s":"%s"}`, secretName, secret.GcrUrl),
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultNamespacedServiceAccount,
					},
					Args: []string{secretName, "--gcr", gcrServiceAccountFile, "-n", namespace},
					ExpectedOutput: `"my-gcr-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
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

			it("creates a secret with the correct annotations for git ssh in the provided namespace and updates the service account", func() {
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

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: namespace,
						Annotations: map[string]string{
							secretcmds.ManagedSecretAnnotationKey: fmt.Sprintf(`{"%s":"%s"}`, secretName, gitRepo),
						},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultNamespacedServiceAccount,
					},
					Args: []string{secretName, "--git-url", gitRepo, "--git-ssh-key", gitSshFile, "-n", namespace},
					ExpectedOutput: `"my-git-ssh-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedGitSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
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

			it("creates a secret with the correct annotations for git basic auth in the provided namespace and updates the service account", func() {
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

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: namespace,
						Annotations: map[string]string{
							secretcmds.ManagedSecretAnnotationKey: fmt.Sprintf(`{"%s":"%s"}`, secretName, gitRepo),
						},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultNamespacedServiceAccount,
					},
					Args: []string{secretName, "--git-url", gitRepo, "--git-user", gitUser, "-n", namespace},
					ExpectedOutput: `"my-git-basic-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedGitSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
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

			it("creates a secret with the correct annotations for docker in the default namespace and updates the service account", func() {
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

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							secretcmds.ManagedSecretAnnotationKey: fmt.Sprintf(`{"%s":"%s"}`, secretName, secret.DockerhubUrl),
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--dockerhub", dockerhubId},
					ExpectedOutput: `"my-docker-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
					},
				}.TestK8s(t, cmdFunc)
			})
		})

		when("creating a generic registry secret", func() {
			var (
				registry               = "my-registry.io/my-repo"
				registryUser           = "my-registry-user"
				registryPassword       = "dummy-password"
				secretName             = "my-registry-cred"
				expectedRegistryConfig = fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\"}}}", registry, registryUser, registryPassword)
			)

			fetcher.passwords["REGISTRY_PASSWORD"] = registryPassword

			it("creates a secret with the correct annotations for the registry in the default namespace and updates the service account", func() {
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

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							secretcmds.ManagedSecretAnnotationKey: fmt.Sprintf(`{"%s":"%s"}`, secretName, registry),
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--registry", registry, "--registry-user", registryUser},
					ExpectedOutput: `"my-registry-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
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

			it("creates a secret with the correct annotations for gcr in the default namespace and updates the service account", func() {
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

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							secretcmds.ManagedSecretAnnotationKey: fmt.Sprintf(`{"%s":"%s"}`, secretName, secret.GcrUrl),
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secretName},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--gcr", gcrServiceAccountFile},
					ExpectedOutput: `"my-gcr-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedDockerSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
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

			it("creates a secret with the correct annotations for git ssh in the default namespace and updates the service account", func() {
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

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							secretcmds.ManagedSecretAnnotationKey: fmt.Sprintf(`{"%s":"%s"}`, secretName, gitRepo),
						},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--git-url", gitRepo, "--git-ssh-key", gitSshFile},
					ExpectedOutput: `"my-git-ssh-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedGitSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
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

			it("creates a secret with the correct annotations for git basic auth in the default namespace and updates the service account", func() {
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

				expectedServiceAccount := &corev1.ServiceAccount{
					ObjectMeta: v1.ObjectMeta{
						Name:      "default",
						Namespace: defaultNamespace,
						Annotations: map[string]string{
							secretcmds.ManagedSecretAnnotationKey: fmt.Sprintf(`{"%s":"%s"}`, secretName, gitRepo),
						},
					},
					Secrets: []corev1.ObjectReference{
						{Name: secretName},
					},
				}

				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{secretName, "--git-url", gitRepo, "--git-user", gitUser},
					ExpectedOutput: `"my-git-basic-cred" created
`,
					ExpectCreates: []runtime.Object{
						expectedGitSecret,
					},
					ExpectUpdates: []clientgotesting.UpdateActionImpl{
						{
							Object: expectedServiceAccount,
						},
					},
				}.TestK8s(t, cmdFunc)
			})
		})
	})

	when("dry run is specified", func() {
		// .dockerconfigjson value is the base64 encoding of
		// `{"auths":{"my-registry.io/my-repo":{"username":"my-registry-user","password":"dummy-password"}}}`
		const expectedYAML = `data:
  .dockerconfigjson: eyJhdXRocyI6eyJteS1yZWdpc3RyeS5pby9teS1yZXBvIjp7InVzZXJuYW1lIjoibXktcmVnaXN0cnktdXNlciIsInBhc3N3b3JkIjoiZHVtbXktcGFzc3dvcmQifX19
metadata:
  creationTimestamp: null
  name: my-registry-cred
  namespace: some-default-namespace
type: kubernetes.io/dockerconfigjson
---
data:
  .dockerconfigjson: eyJhdXRocyI6eyJteS1yZWdpc3RyeS5pby9teS1yZXBvIjp7InVzZXJuYW1lIjoibXktcmVnaXN0cnktdXNlciIsInBhc3N3b3JkIjoiZHVtbXktcGFzc3dvcmQifX19
metadata:
  creationTimestamp: null
  name: my-registry-cred
  namespace: some-default-namespace
type: kubernetes.io/dockerconfigjson
`
		const expectedJSON = `{
    "metadata": {
        "name": "my-registry-cred",
        "namespace": "some-default-namespace",
        "creationTimestamp": null
    },
    "data": {
        ".dockerconfigjson": "eyJhdXRocyI6eyJteS1yZWdpc3RyeS5pby9teS1yZXBvIjp7InVzZXJuYW1lIjoibXktcmVnaXN0cnktdXNlciIsInBhc3N3b3JkIjoiZHVtbXktcGFzc3dvcmQifX19"
    },
    "type": "kubernetes.io/dockerconfigjson"
}
{
    "metadata": {
        "name": "my-registry-cred",
        "namespace": "some-default-namespace",
        "creationTimestamp": null
    },
    "data": {
        ".dockerconfigjson": "eyJhdXRocyI6eyJteS1yZWdpc3RyeS5pby9teS1yZXBvIjp7InVzZXJuYW1lIjoibXktcmVnaXN0cnktdXNlciIsInBhc3N3b3JkIjoiZHVtbXktcGFzc3dvcmQifX19"
    },
    "type": "kubernetes.io/dockerconfigjson"
}
`
		var (
			registry               = "my-registry.io/my-repo"
			registryUser           = "my-registry-user"
			registryPassword       = "dummy-password"
			secretName             = "my-registry-cred"
		)

		fetcher.passwords["REGISTRY_PASSWORD"] = registryPassword

		when("without an output format", func() {
			it("does not create the image and defaults resource output to yaml format", func() {
				testhelpers.CommandTest{
					Objects: []runtime.Object{
						defaultServiceAccount,
					},
					Args: []string{
						secretName,
						"--registry", registry,
						"--registry-user", registryUser,
						"--dry-run",
					},
					ExpectedOutput: expectedYAML,
				}.TestK8s(t, cmdFunc)
			})
		})

		it("does not create a secret and outputs the resources in yaml format", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					defaultServiceAccount,
				},
				Args: []string{
					secretName,
					"--registry", registry,
					"--registry-user", registryUser,
					"--dry-run", "-o", "yaml",
				},
				ExpectedOutput: expectedYAML,
			}.TestK8s(t, cmdFunc)
		})

		it("does not create a secret and outputs the resource in json format", func() {
			testhelpers.CommandTest{
				Objects: []runtime.Object{
					defaultServiceAccount,
				},
				Args: []string{
					secretName,
					"--registry", registry,
					"--registry-user", registryUser,
					"--dry-run", "-o", "json",
				},
				ExpectedOutput: expectedJSON,
			}.TestK8s(t, cmdFunc)
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

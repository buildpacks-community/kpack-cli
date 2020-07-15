// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/ghodss/yaml"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/pivotal/build-service-cli/pkg/commands"
	importpkg "github.com/pivotal/build-service-cli/pkg/import"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	secretpkg "github.com/pivotal/build-service-cli/pkg/secret"
	stackpkg "github.com/pivotal/build-service-cli/pkg/stack"
	storepkg "github.com/pivotal/build-service-cli/pkg/store"
)

const (
	importNamespace          = "kpack"
	kubectlLastAppliedConfig = "kubectl.kubernetes.io/last-applied-configuration"
)

func NewImportCommand(provider k8s.ClientSetProvider, storeFactory *storepkg.Factory, stackFactory *stackpkg.Factory) *cobra.Command {
	var (
		filename   string
		repository string
		username   string
		password   string
	)

	cmd := &cobra.Command{
		Use:   "import -f <filename> -r <repository>",
		Short: "Import dependencies for stores, stacks, and custom cluster builders",
		Long: `Import dependencies for stores, stacks, and custom cluster builders from a dependency descriptor configuration.

This operation will create or update stores, stacks, and custom cluster builders defined in the dependency descriptor.
Images defined in the descriptor will be uploaded to the the registry defined with the '--repository' flag.
Therefore, you must have credentials to access the repository on your machine.

To allow resources imported into the cluster to interact with a private repository, please provide both a username and password for the target registry.
`,
		Example: `kp import -f dependencies.yaml -r my-registry.com/my-project
cat dependencies.yaml | kp import -f - -r my-registry.com/my-project
kp import -f dependencies.yaml -r my-registry.com/my-project -u my-user -p my-password`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			storeFactory.DefaultRepository = repository // FIXME
			storeFactory.Printer = commands.NewPrinter(cmd)

			stackFactory.DefaultRepository = repository // FIXME

			ref, err := name.ParseReference(repository, name.WeakValidation)
			if err != nil {
				return err
			}

			if username != "" && password == "" || username == "" && password != "" {
				return errors.Errorf("please provide both '--repository-user' and '--repository-password'")
			}

			cs, err := provider.GetClientSet("")
			if err != nil {
				return err
			}

			descriptor, err := getDependencyDescriptor(cmd, filename)
			if err != nil {
				return err
			}

			if err := importStores(descriptor, cs.KpackClient, storeFactory, repository); err != nil {
				return err
			}

			if err := importStacks(descriptor, cs.KpackClient, stackFactory); err != nil {
				return err
			}

			sa, err := genCCBSecretAndServiceAccount(cs.K8sClient, ref.Context().RegistryStr(), username, password)
			if err != nil {
				return err
			}

			if err := importCCBs(descriptor, cs.KpackClient, repository, sa); err != nil {
				return err
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "dependency descriptor filename")
	cmd.Flags().StringVarP(&repository, "repository", "r", "", "")
	cmd.Flags().StringVarP(&username, "repository-user", "u", "", "")
	cmd.Flags().StringVarP(&password, "repository-password", "p", "", "")
	_ = cmd.MarkFlagRequired("filename")
	_ = cmd.MarkFlagRequired("repository")
	return cmd
}

func getDependencyDescriptor(cmd *cobra.Command, filename string) (importpkg.DependencyDescriptor, error) {
	var (
		reader io.ReadCloser
		err    error
	)
	if filename == "-" {
		reader = ioutil.NopCloser(cmd.InOrStdin())
	} else {
		reader, err = os.Open(filename)
		if err != nil {
			return importpkg.DependencyDescriptor{}, err
		}
	}
	defer reader.Close()

	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return importpkg.DependencyDescriptor{}, err
	}

	var deps importpkg.DependencyDescriptor
	if err := yaml.Unmarshal(buf, &deps); err != nil {
		return importpkg.DependencyDescriptor{}, err
	}

	if err := deps.Validate(); err != nil {
		return importpkg.DependencyDescriptor{}, err
	}

	return deps, nil
}

func importStores(desc importpkg.DependencyDescriptor, client versioned.Interface, factory *storepkg.Factory, repository string) error {
	for _, store := range desc.Stores {
		var buildpackages []string
		for _, s := range store.Sources {
			buildpackages = append(buildpackages, s.Image)
		}

		curStore, err := client.ExperimentalV1alpha1().Stores().Get(store.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		if k8serrors.IsNotFound(err) {
			newStore, err := factory.MakeStore(store.Name, buildpackages...)
			if err != nil {
				return err
			}

			_, err = client.ExperimentalV1alpha1().Stores().Create(newStore)
			if err != nil {
				return err
			}
		} else {
			updatedStore, storeUpdated, err := factory.AddToStore(curStore, repository, buildpackages...)
			if err != nil {
				return err
			}

			if storeUpdated {
				_, err = client.ExperimentalV1alpha1().Stores().Update(updatedStore)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func importStacks(desc importpkg.DependencyDescriptor, client versioned.Interface, factory *stackpkg.Factory) error {
	for _, stack := range desc.Stacks {
		if stack.Name == desc.DefaultStack {
			desc.Stacks = append(desc.Stacks, importpkg.Stack{
				Name:       "default",
				BuildImage: stack.BuildImage,
				RunImage:   stack.RunImage,
			})
			break
		}
	}

	for _, stack := range desc.Stacks {
		factory.BuildImageRef = stack.BuildImage.Image // FIXME
		factory.RunImageRef = stack.RunImage.Image     // FIXME

		newStack, err := factory.MakeStack(stack.Name)
		if err != nil {
			return err
		}

		curStack, err := client.ExperimentalV1alpha1().Stacks().Get(stack.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		if k8serrors.IsNotFound(err) {
			_, err = client.ExperimentalV1alpha1().Stacks().Create(newStack)
			if err != nil {
				return err
			}
		} else {
			if equality.Semantic.DeepEqual(curStack.Spec, newStack.Spec) {
				continue
			}

			updateStack := curStack.DeepCopy()
			updateStack.Spec = newStack.Spec

			_, err = client.ExperimentalV1alpha1().Stacks().Update(updateStack)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func genCCBSecretAndServiceAccount(client kubernetes.Interface, registry, username, password string) (string, error) {
	if username == "" && password == "" {
		return "", nil
	}

	configJson := secretpkg.DockerConfigJson{Auths: secretpkg.DockerCredentials{
		registry: authn.AuthConfig{
			Username: username,
			Password: password,
		},
	}}

	dockerCfgJson, err := json.Marshal(configJson)
	if err != nil {
		return "", err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "ccb-secret-",
			Namespace:    importNamespace,
		},
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: dockerCfgJson,
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}

	newSecret, err := client.CoreV1().Secrets(importNamespace).Create(secret)
	if err != nil {
		return "", err
	}

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "ccb-serviceaccount-",
			Namespace:    importNamespace,
		},
		Secrets: []corev1.ObjectReference{
			{
				Name:      newSecret.Name,
				Namespace: importNamespace,
			},
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			{
				Name: newSecret.Name,
			},
		},
	}

	newSA, err := client.CoreV1().ServiceAccounts(importNamespace).Create(sa)
	if err != nil {
		return "", err
	}

	return newSA.Name, nil
}

func importCCBs(desc importpkg.DependencyDescriptor, client versioned.Interface, repository string, sa string) error {
	for _, ccb := range desc.CustomClusterBuilders {
		if ccb.Name == desc.DefaultCustomClusterBuilder {
			desc.CustomClusterBuilders = append(desc.CustomClusterBuilders, importpkg.CustomClusterBuilder{
				Name:  "default",
				Stack: ccb.Stack,
				Store: ccb.Store,
				Order: ccb.Order,
			})
			break
		}
	}

	for _, ccb := range desc.CustomClusterBuilders {
		newCCB, err := makeCCB(ccb, repository, sa)
		if err != nil {
			return err
		}

		curCCB, err := client.ExperimentalV1alpha1().CustomClusterBuilders().Get(ccb.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		if k8serrors.IsNotFound(err) {
			_, err = client.ExperimentalV1alpha1().CustomClusterBuilders().Create(newCCB)
			if err != nil {
				return err
			}
		} else {
			if equality.Semantic.DeepEqual(curCCB.Spec, newCCB.Spec) {
				continue
			}

			updateCCB := curCCB.DeepCopy()
			updateCCB.Spec = newCCB.Spec

			_, err = client.ExperimentalV1alpha1().CustomClusterBuilders().Update(updateCCB)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func makeCCB(ccb importpkg.CustomClusterBuilder, repository string, sa string) (*expv1alpha1.CustomClusterBuilder, error) {
	newCCB := &expv1alpha1.CustomClusterBuilder{
		TypeMeta: metav1.TypeMeta{
			Kind:       expv1alpha1.CustomClusterBuilderKind,
			APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        ccb.Name,
			Annotations: map[string]string{},
		},
		Spec: expv1alpha1.CustomClusterBuilderSpec{
			CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
				Tag:   path.Join(repository, ccb.Name),
				Stack: ccb.Stack,
				Store: ccb.Store,
				Order: ccb.Order,
			},
		},
	}

	if sa != "" {
		newCCB.Spec.ServiceAccountRef = corev1.ObjectReference{
			Namespace: importNamespace,
			Name:      sa,
		}
	}

	marshal, err := json.Marshal(newCCB)
	if err != nil {
		return nil, err
	}
	newCCB.Annotations[kubectlLastAppliedConfig] = string(marshal)

	return newCCB, nil
}

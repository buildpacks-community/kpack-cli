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
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/clusterstack"
	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"
	importpkg "github.com/pivotal/build-service-cli/pkg/import"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

const (
	importNamespace          = "kpack"
	kubectlLastAppliedConfig = "kubectl.kubernetes.io/last-applied-configuration"
	importTimestampKey       = "kpack.io/import-timestamp"
)

type TimestampProvider interface {
	GetTimestamp() string
}

func NewImportCommand(
	clientSetProvider k8s.ClientSetProvider,
	timestampProvider TimestampProvider,
	storeFactory *clusterstore.Factory,
	stackFactory *clusterstack.Factory) *cobra.Command {

	var (
		filename string
	)

	cmd := &cobra.Command{
		Use:   "import -f <filename>",
		Short: "Import dependencies for stores, stacks, and cluster builders",
		Long:  `This operation will create or update stores, stacks, and cluster builders defined in the dependency descriptor.`,
		Example: `kp import -f dependencies.yaml
cat dependencies.yaml | kp import -f -`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			configHelper := k8s.DefaultConfigHelper(cs)

			repository, err := configHelper.GetCanonicalRepository()
			if err != nil {
				return err
			}

			serviceAccount, err := configHelper.GetCanonicalServiceAccount()
			if err != nil {
				return err
			}

			logger := commands.NewPrinter(cmd)

			storeFactory.Repository = repository // FIXME
			storeFactory.Printer = logger

			stackFactory.Repository = repository // FIXME

			descriptor, err := getDependencyDescriptor(cmd, filename)
			if err != nil {
				return err
			}

			importHelper := importHelper{descriptor, cs.KpackClient, timestampProvider, logger}

			if err := importHelper.ImportStores(storeFactory, repository); err != nil {
				return err
			}

			if err := importHelper.ImportStacks(stackFactory); err != nil {
				return err
			}

			if err := importHelper.ImportClusterBuilders(repository, serviceAccount); err != nil {
				return err
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "dependency descriptor filename")
	_ = cmd.MarkFlagRequired("filename")
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

type importHelper struct {
	descriptor        importpkg.DependencyDescriptor
	client            kpack.Interface
	timestampProvider TimestampProvider
	logger            *commands.Logger
}

func (i importHelper) ImportStores(factory *clusterstore.Factory, repository string) error {
	for _, store := range i.descriptor.Stores {
		i.logger.Printf("Importing Cluster Store '%s'...", store.Name)

		var buildpackages []string
		for _, s := range store.Sources {
			buildpackages = append(buildpackages, s.Image)
		}

		curStore, err := i.client.KpackV1alpha1().ClusterStores().Get(store.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		if k8serrors.IsNotFound(err) {
			newStore, err := factory.MakeStore(store.Name, buildpackages...)
			newStore.Annotations[importTimestampKey] = i.timestampProvider.GetTimestamp()
			if err != nil {
				return err
			}

			_, err = i.client.KpackV1alpha1().ClusterStores().Create(newStore)
			if err != nil {
				return err
			}
		} else {
			curStore.Annotations[importTimestampKey] = i.timestampProvider.GetTimestamp()
			updatedStore, _, err := factory.AddToStore(curStore, repository, buildpackages...)
			if err != nil {
				return err
			}

			_, err = i.client.KpackV1alpha1().ClusterStores().Update(updatedStore)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (i importHelper) ImportStacks(factory *clusterstack.Factory) error {
	for _, stack := range i.descriptor.Stacks {
		if stack.Name == i.descriptor.DefaultStack {
			i.descriptor.Stacks = append(i.descriptor.Stacks, importpkg.Stack{
				Name:       "default",
				BuildImage: stack.BuildImage,
				RunImage:   stack.RunImage,
			})
			break
		}
	}

	for _, stack := range i.descriptor.Stacks {
		i.logger.Printf("Importing Cluster Stack '%s'...", stack.Name)

		factory.BuildImageRef = stack.BuildImage.Image // FIXME
		factory.RunImageRef = stack.RunImage.Image     // FIXME

		newStack, err := factory.MakeStack(stack.Name)
		if err != nil {
			return err
		}

		curStack, err := i.client.KpackV1alpha1().ClusterStacks().Get(stack.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		if k8serrors.IsNotFound(err) {
			newStack.Annotations[importTimestampKey] = i.timestampProvider.GetTimestamp()
			_, err = i.client.KpackV1alpha1().ClusterStacks().Create(newStack)
			if err != nil {
				return err
			}
		} else {
			updateStack := curStack.DeepCopy()
			updateStack.Annotations[importTimestampKey] = i.timestampProvider.GetTimestamp()
			updateStack.Spec = newStack.Spec

			_, err = i.client.KpackV1alpha1().ClusterStacks().Update(updateStack)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (i importHelper) ImportClusterBuilders(repository string, sa string) error {
	for _, ccb := range i.descriptor.ClusterBuilders {
		if ccb.Name == i.descriptor.DefaultClusterBuilder {
			i.descriptor.ClusterBuilders = append(i.descriptor.ClusterBuilders, importpkg.ClusterBuilder{
				Name:  "default",
				Stack: ccb.Stack,
				Store: ccb.Store,
				Order: ccb.Order,
			})
			break
		}
	}

	for _, ccb := range i.descriptor.ClusterBuilders {
		i.logger.Printf("Importing Cluster Builder '%s'...", ccb.Name)

		newCCB, err := i.makeCCB(ccb, repository, sa)
		if err != nil {
			return err
		}

		curCCB, err := i.client.KpackV1alpha1().ClusterBuilders().Get(ccb.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		if k8serrors.IsNotFound(err) {
			newCCB.Annotations[importTimestampKey] = i.timestampProvider.GetTimestamp()
			_, err = i.client.KpackV1alpha1().ClusterBuilders().Create(newCCB)
			if err != nil {
				return err
			}
		} else {
			updateCCB := curCCB.DeepCopy()
			updateCCB.Annotations[importTimestampKey] = i.timestampProvider.GetTimestamp()
			updateCCB.Spec = newCCB.Spec

			_, err = i.client.KpackV1alpha1().ClusterBuilders().Update(updateCCB)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (i importHelper) makeCCB(ccb importpkg.ClusterBuilder, repository string, sa string) (*v1alpha1.ClusterBuilder, error) {
	newCCB := &v1alpha1.ClusterBuilder{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ClusterBuilderKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        ccb.Name,
			Annotations: map[string]string{},
		},
		Spec: v1alpha1.ClusterBuilderSpec{
			BuilderSpec: v1alpha1.BuilderSpec{
				Tag: path.Join(repository, ccb.Name),
				Stack: corev1.ObjectReference{
					Name: ccb.Stack,
					Kind: v1alpha1.ClusterStackKind,
				},
				Store: corev1.ObjectReference{
					Name: ccb.Store,
					Kind: v1alpha1.ClusterStoreKind,
				},
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

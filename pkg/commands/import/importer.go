// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"encoding/json"
	"path"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

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

type importer struct {
	client            kpack.Interface
	timestampProvider TimestampProvider
	commandHelper     *commands.CommandHelper
	objs              []runtime.Object
}

func (i *importer) objects() []runtime.Object {
	return i.objs
}

func (i *importer) importClusterStores(clusterStores []importpkg.ClusterStore, factory *clusterstore.Factory) error {
	for _, store := range clusterStores {
		if err := i.commandHelper.PrintStatus("Importing ClusterStore '%s'...", store.Name); err != nil {
			return err
		}

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
			if err != nil {
				return err
			}

			if factory.ValidateOnly {
				continue
			}

			newStore.Annotations[importTimestampKey] = i.timestampProvider.GetTimestamp()

			if !i.commandHelper.IsDryRun() {
				if newStore, err = i.client.KpackV1alpha1().ClusterStores().Create(newStore); err != nil {
					return err
				}
			}
			i.trackObj(newStore)
		} else {
			updatedStore, _, err := factory.AddToStore(curStore, buildpackages...)
			if err != nil {
				return err
			}

			curStore.Annotations = k8s.MergeAnnotations(curStore.Annotations, map[string]string{importTimestampKey: i.timestampProvider.GetTimestamp()})

			if !i.commandHelper.IsDryRun() {
				if updatedStore, err = i.client.KpackV1alpha1().ClusterStores().Update(updatedStore); err != nil {
					return err
				}
			}
			i.trackObj(updatedStore)
		}
	}
	return nil
}

func (i *importer) importClusterStacks(clusterStacks []importpkg.ClusterStack, factory *clusterstack.Factory) error {
	for _, stack := range clusterStacks {
		if err := i.commandHelper.PrintStatus("Importing ClusterStack '%s'...", stack.Name); err != nil {
			return err
		}

		newStack, err := factory.MakeStack(stack.Name, stack.BuildImage.Image, stack.RunImage.Image)
		if err != nil {
			return err
		}

		if factory.ValidateOnly {
			continue
		}

		newStack.Annotations[importTimestampKey] = i.timestampProvider.GetTimestamp()

		curStack, err := i.client.KpackV1alpha1().ClusterStacks().Get(stack.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		if k8serrors.IsNotFound(err) {
			if !i.commandHelper.IsDryRun() {
				if newStack, err = i.client.KpackV1alpha1().ClusterStacks().Create(newStack); err != nil {
					return err
				}
			}
			i.trackObj(newStack)
		} else {
			updateStack := curStack.DeepCopy()
			updateStack.Spec = newStack.Spec
			updateStack.Annotations = k8s.MergeAnnotations(updateStack.Annotations, newStack.Annotations)

			if !i.commandHelper.IsDryRun() {
				if updateStack, err = i.client.KpackV1alpha1().ClusterStacks().Update(updateStack); err != nil {
					return err
				}
			}
			i.trackObj(updateStack)
		}
	}
	return nil
}

func (i *importer) importClusterBuilders(clusterBuilders []importpkg.ClusterBuilder, repository string, sa string) error {
	for _, ccb := range clusterBuilders {
		if err := i.commandHelper.PrintStatus("Importing ClusterBuilder '%s'...", ccb.Name); err != nil {
			return err
		}

		newCB, err := i.makeClusterBuilder(ccb, repository, sa)
		if err != nil {
			return err
		}

		newCB.Annotations[importTimestampKey] = i.timestampProvider.GetTimestamp()

		curCCB, err := i.client.KpackV1alpha1().ClusterBuilders().Get(ccb.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		if k8serrors.IsNotFound(err) {
			if !i.commandHelper.IsDryRun() {
				if newCB, err = i.client.KpackV1alpha1().ClusterBuilders().Create(newCB); err != nil {
					return err
				}
			}
			i.trackObj(newCB)
		} else {
			updateCB := curCCB.DeepCopy()
			updateCB.Spec = newCB.Spec
			updateCB.Annotations = k8s.MergeAnnotations(updateCB.Annotations, newCB.Annotations)

			if !i.commandHelper.IsDryRun() {
				if updateCB, err = i.client.KpackV1alpha1().ClusterBuilders().Update(updateCB); err != nil {
					return err
				}
			}
			i.trackObj(updateCB)
		}
	}
	return nil
}

func (i *importer) trackObj(obj runtime.Object) {
	i.objs = append(i.objs, obj)
}

func (i importer) makeClusterBuilder(ccb importpkg.ClusterBuilder, repository string, sa string) (*v1alpha1.ClusterBuilder, error) {
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
					Name: ccb.ClusterStack,
					Kind: v1alpha1.ClusterStackKind,
				},
				Store: corev1.ObjectReference{
					Name: ccb.ClusterStore,
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

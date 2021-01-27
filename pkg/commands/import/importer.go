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
	"github.com/pivotal/build-service-cli/pkg/lifecycle"
	"github.com/pivotal/build-service-cli/pkg/registry"
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
	waiter            commands.ResourceWaiter
}

func (i *importer) objects() []runtime.Object {
	return i.objs
}

type ImageUpdater interface {
	UpdateImage(srcImgStr string, tlsCfg registry.TLSConfig, hooks ...lifecycle.PreUpdateHook) (*corev1.ConfigMap, error)
}

func (i *importer) importLifecycle(srcImageTag string, cfg lifecycle.ImageUpdaterConfig) error {
	if err := i.commandHelper.PrintStatus("Importing Lifecycle..."); err != nil {
		return err
	}

	configMap, err := lifecycle.UpdateImage(srcImageTag, cfg, func(configMap *corev1.ConfigMap) {
		configMap.Annotations = k8s.MergeAnnotations(configMap.Annotations, map[string]string{importTimestampKey: i.timestampProvider.GetTimestamp()})
	})
	if err != nil {
		return err
	}

	i.trackObj(configMap)
	return nil
}

func (i *importer) importClusterStores(clusterStores []importpkg.ClusterStore, factory *clusterstore.Factory) (map[string]int64, error) {
	storeToGen := map[string]int64{}
	for _, store := range clusterStores {
		if err := i.commandHelper.PrintStatus("Importing ClusterStore '%s'...", store.Name); err != nil {
			return nil, err
		}

		var buildpackages []string
		for _, s := range store.Sources {
			buildpackages = append(buildpackages, s.Image)
		}

		curStore, err := i.client.KpackV1alpha1().ClusterStores().Get(store.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return nil, err
		}

		if k8serrors.IsNotFound(err) {
			newStore, err := factory.MakeStore(store.Name, buildpackages...)
			if err != nil {
				return nil, err
			}

			newStore.Annotations[importTimestampKey] = i.timestampProvider.GetTimestamp()

			if !i.commandHelper.IsDryRun() {
				if newStore, err = i.client.KpackV1alpha1().ClusterStores().Create(newStore); err != nil {
					return nil, err
				}
				if err := i.waiter.Wait(newStore); err != nil {
					return nil, err
				}
				storeToGen[newStore.Name] = newStore.Generation
			}
			i.trackObj(newStore)
		} else {
			updatedStore, _, err := factory.AddToStore(curStore, buildpackages...)
			if err != nil {
				return nil, err
			}

			updatedStore.Annotations = k8s.MergeAnnotations(updatedStore.Annotations, map[string]string{importTimestampKey: i.timestampProvider.GetTimestamp()})

			if !i.commandHelper.IsDryRun() {
				if updatedStore, err = i.client.KpackV1alpha1().ClusterStores().Update(updatedStore); err != nil {
					return nil, err
				}
				if err := i.waiter.Wait(updatedStore); err != nil {
					return nil, err
				}
				storeToGen[updatedStore.Name] = updatedStore.Generation
			}
			i.trackObj(updatedStore)
		}
	}
	return storeToGen, nil
}

func (i *importer) importClusterStacks(clusterStacks []importpkg.ClusterStack, factory *clusterstack.Factory) (map[string]int64, error) {
	stackToGen := map[string]int64{}
	for _, stack := range clusterStacks {
		if err := i.commandHelper.PrintStatus("Importing ClusterStack '%s'...", stack.Name); err != nil {
			return nil, err
		}

		newStack, err := factory.MakeStack(stack.Name, stack.BuildImage.Image, stack.RunImage.Image)
		if err != nil {
			return nil, err
		}

		newStack.Annotations[importTimestampKey] = i.timestampProvider.GetTimestamp()

		curStack, err := i.client.KpackV1alpha1().ClusterStacks().Get(stack.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return nil, err
		}

		if k8serrors.IsNotFound(err) {
			if !i.commandHelper.IsDryRun() {
				if newStack, err = i.client.KpackV1alpha1().ClusterStacks().Create(newStack); err != nil {
					return nil, err
				}
				if err := i.waiter.Wait(newStack); err != nil {
					return nil, err
				}
				stackToGen[newStack.Name] = newStack.Generation
			}
			i.trackObj(newStack)
		} else {
			updateStack := curStack.DeepCopy()
			updateStack.Spec = newStack.Spec
			updateStack.Annotations = k8s.MergeAnnotations(updateStack.Annotations, newStack.Annotations)

			if !i.commandHelper.IsDryRun() {
				if updateStack, err = i.client.KpackV1alpha1().ClusterStacks().Update(updateStack); err != nil {
					return nil, err
				}
				if err := i.waiter.Wait(updateStack); err != nil {
					return nil, err
				}
				stackToGen[updateStack.Name] = updateStack.Generation
			}
			i.trackObj(updateStack)
		}
	}
	return stackToGen, nil
}

func (i *importer) importClusterBuilders(clusterBuilders []importpkg.ClusterBuilder, repository, sa string, storeToGen, stackToGen map[string]int64) error {
	for _, cb := range clusterBuilders {
		if err := i.commandHelper.PrintStatus("Importing ClusterBuilder '%s'...", cb.Name); err != nil {
			return err
		}

		newCB, err := makeClusterBuilder(cb, repository, sa)
		if err != nil {
			return err
		}

		newCB.Annotations[importTimestampKey] = i.timestampProvider.GetTimestamp()

		curCB, err := i.client.KpackV1alpha1().ClusterBuilders().Get(cb.Name, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		waitCondition := builderHasResolved(storeToGen[newCB.Spec.Store.Name], stackToGen[newCB.Spec.Stack.Name])
		if k8serrors.IsNotFound(err) {
			if !i.commandHelper.IsDryRun() {
				if newCB, err = i.client.KpackV1alpha1().ClusterBuilders().Create(newCB); err != nil {
					return err
				}
				if err := i.waiter.Wait(newCB, waitCondition); err != nil {
					return err
				}
			}
			i.trackObj(newCB)
		} else {
			updateCB := curCB.DeepCopy()
			updateCB.Spec = newCB.Spec
			updateCB.Annotations = k8s.MergeAnnotations(updateCB.Annotations, newCB.Annotations)

			if !i.commandHelper.IsDryRun() {
				if updateCB, err = i.client.KpackV1alpha1().ClusterBuilders().Update(updateCB); err != nil {
					return err
				}
				if err := i.waiter.Wait(updateCB, waitCondition); err != nil {
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

func makeClusterBuilder(ccb importpkg.ClusterBuilder, repository string, sa string) (*v1alpha1.ClusterBuilder, error) {
	newCB := &v1alpha1.ClusterBuilder{
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
		newCB.Spec.ServiceAccountRef = corev1.ObjectReference{
			Namespace: importNamespace,
			Name:      sa,
		}
	}

	marshal, err := json.Marshal(newCB)
	if err != nil {
		return nil, err
	}
	newCB.Annotations[kubectlLastAppliedConfig] = string(marshal)

	return newCB, nil
}

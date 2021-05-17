// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"golang.org/x/sync/errgroup"

	"github.com/pivotal/build-service-cli/pkg/config"
)

type StoreRefGetter interface {
	RelocatedBuildpackage(authn.Keychain, config.KpConfig, string) (string, error)
}
type StackRefGetter interface {
	RelocatedBuildImage(authn.Keychain, config.KpConfig, string) (string, error)
	RelocatedRunImage(authn.Keychain, config.KpConfig, string) (string, error)
}

type Differ interface {
	Diff(dOld, dNew interface{}) (string, error)
}

type ImportDiffer struct {
	Differ         Differ
	StoreRefGetter StoreRefGetter
	StackRefGetter StackRefGetter
}

func (id *ImportDiffer) DiffLifecycle(oldImg string, newImg string) (string, error) {
	return id.Differ.Diff(oldImg, newImg)
}

func (id *ImportDiffer) DiffClusterStore(keychain authn.Keychain, kpConfig config.KpConfig, oldCS *v1alpha1.ClusterStore, newCS ClusterStore) (string, error) {
	type void struct{}
	newBPs := map[string]void{}
	mux := &sync.Mutex{}
	errs, _ := errgroup.WithContext(context.Background())

	for _, bp := range newCS.Sources {
		image := bp.Image
		errs.Go(func() error {
			relocatedBP, err := id.StoreRefGetter.RelocatedBuildpackage(keychain, kpConfig, image)
			if err != nil {
				return err
			}
			mux.Lock()
			newBPs[relocatedBP] = void{}
			mux.Unlock()
			return nil
		})
	}

	if err := errs.Wait(); err != nil {
		return "", err
	}

	oldCSStr := ""
	if oldCS != nil {
		for _, s := range oldCS.Spec.Sources {
			delete(newBPs, s.Image)
		}
		oldCSStr = fmt.Sprintf(`Name: %s
Sources:`, oldCS.Name)
	}

	if len(newBPs) == 0 {
		return "", nil
	}

	newCS.Sources = []Source{}
	for img := range newBPs {
		newCS.Sources = append(newCS.Sources, Source{Image: img})
	}

	return id.Differ.Diff(oldCSStr, newCS)
}

func (id *ImportDiffer) DiffClusterStack(keychain authn.Keychain, kpConfig config.KpConfig, oldCS *v1alpha1.ClusterStack, newCS ClusterStack) (diff string, err error) {
	newCS.BuildImage.Image, err = id.StackRefGetter.RelocatedBuildImage(keychain, kpConfig, newCS.BuildImage.Image)
	if err != nil {
		return "", err
	}
	newCS.RunImage.Image, err = id.StackRefGetter.RelocatedRunImage(keychain, kpConfig, newCS.RunImage.Image)
	if err != nil {
		return "", err
	}

	var oldDiffableStack interface{}
	if oldCS != nil {
		oldDiffableStack = ClusterStack{
			Name:       oldCS.Name,
			BuildImage: Source{Image: oldCS.Spec.BuildImage.Image},
			RunImage:   Source{Image: oldCS.Spec.RunImage.Image},
		}
	}

	return id.Differ.Diff(oldDiffableStack, newCS)
}

func (id *ImportDiffer) DiffClusterBuilder(oldCB *v1alpha1.ClusterBuilder, newCB ClusterBuilder) (string, error) {
	var oldDiffableCB interface{}
	if oldCB != nil {
		oldDiffableCB = ClusterBuilder{
			Name:         oldCB.Name,
			ClusterStack: oldCB.Spec.Stack.Name,
			ClusterStore: oldCB.Spec.Store.Name,
			Order:        oldCB.Spec.Order,
		}
	}

	return id.Differ.Diff(oldDiffableCB, newCB)
}

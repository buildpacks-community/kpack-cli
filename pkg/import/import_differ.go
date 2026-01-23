// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"golang.org/x/sync/errgroup"

	"github.com/buildpacks-community/kpack-cli/pkg/config"
)

type RelocatedImageProvider interface {
	RelocatedImage(authn.Keychain, config.KpConfig, string) (string, error)
}

type Differ interface {
	Diff(dOld, dNew interface{}) (string, error)
}

type ImportDiffer struct {
	Differ                 Differ
	RelocatedImageProvider RelocatedImageProvider
}

func (id *ImportDiffer) DiffClusterLifecycle(keychain authn.Keychain, kpConfig config.KpConfig, oldCL *v1alpha2.ClusterLifecycle, newCL ClusterLifecycle) (diff string, err error) {
	newCL.Image, err = id.RelocatedImageProvider.RelocatedImage(keychain, kpConfig, newCL.Image)
	if err != nil {
		return "", err
	}

	var oldDiffableLifecycle interface{}
	if oldCL != nil {
		oldDiffableLifecycle = ClusterLifecycle{
			Name:  oldCL.Name,
			Image: oldCL.Spec.ImageSource.Image,
		}
	}

	return id.Differ.Diff(oldDiffableLifecycle, newCL)
}

func (id *ImportDiffer) DiffClusterBuildpack(keychain authn.Keychain, kpConfig config.KpConfig, oldCBP *v1alpha2.ClusterBuildpack, newCBP ClusterBuildpack) (diff string, err error) {
	newCBP.Image, err = id.RelocatedImageProvider.RelocatedImage(keychain, kpConfig, newCBP.Image)
	if err != nil {
		return "", err
	}

	var oldDiffableBuildpack interface{}
	if oldCBP != nil {
		oldDiffableBuildpack = ClusterBuildpack{
			Name:  oldCBP.Name,
			Image: oldCBP.Spec.ImageSource.Image,
		}
	}

	return id.Differ.Diff(oldDiffableBuildpack, newCBP)
}

func (id *ImportDiffer) DiffClusterStore(keychain authn.Keychain, kpConfig config.KpConfig, oldCS *v1alpha2.ClusterStore, newCS ClusterStore) (string, error) {
	type void struct{}
	newBPs := map[string]void{}
	mux := &sync.Mutex{}
	errs, _ := errgroup.WithContext(context.Background())

	for _, bp := range newCS.Sources {
		image := bp.Image
		errs.Go(func() error {
			relocatedBP, err := id.RelocatedImageProvider.RelocatedImage(keychain, kpConfig, image)
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

func (id *ImportDiffer) DiffClusterStack(keychain authn.Keychain, kpConfig config.KpConfig, oldCS *v1alpha2.ClusterStack, newCS ClusterStack) (diff string, err error) {
	newCS.BuildImage.Image, err = id.RelocatedImageProvider.RelocatedImage(keychain, kpConfig, newCS.BuildImage.Image)
	if err != nil {
		return "", err
	}
	newCS.RunImage.Image, err = id.RelocatedImageProvider.RelocatedImage(keychain, kpConfig, newCS.RunImage.Image)
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

func (id *ImportDiffer) DiffClusterBuilder(oldCB *v1alpha2.ClusterBuilder, newCB ClusterBuilder) (string, error) {
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

// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pkg/errors"

	"github.com/buildpacks-community/kpack-cli/pkg/import/descriptor"
)

const CurrentAPIVersion = descriptor.APIVersionV1

type API struct {
	Version string `yaml:"apiVersion" json:"apiVersion"`
}

// Type aliases to use types from the conversion package
type (
	DependencyDescriptor         = descriptor.DependencyDescriptor
	DependencyDescriptorV1Alpha1 = descriptor.DependencyDescriptorV1Alpha1
	DependencyDescriptorV1Alpha3 = descriptor.DependencyDescriptorV1Alpha3
	Source                       = descriptor.Source
	ClusterLifecycle             = descriptor.ClusterLifecycle
	ClusterBuildpack             = descriptor.ClusterBuildpack
	ClusterStore                 = descriptor.ClusterStore
	ClusterStack                 = descriptor.ClusterStack
	ClusterBuilder               = descriptor.ClusterBuilder
)

func ValidateDescriptor(d DependencyDescriptor) error {
	lifecycleSet := map[string]bool{}
	for _, lifecycle := range d.ClusterLifecycles {
		if lifecycle.Name == "" {
			return errors.New("cluster lifecycle name cannot be empty")
		}
		if lifecycleSet[lifecycle.Name] {
			return errors.Errorf("duplicate cluster lifecycle name '%s'", lifecycle.Name)
		}
		lifecycleSet[lifecycle.Name] = true

		_, err := name.ParseReference(lifecycle.Image, name.WeakValidation)
		if err != nil {
			return err
		}
	}

	buildpackSet := map[string]bool{}
	for _, buildpack := range d.ClusterBuildpacks {
		if buildpack.Name == "" {
			return errors.New("cluster buildpack name cannot be empty")
		}
		if buildpackSet[buildpack.Name] {
			return errors.Errorf("duplicate cluster buildpack name '%s'", buildpack.Name)
		}
		buildpackSet[buildpack.Name] = true

		_, err := name.ParseReference(buildpack.Image, name.WeakValidation)
		if err != nil {
			return err
		}
	}

	storeSet := map[string]bool{}
	for _, store := range d.ClusterStores {
		if storeSet[store.Name] {
			return errors.Errorf("duplicate store name '%s'", store.Name)
		}
		storeSet[store.Name] = true

		for _, src := range store.Sources {
			_, err := name.ParseReference(src.Image, name.WeakValidation)
			if err != nil {
				return err
			}
		}
	}

	stackSet := map[string]bool{}
	for _, stack := range d.ClusterStacks {
		if stackSet[stack.Name] {
			return errors.Errorf("duplicate stack name '%s'", stack.Name)
		}
		stackSet[stack.Name] = true

		_, err := name.ParseReference(stack.BuildImage.Image, name.WeakValidation)
		if err != nil {
			return err
		}

		_, err = name.ParseReference(stack.RunImage.Image, name.WeakValidation)
		if err != nil {
			return err
		}
	}

	if _, ok := lifecycleSet[d.DefaultClusterLifecycle]; !ok && d.DefaultClusterLifecycle != "" {
		return errors.Errorf("default cluster lifecycle '%s' not found", d.DefaultClusterLifecycle)
	}

	if _, ok := buildpackSet[d.DefaultClusterBuildpack]; !ok && d.DefaultClusterBuildpack != "" {
		return errors.Errorf("default cluster buildpack '%s' not found", d.DefaultClusterBuildpack)
	}

	if _, ok := stackSet[d.DefaultClusterStack]; !ok && d.DefaultClusterStack != "" {
		return errors.Errorf("default cluster stack '%s' not found", d.DefaultClusterStack)
	}

	ccbSet := map[string]bool{}
	for _, ccb := range d.ClusterBuilders {
		if ccbSet[ccb.Name] {
			return errors.Errorf("duplicate cluster builder name '%s'", ccb.Name)
		}
		ccbSet[ccb.Name] = true
	}

	if _, ok := ccbSet[d.DefaultClusterBuilder]; !ok && d.DefaultClusterBuilder != "" {
		return errors.Errorf("default cluster builder '%s' not found", d.DefaultClusterBuilder)
	}

	return nil
}

func GetClusterLifecycles(d DependencyDescriptor) []ClusterLifecycle {
	lifecycles := d.ClusterLifecycles
	for _, lc := range d.ClusterLifecycles {
		if lc.Name == d.DefaultClusterLifecycle {
			lifecycles = append(lifecycles, ClusterLifecycle{
				Name:  v1alpha2.DefaultLifecycleName,
				Image: lc.Image,
			})
			break
		}
	}
	return lifecycles
}

func GetClusterBuildpacks(d DependencyDescriptor) []ClusterBuildpack {
	buildpacks := d.ClusterBuildpacks
	for _, bp := range d.ClusterBuildpacks {
		if bp.Name == d.DefaultClusterBuildpack {
			buildpacks = append(buildpacks, ClusterBuildpack{
				Name:  "default",
				Image: bp.Image,
			})
			break
		}
	}
	return buildpacks
}

func GetClusterStores(d DependencyDescriptor) []ClusterStore {
	return d.ClusterStores
}

func GetClusterStacks(d DependencyDescriptor) []ClusterStack {
	stacks := d.ClusterStacks
	for _, stack := range d.ClusterStacks {
		if stack.Name == d.DefaultClusterStack {
			stacks = append(stacks, ClusterStack{
				Name:       "default",
				BuildImage: stack.BuildImage,
				RunImage:   stack.RunImage,
			})
			break
		}
	}
	return stacks
}

func GetClusterBuilders(d DependencyDescriptor) []ClusterBuilder {
	builders := d.ClusterBuilders
	for _, cb := range d.ClusterBuilders {
		if cb.Name == d.DefaultClusterBuilder {
			builders = append(builders, ClusterBuilder{
				Name:         "default",
				ClusterStack: cb.ClusterStack,
				ClusterStore: cb.ClusterStore,
				Order:        cb.Order,
			})
			break
		}
	}
	return builders
}

// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pkg/errors"
)

type DependencyDescriptorV3 struct {
	APIVersion            string           `yaml:"apiVersion" json:"apiVersion"`
	Kind                  string           `yaml:"kind" json:"kind"`
	DefaultClusterStack   string           `yaml:"defaultClusterStack" json:"defaultClusterStack"`
	DefaultClusterBuilder string           `yaml:"defaultClusterBuilder" json:"defaultClusterBuilder"`
	Lifecycle             Lifecycle        `yaml:"lifecycle" json:"lifecycle"`
	ClusterStores         []ClusterStore   `yaml:"clusterStores" json:"clusterStores"`
	ClusterStacks         []ClusterStack   `yaml:"clusterStacks" json:"clusterStacks"`
	ClusterBuilders       []ClusterBuilder `yaml:"clusterBuilders" json:"clusterBuilders"`
}

type Lifecycle Source

type ClusterStore struct {
	Name    string   `yaml:"name" json:"name"`
	Sources []Source `yaml:"sources" json:"sources"`
}

func (d DependencyDescriptorV3) Validate() error {
	storeSet := map[string]interface{}{}
	for _, store := range d.ClusterStores {
		if name, ok := storeSet[store.Name]; ok {
			return errors.Errorf("duplicate store name '%s'", name)
		}
		storeSet[store.Name] = nil

		for _, src := range store.Sources {
			_, err := name.ParseReference(src.Image, name.WeakValidation)
			if err != nil {
				return err
			}
		}
	}

	stackSet := map[string]interface{}{}
	for _, stack := range d.ClusterStacks {
		if name, ok := stackSet[stack.Name]; ok {
			return errors.Errorf("duplicate stack name '%s'", name)
		}
		stackSet[stack.Name] = nil

		_, err := name.ParseReference(stack.BuildImage.Image, name.WeakValidation)
		if err != nil {
			return err
		}

		_, err = name.ParseReference(stack.RunImage.Image, name.WeakValidation)
		if err != nil {
			return err
		}
	}

	if _, ok := stackSet[d.DefaultClusterStack]; !ok && d.DefaultClusterStack != "" {
		return errors.Errorf("default cluster stack '%s' not found", d.DefaultClusterStack)
	}

	ccbSet := map[string]interface{}{}
	for _, ccb := range d.ClusterBuilders {
		if name, ok := ccbSet[ccb.Name]; ok {
			return errors.Errorf("duplicate cluster builder name '%s'", name)
		}
		ccbSet[ccb.Name] = nil
	}

	if _, ok := ccbSet[d.DefaultClusterBuilder]; !ok && d.DefaultClusterBuilder != "" {
		return errors.Errorf("default cluster builder '%s' not found", d.DefaultClusterBuilder)
	}

	return nil
}

func (d DependencyDescriptorV3) GetLifecycleImage() string {
	return d.Lifecycle.Image
}

func (d DependencyDescriptorV3) HasLifecycleImage() bool {
	return d.Lifecycle.Image != ""
}

func (d DependencyDescriptorV3) GetClusterStacks() []ClusterStack {
	for _, stack := range d.ClusterStacks {
		if stack.Name == d.DefaultClusterStack {
			d.ClusterStacks = append(d.ClusterStacks, ClusterStack{
				Name:       "default",
				BuildImage: stack.BuildImage,
				RunImage:   stack.RunImage,
			})
			break
		}
	}
	return d.ClusterStacks
}

func (d DependencyDescriptorV3) GetClusterBuilders() []ClusterBuilder {
	for _, cb := range d.ClusterBuilders {
		if cb.Name == d.DefaultClusterBuilder {
			d.ClusterBuilders = append(d.ClusterBuilders, ClusterBuilder{
				Name:         "default",
				ClusterStack: cb.ClusterStack,
				ClusterStore: cb.ClusterStore,
				Order:        cb.Order,
			})
			break
		}
	}
	return d.ClusterBuilders
}

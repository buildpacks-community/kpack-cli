// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
)

const CurrentAPIVersion = "kp.kpack.io/v1alpha2"

type API struct {
	Version string `yaml:"apiVersion" json:"apiVersion"`
}

type DependencyDescriptor struct {
	APIVersion            string           `yaml:"apiVersion"`
	Kind                  string           `yaml:"kind"`
	DefaultClusterStack   string           `yaml:"defaultClusterStack"`
	DefaultClusterBuilder string           `yaml:"defaultClusterBuilder"`
	ClusterStores         []ClusterStore   `yaml:"clusterStores"`
	ClusterStacks         []ClusterStack   `yaml:"clusterStacks"`
	ClusterBuilders       []ClusterBuilder `yaml:"clusterBuilders"`
}

type ClusterStore struct {
	Name    string   `yaml:"name"`
	Sources []Source `yaml:"sources"`
}

type Source struct {
	Image string `yaml:"image"`
}

type ClusterStack struct {
	Name       string `yaml:"name"`
	BuildImage Source `yaml:"buildImage"`
	RunImage   Source `yaml:"runImage"`
}

type ClusterBuilder struct {
	Name         string                `yaml:"name"`
	ClusterStack string                `yaml:"clusterStack"`
	ClusterStore string                `yaml:"clusterStore"`
	Order        []v1alpha1.OrderEntry `yaml:"order"`
}

func (d DependencyDescriptor) Validate() error {
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

	if _, ok := stackSet[d.DefaultClusterStack]; !ok {
		return errors.Errorf("default cluster stack '%s' not found", d.DefaultClusterStack)
	}

	ccbSet := map[string]interface{}{}
	for _, ccb := range d.ClusterBuilders {
		if name, ok := ccbSet[ccb.Name]; ok {
			return errors.Errorf("duplicate cluster builder name '%s'", name)
		}
		ccbSet[ccb.Name] = nil

		if _, ok := storeSet[ccb.ClusterStore]; !ok {
			return errors.Errorf("cluster builder '%s' references unknown cluster store '%s'", ccb.Name, ccb.ClusterStore)
		}

		if _, ok := stackSet[ccb.ClusterStack]; !ok {
			return errors.Errorf("cluster builder '%s' references unknown cluster stack '%s'", ccb.Name, ccb.ClusterStack)
		}
	}

	if _, ok := ccbSet[d.DefaultClusterBuilder]; !ok {
		return errors.Errorf("default cluster builder '%s' not found", d.DefaultClusterBuilder)
	}

	return nil
}

func (d DependencyDescriptor) GetClusterStacks() []ClusterStack {
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

func (d DependencyDescriptor) GetClusterBuilders() []ClusterBuilder {
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

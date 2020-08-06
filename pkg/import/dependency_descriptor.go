// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
)

type DependencyDescriptor struct {
	APIVersion            string           `yaml:"apiVersion"`
	Kind                  string           `yaml:"kind"`
	DefaultStack          string           `yaml:"defaultStack"`
	DefaultClusterBuilder string           `yaml:"defaultClusterBuilder"`
	Stores                []Store          `yaml:"stores"`
	Stacks                []Stack          `yaml:"stacks"`
	ClusterBuilders       []ClusterBuilder `yaml:"clusterBuilders"`
}

type Store struct {
	Name    string   `yaml:"name"`
	Sources []Source `yaml:"sources"`
}

type Source struct {
	Image string `yaml:"image"`
}

type Stack struct {
	Name       string `yaml:"name"`
	BuildImage Source `yaml:"buildImage"`
	RunImage   Source `yaml:"runImage"`
}

type ClusterBuilder struct {
	Name  string                `yaml:"name"`
	Stack string                `yaml:"stack"`
	Store string                `yaml:"store"`
	Order []v1alpha1.OrderEntry `yaml:"order"`
}

func (d DependencyDescriptor) Validate() error {
	storeSet := map[string]interface{}{}
	for _, store := range d.Stores {
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
	for _, stack := range d.Stacks {
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

	if _, ok := stackSet[d.DefaultStack]; !ok {
		return errors.Errorf("default stack '%s' not found", d.DefaultStack)
	}

	ccbSet := map[string]interface{}{}
	for _, ccb := range d.ClusterBuilders {
		if name, ok := ccbSet[ccb.Name]; ok {
			return errors.Errorf("duplicate cluster builder name '%s'", name)
		}
		ccbSet[ccb.Name] = nil

		if _, ok := storeSet[ccb.Store]; !ok {
			return errors.Errorf("cluster builder '%s' references unknown store '%s'", ccb.Name, ccb.Store)
		}

		if _, ok := stackSet[ccb.Stack]; !ok {
			return errors.Errorf("cluster builder '%s' references unknown stack '%s'", ccb.Name, ccb.Stack)
		}
	}

	if _, ok := ccbSet[d.DefaultClusterBuilder]; !ok {
		return errors.Errorf("default cluster builder '%s' not found", d.DefaultClusterBuilder)
	}

	return nil
}

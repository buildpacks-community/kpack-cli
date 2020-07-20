// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import (
	"github.com/google/go-containerregistry/pkg/name"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pkg/errors"
)

type DependencyDescriptor struct {
	APIVersion                  string                 `yaml:"apiVersion"`
	Kind                        string                 `yaml:"kind"`
	DefaultStack                string                 `yaml:"defaultStack"`
	DefaultCustomClusterBuilder string                 `yaml:"defaultCustomClusterBuilder"`
	Stores                      []Store                `yaml:"stores"`
	Stacks                      []Stack                `yaml:"stacks"`
	CustomClusterBuilders       []CustomClusterBuilder `yaml:"customClusterBuilders"`
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

type CustomClusterBuilder struct {
	Name  string                   `yaml:"name"`
	Stack string                   `yaml:"stack"`
	Store string                   `yaml:"store"`
	Order []expv1alpha1.OrderEntry `yaml:"order"`
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
	for _, ccb := range d.CustomClusterBuilders {
		if name, ok := ccbSet[ccb.Name]; ok {
			return errors.Errorf("duplicate custom cluster builder name '%s'", name)
		}
		ccbSet[ccb.Name] = nil

		if _, ok := storeSet[ccb.Store]; !ok {
			return errors.Errorf("custom cluster builder '%s' references unknown store '%s'", ccb.Name, ccb.Store)
		}

		if _, ok := stackSet[ccb.Stack]; !ok {
			return errors.Errorf("custom cluster builder '%s' references unknown stack '%s'", ccb.Name, ccb.Stack)
		}
	}

	if _, ok := ccbSet[d.DefaultCustomClusterBuilder]; !ok {
		return errors.Errorf("default custom cluster builder '%s' not found", d.DefaultCustomClusterBuilder)
	}

	return nil
}

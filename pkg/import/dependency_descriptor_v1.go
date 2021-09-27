// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"

const APIVersionV1 = "kp.kpack.io/v1alpha1"

type DependencyDescriptorV1 struct {
	APIVersion            string             `yaml:"apiVersion"`
	Kind                  string             `yaml:"kind"`
	DefaultStack          string             `yaml:"defaultStack"`
	DefaultClusterBuilder string             `yaml:"defaultClusterBuilder"`
	Stores                []ClusterStore     `yaml:"stores"`
	Stacks                []ClusterStack     `yaml:"stacks"`
	ClusterBuilders       []ClusterBuilderV1 `yaml:"clusterBuilders"`
}

type ClusterBuilderV1 struct {
	Name  string                `yaml:"name"`
	Stack string                `yaml:"stack"`
	Store string                `yaml:"store"`
	Order []v1alpha1.OrderEntry `yaml:"order"`
}

func (d1 DependencyDescriptorV1) ToNextVersion() DependencyDescriptor {
	var d DependencyDescriptor
	d.APIVersion = d1.APIVersion
	d.Kind = d1.Kind
	d.DefaultClusterStack = d1.DefaultStack
	d.DefaultClusterBuilder = d1.DefaultClusterBuilder
	d.ClusterStores = d1.Stores
	d.ClusterStacks = d1.Stacks
	for _, cb := range d1.ClusterBuilders {
		d.ClusterBuilders = append(d.ClusterBuilders, ClusterBuilder{
			Name:         cb.Name,
			ClusterStack: cb.Stack,
			ClusterStore: cb.Store,
			Order:        cb.Order,
		})
	}
	return d
}

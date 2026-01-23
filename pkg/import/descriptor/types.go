// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package conversion provides types and functions to convert older API versions
// of dependency descriptors (v1alpha1, v1alpha3) to the current v1 format.
package descriptor

import (
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
)

// Source represents an image source
type Source struct {
	Image string `yaml:"image"`
}

// ClusterStore represents a ClusterStore in the descriptor
type ClusterStore struct {
	Name    string   `yaml:"name" json:"name"`
	Sources []Source `yaml:"sources" json:"sources"`
}

// ClusterStack represents a ClusterStack in the descriptor
type ClusterStack struct {
	Name       string `yaml:"name" json:"name"`
	BuildImage Source `yaml:"buildImage" json:"buildImage"`
	RunImage   Source `yaml:"runImage" json:"runImage"`
}

// ClusterBuilder represents a ClusterBuilder in the descriptor (v1alpha3+)
type ClusterBuilder struct {
	Name         string                       `yaml:"name" json:"name"`
	ClusterStack string                       `yaml:"clusterStack" json:"clusterStack"`
	ClusterStore string                       `yaml:"clusterStore" json:"clusterStore"`
	Order        []v1alpha2.BuilderOrderEntry `yaml:"order" json:"order"`
}

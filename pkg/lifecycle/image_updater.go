// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"context"
	"io"
	"path"

	ggcrv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	buildk8s "github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/registry"
)

const (
	lifecycleImageName     = "lifecycle"
	lifecycleMetadataLabel = "io.buildpacks.lifecycle.metadata"
)

type PreUpdateHook func(*corev1.ConfigMap)

type ImageUpdaterConfig struct {
	DryRun       bool
	IOWriter     io.Writer
	ImgFetcher   registry.Fetcher
	ImgRelocator registry.Relocator
	ClientSet    buildk8s.ClientSet
	TLSConfig    registry.TLSConfig
}

func UpdateImage(ctx context.Context, srcImgLocation string, cfg ImageUpdaterConfig, hooks ...PreUpdateHook) (*corev1.ConfigMap, error) {
	cm, err := getConfigMap(ctx, cfg.ClientSet.K8sClient)
	if err != nil {
		return cm, err
	}

	img, err := cfg.ImgFetcher.Fetch(srcImgLocation, cfg.TLSConfig)
	if err != nil {
		return cm, err
	}

	if err = validateImage(img); err != nil {
		return cm, err
	}

	relocatedImgTag, err := relocateImageToCanonicalRepo(ctx, img, cfg)
	if err != nil {
		return cm, err
	}

	cm.Data[lifecycleImageKey] = relocatedImgTag

	for _, h := range hooks {
		h(cm)
	}

	if !cfg.DryRun {
		cm, err = updateConfigMap(ctx, cm, cfg.ClientSet.K8sClient)
	}
	return cm, err
}

func validateImage(img ggcrv1.Image) error {
	hasLabel, err := imagehelpers.HasLabel(img, lifecycleMetadataLabel)
	if err != nil {
		return err
	}

	if !hasLabel {
		return errors.New("image missing lifecycle metadata")
	}
	return nil
}

func relocateImageToCanonicalRepo(ctx context.Context, img ggcrv1.Image, cfg ImageUpdaterConfig) (string, error) {
	canonicalRepo, err := buildk8s.DefaultConfigHelper(cfg.ClientSet).GetCanonicalRepository(ctx)
	if err != nil {
		return "", err
	}

	dstImgLocation := path.Join(canonicalRepo, lifecycleImageName)
	return cfg.ImgRelocator.Relocate(img, dstImgLocation, cfg.IOWriter, cfg.TLSConfig)
}

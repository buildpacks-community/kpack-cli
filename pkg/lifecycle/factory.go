// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"context"
	"fmt"
	"path"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
)

const (
	lifecycleImageName     = "lifecycle"
	lifecycleMetadataLabel = "io.buildpacks.lifecycle.metadata"
)

type PreUpdateHook func(*corev1.ConfigMap)

type Factory struct {
	Relocator registry.Relocator
	Fetcher   registry.Fetcher
	ClientSet k8s.Interface
}

func NewFactory(relocator registry.Relocator, fetcher registry.Fetcher, clientSet k8s.Interface) *Factory {
	return &Factory{
		Relocator: relocator,
		Fetcher:   fetcher,
		ClientSet: clientSet,
	}
}

func (f *Factory) UpdateLifecycle(ctx context.Context, keychain authn.Keychain, kpConfig config.KpConfig, dryRun bool, img string, hooks ...PreUpdateHook) (*corev1.ConfigMap, error) {
	cm, err := getConfigMap(ctx, f.ClientSet)
	if err != nil {
		return cm, err
	}

	image, err := f.Fetcher.Fetch(keychain, img)
	if err != nil {
		return cm, err
	}

	if err = validateImage(image); err != nil {
		return cm, err
	}

	relocatedImgTag, err := f.relocateLifecycleImage(keychain, kpConfig, image)
	if err != nil {
		return cm, err
	}

	cm.Data[lifecycleImageKey] = relocatedImgTag

	for _, h := range hooks {
		h(cm)
	}

	if !dryRun {
		cm, err = updateConfigMap(ctx, cm, f.ClientSet)
	}
	return cm, err
}

func (f *Factory) relocateLifecycleImage(keychain authn.Keychain, kpConfig config.KpConfig, image v1.Image) (string, error) {
	dstImgLocation, err := getLifecycleRef(kpConfig)
	if err != nil {
		return "", err
	}

	return f.Relocator.Relocate(keychain, image, dstImgLocation)
}

func (f *Factory) RelocatedLifecycleImage(keychain authn.Keychain, kpConfig config.KpConfig, img string) (string, error) {
	image, err := f.Fetcher.Fetch(keychain, img)
	if err != nil {
		return "", err
	}

	digest, err := image.Digest()
	if err != nil {
		return "", err
	}

	relocatedLifecyclePath, err := getLifecycleRef(kpConfig)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s@%s", relocatedLifecyclePath, digest.String()), nil
}

func getLifecycleRef(kpConfig config.KpConfig) (string, error) {
	canonicalRepo, err := kpConfig.CanonicalRepository()
	if err != nil {
		return "", err
	}

	dstImgLocation := path.Join(canonicalRepo, lifecycleImageName)
	return dstImgLocation, nil
}

func validateImage(img v1.Image) error {
	hasLabel, err := imagehelpers.HasLabel(img, lifecycleMetadataLabel)
	if err != nil {
		return err
	}

	if !hasLabel {
		return errors.New("image missing lifecycle metadata")
	}
	return nil
}

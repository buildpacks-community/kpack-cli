// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"context"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigHelper interface {
	GetCanonicalRepository(ctx context.Context) (string, error)
	GetCanonicalServiceAccount(ctx context.Context) (string, error)
}

const (
	kpNamespace                = "kpack"
	kpConfigMapName            = "kp-config"
	canonicalRepositoryKey     = "canonical.repository"
	canonicalServiceAccountKey = "canonical.repository.serviceaccount"
)

type defaultConfigHelper struct {
	cs ClientSet
}

func DefaultConfigHelper(cs ClientSet) ConfigHelper {
	return defaultConfigHelper{cs: cs}
}

func (d defaultConfigHelper) GetCanonicalRepository(ctx context.Context) (string, error) {
	val, err := d.getValue(ctx, canonicalRepositoryKey)
	if err != nil {
		return val, errors.Wrapf(err, "failed to get canonical repository")
	}

	if val == "" {
		return "", errors.Errorf("failed to get canonical repository")
	}
	return val, err
}

func (d defaultConfigHelper) GetCanonicalServiceAccount(ctx context.Context) (string, error) {
	val, err := d.getValue(ctx, canonicalServiceAccountKey)
	if err != nil {
		return val, errors.Wrapf(err, "failed to get canonical service account")
	}

	if val == "" {
		return "", errors.Errorf("failed to get canonical service account")
	}
	return val, err
}

func (d defaultConfigHelper) getValue(ctx context.Context, key string) (string, error) {
	var value string

	kpConfig, err := d.cs.K8sClient.CoreV1().ConfigMaps(kpNamespace).Get(ctx, kpConfigMapName, metav1.GetOptions{})
	if err != nil {
		return value, err
	}

	value, ok := kpConfig.Data[key]
	if !ok {
		return value, errors.Errorf("key %q not found in configmap %q", key, kpConfigMapName)
	}

	return value, nil
}

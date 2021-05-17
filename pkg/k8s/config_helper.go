// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/config"
)

type ConfigHelper interface {
	GetCanonicalRepository(ctx context.Context) (string, error)
	GetCanonicalServiceAccount(ctx context.Context) (string, error)
	GetKpConfig(ctx context.Context) (config.KpConfig, error)
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

func (d defaultConfigHelper) GetKpConfig(ctx context.Context) (config.KpConfig, error) {
	kpConfig, err := d.cs.K8sClient.CoreV1().ConfigMaps(kpNamespace).Get(ctx, kpConfigMapName, metav1.GetOptions{})
	if err != nil {
		return config.KpConfig{}, err
	}

	repository, ok := kpConfig.Data[canonicalRepositoryKey]
	if !ok {
		return config.KpConfig{}, errors.Errorf("key %q not found in configmap %q", canonicalRepositoryKey, kpConfigMapName)
	}

	serviceAccount, ok := kpConfig.Data[canonicalServiceAccountKey]
	if !ok {
		return config.KpConfig{}, errors.Errorf("key %q not found in configmap %q", canonicalServiceAccountKey, kpConfigMapName)
	}

	return config.KpConfig{CanonicalRepository: repository, ServiceAccount: corev1.ObjectReference{Name: serviceAccount, Namespace: kpNamespace}}, err
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

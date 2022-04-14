// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"context"

	"github.com/pkg/errors"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const (
	lifecycleNamespace     = "kpack"
	lifecycleConfigMapName = "lifecycle-image"
	lifecycleImageKey      = "image"
)

func GetImage(ctx context.Context, c kubernetes.Interface) (string, error) {
	cm, err := getConfigMap(ctx, c)
	if err != nil {
		return "", err
	}
	return cm.Data[lifecycleImageKey], err
}

func getConfigMap(ctx context.Context, c kubernetes.Interface) (*v1.ConfigMap, error) {
	cm, err := c.CoreV1().ConfigMaps(lifecycleNamespace).Get(ctx, lifecycleConfigMapName, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		err = errors.Errorf("configmap %q not found in %q namespace", lifecycleConfigMapName, lifecycleNamespace)
	}
	return cm, err
}

func patchConfigMap(ctx context.Context, cm *v1.ConfigMap, c kubernetes.Interface) (*v1.ConfigMap, error) {
	existingCM, err := getConfigMap(ctx, c)
	if err != nil {
		return nil, err
	}

	patch, err := k8s.CreatePatch(existingCM, cm)
	if err != nil {
		return nil, err
	}

	return c.CoreV1().ConfigMaps(lifecycleNamespace).Patch(ctx, cm.Name, types.MergePatchType, patch, metav1.PatchOptions{})
}

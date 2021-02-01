// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

const (
	lifecycleNamespace     = "kpack"
	lifecycleConfigMapName = "lifecycle-image"
	lifecycleImageKey      = "image"
)

func GetImage(c k8s.Interface) (string, error) {
	cm, err := getConfigMap(c)
	if err != nil {
		return "", err
	}
	return cm.Data[lifecycleImageKey], err
}

func getConfigMap(c k8s.Interface) (*v1.ConfigMap, error) {
	cm, err := c.CoreV1().ConfigMaps(lifecycleNamespace).Get(lifecycleConfigMapName, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		err = errors.Errorf("configmap %q not found in %q namespace", lifecycleConfigMapName, lifecycleNamespace)
	}
	return cm, err
}

func updateConfigMap(cm *v1.ConfigMap, c k8s.Interface) (*v1.ConfigMap, error) {
	return c.CoreV1().ConfigMaps(lifecycleNamespace).Update(cm)
}

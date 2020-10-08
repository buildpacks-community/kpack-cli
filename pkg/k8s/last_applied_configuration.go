// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type Annotatable interface {
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
	runtime.Object
}

const (
	kubectlLastAppliedConfig = "kubectl.kubernetes.io/last-applied-configuration"
)

func SetLastAppliedCfg(obj Annotatable) error {
	if reflect.ValueOf(obj).Kind() != reflect.Ptr {
		return errors.Errorf("last applied configuration can only be set on a pointer object")
	}

	cfg, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	a := obj.GetAnnotations()
	if a == nil {
		a = map[string]string{}
	}

	a[kubectlLastAppliedConfig] = string(cfg)
	obj.SetAnnotations(a)

	return nil
}

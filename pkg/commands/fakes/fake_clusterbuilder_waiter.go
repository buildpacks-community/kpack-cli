// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"k8s.io/apimachinery/pkg/runtime"
	watchTools "k8s.io/client-go/tools/watch"
)

type WaitCall struct {
	Object      runtime.Object
	ExtraChecks []watchTools.ConditionFunc
}

type FakeWaiter struct {
	WaitCalls []WaitCall
}

func (f *FakeWaiter) Wait(ob runtime.Object, checks ...watchTools.ConditionFunc) error {
	f.WaitCalls = append(f.WaitCalls, WaitCall{
		Object:      ob,
		ExtraChecks: checks,
	})
	return nil
}

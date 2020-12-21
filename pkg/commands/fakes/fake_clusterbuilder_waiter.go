// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type BuilderWaitCall struct {
	Object               runtime.Object
	CStoreGen, CStackGen int64
}

type FakeWaiter struct {
	WaitCalls        []runtime.Object
	BuilderWaitCalls []BuilderWaitCall
}

func (f *FakeWaiter) Wait(ob runtime.Object) error {
	f.WaitCalls = append(f.WaitCalls, ob)
	return nil
}

func (f *FakeWaiter) BuilderWait(ob runtime.Object, cStoreGen, cStackgen int64) error {
	f.BuilderWaitCalls = append(f.BuilderWaitCalls, BuilderWaitCall{
		Object:    ob,
		CStoreGen: cStoreGen,
		CStackGen: cStackgen,
	})
	return nil
}

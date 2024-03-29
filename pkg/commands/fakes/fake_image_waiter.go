// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"context"
	"io"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
)

type FakeImageWaiter struct {
	Calls []*v1alpha2.Image
}

func (f *FakeImageWaiter) Wait(ctx context.Context, writer io.Writer, image *v1alpha2.Image) (string, error) {
	f.Calls = append(f.Calls, image)
	return "", nil
}

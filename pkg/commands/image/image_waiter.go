// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"io"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
)

type ImageWaiter interface {
	Wait(ctx context.Context, writer io.Writer, image *v1alpha1.Image) (string, error)
}

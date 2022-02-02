package kpackcompat

import (
    "context"
    "io"

    "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
    "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
)

type v1alpha1Waiter interface {
    Wait(ctx context.Context, writer io.Writer, image *v1alpha1.Image) (string, error)
}

type wrapper struct {
    v1alpha1Waiter v1alpha1Waiter
}

func (w *wrapper) Wait(ctx context.Context, writer io.Writer, image *v1alpha2.Image) (string, error) {
    v1Image, err := convertToV1Image(ctx, image)
    if err != nil {
        return "", err
    }

    return w.v1alpha1Waiter.Wait(ctx, writer, v1Image)
}

func NewImageWaiterForV1alpha2(v1alpha1Waiter v1alpha1Waiter) *wrapper {
    return &wrapper{v1alpha1Waiter}
}

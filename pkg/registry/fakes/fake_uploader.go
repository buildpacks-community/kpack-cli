// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"fmt"
	"io"

	"github.com/pivotal/build-service-cli/pkg/registry"
)

type SourceUploader struct {
	imageRef string
	skip     bool
}

func NewSourceUploader(imageRef string) *SourceUploader {
	return &SourceUploader{
		imageRef: imageRef,
	}
}

func (f *SourceUploader) Upload(_, _ string, writer io.Writer, _ registry.TLSConfig) (string, error) {
	var message string
	if f.skip {
		message = fmt.Sprintf("\tSkipping '%s'\n", f.imageRef)
	} else {
		message = fmt.Sprintf("\tUploading '%s'\n", f.imageRef)
	}

	_, err := writer.Write([]byte(message))
	return f.imageRef, err
}

func (f *SourceUploader) SetImageRef(ref string) {
	f.imageRef = ref
}

func (f *SourceUploader) SetSkipUpload(skip bool) {
	f.skip = skip
}

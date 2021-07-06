// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"fmt"
	"io"

	"github.com/google/go-containerregistry/pkg/authn"
)

type SourceUploader struct {
	changeState bool
	writer      io.Writer
}

func NewFakeSourceUploader(writer io.Writer, changeState bool) *SourceUploader {
	return &SourceUploader{
		writer:      writer,
		changeState: changeState,
	}
}

func (f *SourceUploader) Upload(keychain authn.Keychain, dstImgRefStr, srcPath string) (string, error) {
	uploadPath := fmt.Sprintf("%s:source-id", dstImgRefStr)
	var message string
	if !f.changeState {
		message = fmt.Sprintf("\tSkipping '%s'\n", uploadPath)
	} else {
		message = fmt.Sprintf("\tUploading '%s'\n", uploadPath)
	}

	_, err := f.writer.Write([]byte(message))
	return uploadPath, err
}

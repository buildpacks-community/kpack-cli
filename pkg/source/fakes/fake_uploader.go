// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

type SourceUploader struct {
	ImageRef string
}

func (f *SourceUploader) Upload(_, _ string) (string, error) {
	return f.ImageRef, nil
}

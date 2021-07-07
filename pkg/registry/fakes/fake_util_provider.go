// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"io"

	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
)

type UtilProvider struct {
	FakeFetcher registry.Fetcher
}

func (u UtilProvider) Relocator(writer io.Writer, _ registry.TLSConfig, changeState bool) registry.Relocator {
	return &Relocator{
		skip:   !changeState,
		writer: writer,
	}
}

func (u UtilProvider) Fetcher(_ registry.TLSConfig) registry.Fetcher {
	return u.FakeFetcher
}

func (u UtilProvider) SourceUploader(writer io.Writer, tlsConfig registry.TLSConfig, changeState bool) registry.SourceUploader {
	return NewFakeSourceUploader(writer, changeState)
}

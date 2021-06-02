// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package fakes

import (
	"io"

	"github.com/pivotal/build-service-cli/pkg/registry"
)

type UtilProvider struct {
	FakeFetcher        registry.Fetcher
	FakeRelocator      *Relocator
	FakeSourceUploader registry.SourceUploader
}

func (u UtilProvider) Relocator(writer io.Writer, _ registry.TLSConfig, _ bool) registry.Relocator {
	u.FakeRelocator.SetWriter(writer)
	return u.FakeRelocator
}

func (u UtilProvider) Fetcher(_ registry.TLSConfig) registry.Fetcher {
	return u.FakeFetcher
}

func (u UtilProvider) SourceUploader(_ bool) registry.SourceUploader {
	return u.FakeSourceUploader
}

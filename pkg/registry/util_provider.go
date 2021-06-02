// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package registry

import "io"

type UtilProvider interface {
	Relocator(writer io.Writer, tlsCfg TLSConfig, changeState bool) Relocator
	SourceUploader(changeState bool) SourceUploader
	Fetcher(config TLSConfig) Fetcher
}

type DefaultUtilProvider struct{}

func (d DefaultUtilProvider) Relocator(writer io.Writer, tlsCfg TLSConfig, changeState bool) Relocator {
	if changeState {
		return NewDefaultRelocator(writer, tlsCfg)
	} else {
		return NewDiscardRelocator(writer)
	}
}

func (d DefaultUtilProvider) SourceUploader(changeState bool) SourceUploader {
	if changeState {
		return DefaultSourceUploader{}
	} else {
		return DiscardSourceUploader{}
	}
}

func (d DefaultUtilProvider) Fetcher(config TLSConfig) Fetcher {
	return NewDefaultFetcher(config)
}

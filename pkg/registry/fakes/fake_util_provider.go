package fakes

import "github.com/pivotal/build-service-cli/pkg/registry"

type UtilProvider struct {
	FakeFetcher        registry.Fetcher
	FakeRelocator      registry.Relocator
	FakeSourceUploader registry.SourceUploader
}

func (u UtilProvider) Fetcher() registry.Fetcher {
	return u.FakeFetcher
}

func (u UtilProvider) Relocator(changeState bool) registry.Relocator {
	return u.FakeRelocator
}

func (u UtilProvider) SourceUploader(changeState bool) registry.SourceUploader {
	return u.FakeSourceUploader
}

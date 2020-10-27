package registry

type UtilProvider interface {
	Relocator(changeState bool) Relocator
	SourceUploader(changeState bool) SourceUploader
	Fetcher() Fetcher
}

type DefaultUtilProvider struct{}

func (d DefaultUtilProvider) Relocator(changeState bool) Relocator {
	if changeState {
		return DefaultRelocator{}
	} else {
		return DiscardRelocator{}
	}
}

func (d DefaultUtilProvider) SourceUploader(changeState bool) SourceUploader {
	if changeState {
		return DefaultSourceUploader{}
	} else {
		return DiscardSourceUploader{}
	}
}

func (d DefaultUtilProvider) Fetcher() Fetcher {
	return DefaultFetcher{}
}

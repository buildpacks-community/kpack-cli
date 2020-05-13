package fakes

type SourceUploader struct {
	ImageRef string
}

func (f *SourceUploader) Upload(_, _ string) (string, error) {
	return f.ImageRef, nil
}

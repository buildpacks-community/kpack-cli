package _import

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"

	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/registry"
)

type DefaultRelocatedImageProvider struct {
	fetcher registry.Fetcher
}

func NewDefaultRelocatedImageProvider(fetcher registry.Fetcher) *DefaultRelocatedImageProvider {
	return &DefaultRelocatedImageProvider{fetcher: fetcher}
}

func (r *DefaultRelocatedImageProvider) RelocatedImage(keychain authn.Keychain, kpConfig config.KpConfig, srcImage string) (string, error) {
	relocationRepo, err := kpConfig.DefaultRepository()
	if err != nil {
		return "", err
	}

	repository, err := name.NewRepository(relocationRepo)
	if err != nil {
		return "", err
	}

	img, err := r.fetcher.Fetch(keychain, srcImage)
	if err != nil {
		return "", err
	}

	digest, err := img.Digest()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s@%s", repository, digest), nil
}

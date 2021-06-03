// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"io"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type Relocator interface {
	Relocate(keychain authn.Keychain, src v1.Image, destination string) (string, error)
}

type DiscardRelocator struct {
	writer io.Writer
}

func NewDiscardRelocator(writer io.Writer) DiscardRelocator {
	return DiscardRelocator{writer: writer}
}

func (d DiscardRelocator) Relocate(keychain authn.Keychain, src v1.Image, destination string) (string, error) {
	cfg, err := getDstImageInfo(src, destination)
	if err != nil {
		return "", err
	}

	_, err = d.writer.Write([]byte(fmt.Sprintf("\tSkipping '%s'\n", cfg.refDigestStr)))
	return cfg.refDigestStr, err
}

type DefaultRelocator struct {
	tlsCfg TLSConfig
	writer io.Writer
}

func NewDefaultRelocator(writer io.Writer, tlsCfg TLSConfig) DefaultRelocator {
	return DefaultRelocator{writer: writer, tlsCfg: tlsCfg}
}

func (d DefaultRelocator) Relocate(keychain authn.Keychain, src v1.Image, destination string) (string, error) {
	cfg, err := getDstImageInfo(src, destination)
	if err != nil {
		return "", err
	}

	if _, err := d.writer.Write([]byte(fmt.Sprintf("\tUploading '%s'", cfg.refDigestStr))); err != nil {
		return cfg.refDigestStr, err
	}

	spinner := newUploadSpinner(d.writer, cfg.size)
	defer spinner.Stop()
	go spinner.Write()

	transport, err := d.tlsCfg.Transport()
	if err != nil {
		return cfg.refDigestStr, err
	}
	imgWriteOptions := []remote.Option{
		remote.WithAuthFromKeychain(keychain),
		remote.WithTransport(transport),
	}

	err = remote.Write(cfg.refRepo, src, imgWriteOptions...)
	if err != nil {
		return cfg.refDigestStr, newImageAccessError(cfg.refRepo.Context().RegistryStr(), err)
	}

	return cfg.refDigestStr, remote.Tag(cfg.tag, src, imgWriteOptions...)
}

type relocateImageInfo struct {
	refRepo      name.Reference
	refDigestStr string
	tag          name.Tag
	size         int64
}

func getDstImageInfo(srcImage v1.Image, dstRepoStr string) (relocateImageInfo, error) {
	imgInfo := relocateImageInfo{}

	refDstRepo, err := name.ParseReference(dstRepoStr, name.WeakValidation)
	if err != nil {
		return imgInfo, err
	}

	refContext := refDstRepo.Context()
	refName := fmt.Sprintf("%s/%s", refContext.RegistryStr(), refContext.RepositoryStr())

	refDstRepo, err = name.ParseReference(refName, name.WeakValidation)
	if err != nil {
		return imgInfo, err
	}

	digest, err := srcImage.Digest()
	if err != nil {
		return imgInfo, err
	}

	size, err := imageSize(srcImage)
	if err != nil {
		return imgInfo, err
	}

	imgInfo = relocateImageInfo{
		refRepo:      refDstRepo,
		refDigestStr: fmt.Sprintf("%s@%s", refDstRepo, digest),
		tag:          refDstRepo.Context().Tag(timestampTag()),
		size:         size,
	}
	return imgInfo, err
}

func timestampTag() string {
	now := time.Now()
	return fmt.Sprintf("%s%02d%02d%02d", now.Format("20060102"), now.Hour(), now.Minute(), now.Second())
}

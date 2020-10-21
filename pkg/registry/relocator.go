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

type relocateCfg struct {
	imgInfo         relocateImageInfo
	imgWriteOptions []remote.Option
}

type relocateImageInfo struct {
	refRepo      name.Reference
	refDigestStr string
	tag          name.Tag
	size         int64
}

type Relocator struct{}

func (r Relocator) Relocate(srcImage v1.Image, dstRepoStr string, writer io.Writer, tlsCfg TLSConfig) (string, error) {
	cfg, err := getRelocateCfg(srcImage, dstRepoStr, tlsCfg)
	if err != nil {
		return "", err
	}

	if _, err := writer.Write([]byte(fmt.Sprintf("\tUploading '%s'", cfg.imgInfo.refDigestStr))); err != nil {
		return "", err
	}

	spinner := newUploadSpinner(writer, cfg.imgInfo.size)
	defer spinner.Stop()
	go spinner.Write()

	err = remote.Write(cfg.imgInfo.refRepo, srcImage, cfg.imgWriteOptions...)
	if err != nil {
		return cfg.imgInfo.refDigestStr, newImageAccessError(cfg.imgInfo.refRepo.Context().RegistryStr(), err)
	}

	err = remote.Tag(cfg.imgInfo.tag, srcImage, cfg.imgWriteOptions...)
	return cfg.imgInfo.refDigestStr, err
}

func getRelocateCfg(srcImage v1.Image, dstRepoStr string, tlsCfg TLSConfig) (relocateCfg, error) {
	var cfg relocateCfg

	imgInfo, err := getDstImageInfo(srcImage, dstRepoStr)
	if err != nil {
		return cfg, err
	}

	transport, err := tlsCfg.Transport()
	if err != nil {
		return cfg, err
	}

	cfg = relocateCfg{
		imgInfo: imgInfo,
		imgWriteOptions: []remote.Option{
			remote.WithAuthFromKeychain(authn.DefaultKeychain),
			remote.WithTransport(transport),
		},
	}
	return cfg, err
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

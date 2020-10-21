// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"

	"github.com/pivotal/build-service-cli/pkg/archive"
)

type uploadImageInfo struct {
	image        v1.Image
	refTag       name.Reference
	refDigestStr string
	size         int64
}

type uploadCfg struct {
	imgInfo         uploadImageInfo
	srcTarPath      string
	imgWriteOptions []remote.Option
}

type SourceUploader struct{}

func (s SourceUploader) Upload(dstImgRefStr, srcPath string, writer io.Writer, tlsCfg TLSConfig) (string, error) {
	cfg, err := getImageUploadCfg(dstImgRefStr, srcPath, tlsCfg)
	defer os.RemoveAll(cfg.srcTarPath)
	if err != nil {
		return "", err
	}

	i := cfg.imgInfo
	writer.Write([]byte(fmt.Sprintf("\tUploading '%s'", i.refDigestStr)))

	spinner := newUploadSpinner(writer, i.size)
	defer spinner.Stop()
	go spinner.Write()

	err = remote.Write(i.refTag, i.image, cfg.imgWriteOptions...)
	if err != nil {
		return i.refDigestStr, newImageAccessError(i.refTag.String(), err)
	}

	return i.refDigestStr, err
}

func getImageUploadCfg(imgRefStr, srcPath string, tlsCfg TLSConfig) (uploadCfg, error) {
	var cfg uploadCfg

	transport, err := tlsCfg.Transport()
	if err != nil {
		return cfg, err
	}

	var srcTarPath string
	if archive.IsZip(srcPath) {
		srcTarPath, err = archive.ZipToTar(srcPath)
	} else {
		srcTarPath, err = archive.CreateTar(srcPath)
	}

	if err != nil {
		return cfg, err
	}

	info, err := getImageInfo(imgRefStr, srcTarPath)
	if err != nil {
		return cfg, err
	}

	cfg = uploadCfg{
		imgInfo:    info,
		srcTarPath: srcTarPath,
		imgWriteOptions: []remote.Option{
			remote.WithAuthFromKeychain(authn.DefaultKeychain),
			remote.WithTransport(transport),
		},
	}
	return cfg, err
}

func getImageInfo(imgRefStr, tarPath string) (uploadImageInfo, error) {
	var info uploadImageInfo

	image, err := getImageFromSrcTar(tarPath)
	if err != nil {
		return info, err
	}

	tagStr := fmt.Sprint(time.Now().UnixNano())
	refTag, err := name.ParseReference(fmt.Sprintf("%s:%s", imgRefStr, tagStr))
	if err != nil {
		return info, err
	}

	digest, err := image.Digest()
	if err != nil {
		return info, err
	}

	size, err := imageSize(image)
	if err != nil {
		return info, errors.WithStack(err)
	}

	info = uploadImageInfo{
		image:        image,
		refTag:       refTag,
		refDigestStr: fmt.Sprintf("%s@%s", imgRefStr, digest),
		size:         size,
	}
	return info, err
}

func getImageFromSrcTar(tarFilepath string) (v1.Image, error) {
	image, err := random.Image(0, 0)
	if err != nil {
		return image, err
	}

	layer, err := tarball.LayerFromFile(tarFilepath)
	if err != nil {
		return image, err
	}

	image, err = mutate.AppendLayers(image, layer)
	if err != nil {
		return image, errors.Wrap(err, "adding layer")
	}

	return image, nil
}

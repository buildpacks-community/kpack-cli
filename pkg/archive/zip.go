// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package archive

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func IsZip(source string) bool {
	file, err := os.Open(source)
	if err != nil {
		return false
	}

	defer file.Close()

	// http://golang.org/pkg/net/http/#DetectContentType
	buff := make([]byte, 512)

	_, err = file.Read(buff)
	if err != nil {
		return false
	}

	filetype := http.DetectContentType(buff)

	return filetype == "application/zip"
}

func ExtractZip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {

		fpath := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func ZipToTar(srcZip string) (string, error) {
	tarFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", fmt.Errorf("create file for tar: %s", err)
	}
	defer tarFile.Close()

	tw := tar.NewWriter(tarFile)
	defer tw.Close()

	zipReader, err := zip.OpenReader(srcZip)
	if err != nil {
		return "", err
	}
	defer zipReader.Close()

	var fileMode int64
	for _, f := range zipReader.File {
		fileMode = -1
		if isFatFile(f.FileHeader) {
			fileMode = 0777
		}

		var header *tar.Header
		if f.Mode()&os.ModeSymlink != 0 {
			target, err := func() (string, error) {
				r, err := f.Open()
				if err != nil {
					return "", nil
				}
				defer r.Close()

				// contents is the target of the symlink
				target, err := ioutil.ReadAll(r)
				if err != nil {
					return "", err
				}

				return string(target), nil
			}()

			if err != nil {
				return "", err
			}

			header, err = tar.FileInfoHeader(f.FileInfo(), target)
			if err != nil {
				return "", err
			}
		} else {
			header, err = tar.FileInfoHeader(f.FileInfo(), f.Name)
			if err != nil {
				return "", err
			}
		}

		header.Name = filepath.ToSlash(f.Name)
		finalizeHeader(header, 0, 0, fileMode)

		if err := tw.WriteHeader(header); err != nil {
			return "", err
		}

		if f.Mode().IsRegular() {
			err := func() error {
				fi, err := f.Open()
				if err != nil {
					return err
				}
				defer fi.Close()

				_, err = io.Copy(tw, fi)
				return err
			}()

			if err != nil {
				return "", err
			}
		}
	}

	return tarFile.Name(), nil
}

func isFatFile(header zip.FileHeader) bool {
	var (
		creatorFAT  uint16 = 0
		creatorVFAT uint16 = 14
	)

	// This identifies FAT files, based on the `zip` source: https://golang.org/src/archive/zip/struct.go
	firstByte := header.CreatorVersion >> 8
	return firstByte == creatorFAT || firstByte == creatorVFAT
}

// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package archive

import (
	"archive/zip"
	"fmt"
	"io"
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

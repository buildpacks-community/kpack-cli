// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package archive_test

import (
	"archive/tar"
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/kpack-cli/pkg/archive"
)

func TestZip(t *testing.T) {
	spec.Run(t, "Test Zip operations", testZip)
}

func testZip(t *testing.T, when spec.G, it spec.S) {
	when("#IsZip", func() {
		it("returns true if the file is a zip file", func() {
			file, err := ioutil.TempFile("", "file.zip")
			require.NoError(t, err)

			defer os.Remove(file.Name())

			createZip(file, t, err)

			assert.True(t, archive.IsZip(file.Name()))
		})

		it("returns false if the file is not a zip file", func() {
			file, err := ioutil.TempFile("", "file.zip")
			require.NoError(t, err)

			defer os.Remove(file.Name())

			err = ioutil.WriteFile(file.Name(), []byte("this is not a zip file"), os.ModePerm)
			require.NoError(t, err)

			assert.False(t, archive.IsZip(file.Name()))
		})
	})

	when("#ExtractZip", func() {
		it("extracts a zip file into the provided directory", func() {
			file, err := ioutil.TempFile("", "file.zip")
			require.NoError(t, err)

			tempDir, err := ioutil.TempDir("", "test")
			require.NoError(t, err)

			defer os.Remove(file.Name())
			defer os.RemoveAll(tempDir)

			createZip(file, t, err)

			err = archive.ExtractZip(file.Name(), tempDir)
			require.NoError(t, err)

			var files []string

			filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
				files = append(files, path)
				return nil
			})
			require.NoError(t, err)

			assert.Contains(t, files, filepath.Join(tempDir, "file1.txt"))
			assert.Contains(t, files, filepath.Join(tempDir, "file2.txt"))
			assert.Contains(t, files, filepath.Join(tempDir, "file3.txt"))
		})
	})

	when("#ZipToTar", func() {
		it("writes the zip file as a tar file", func() {
			file, err := ioutil.TempFile("", "file.zip")
			require.NoError(t, err)

			tempDir, err := ioutil.TempDir("", "test")
			require.NoError(t, err)

			defer os.Remove(file.Name())
			defer os.RemoveAll(tempDir)

			createZip(file, t, err)

			tarFile, err := archive.ZipToTar(file.Name())
			require.NoError(t, err)
			defer os.RemoveAll(tarFile)

			expectedFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
			count := 0
			err = checkTar(tarFile, func(header *tar.Header) {
				assert.Contains(t, expectedFiles, header.Name)
				count += 1
			})
			require.NoError(t, err)
			require.Equal(t, len(expectedFiles), count, "tar did not contain the expected number of files")
		})

		it("writes the tar to the dest dir with 0777 when files are compressed in fat (MSDOS) format", func() {
			zipFile := filepath.Join("testdata", "fat-zip-to-tar.zip")

			tarFile, err := archive.ZipToTar(zipFile)
			require.NoError(t, err)
			defer os.RemoveAll(tarFile)

			count := 0
			err = checkTar(tarFile, func(header *tar.Header) {
				assert.Equal(t, int64(0777), header.Mode)
				count += 1
			})
			require.NoError(t, err)
			require.Equal(t, 1, count, "tar did not contain the expected number of files")
		})
	})
}

func checkTar(tarPath string, checkHeader func(header *tar.Header)) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()
	tr := tar.NewReader(f)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		checkHeader(header)
	}
	return nil
}

func createZip(file *os.File, t *testing.T, err error) {
	writer := zip.NewWriter(file)
	var files = []struct {
		Name, Body string
	}{
		{"file1.txt", "Contents of file 1"},
		{"file2.txt", "Contents of file 2"},
		{"file3.txt", "Contents of file 3"},
	}
	for _, file := range files {
		f, err := writer.Create(file.Name)
		require.NoError(t, err)
		_, err = f.Write([]byte(file.Body))
		require.NoError(t, err)
	}

	err = writer.Close()
	require.NoError(t, err)
}

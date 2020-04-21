package archive_test

import (
	"archive/zip"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pivotal/build-service-cli/pkg/archive"
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

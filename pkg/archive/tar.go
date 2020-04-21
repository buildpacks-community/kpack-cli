package archive

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func CreateTar(path string) (string, error) {
	fh, err := ioutil.TempFile("", "")
	if err != nil {
		return "", fmt.Errorf("create file for tar: %s", err)
	}
	defer fh.Close()

	tw := tar.NewWriter(fh)
	defer tw.Close()

	if err := writeDirToTar(tw, path, "/", 0, 0, -1); err != nil {
		return "", err
	}

	return fh.Name(), nil
}

func writeDirToTar(tw *tar.Writer, srcDir, basePath string, uid, gid int, mode int64) error {
	return filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.Mode()&os.ModeSocket != 0 {
			return nil
		}

		var header *tar.Header
		if fi.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(file)
			if err != nil {
				return err
			}

			header, err = tar.FileInfoHeader(fi, target)
			if err != nil {
				return err
			}
		} else {
			header, err = tar.FileInfoHeader(fi, fi.Name())
			if err != nil {
				return err
			}
		}

		relPath, err := filepath.Rel(srcDir, file)
		if err != nil {
			return err
		} else if relPath == "." {
			return nil
		}

		header.Name = filepath.ToSlash(filepath.Join(basePath, relPath))
		finalizeHeader(header, uid, gid, mode)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if fi.Mode().IsRegular() {
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}

		return nil
	})
}

func finalizeHeader(header *tar.Header, uid, gid int, mode int64) {
	if mode != -1 {
		header.Mode = mode
	}
	header.Uid = uid
	header.Gid = gid
	header.Uname = ""
	header.Gname = ""
}

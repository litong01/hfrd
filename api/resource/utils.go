package resource

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func unTarGz(tarGzFilePath, destPath string) error {
	tarGzFile, err := os.Open(tarGzFilePath)
	if err != nil {
		return err
	}
	defer tarGzFile.Close()
	gzr, err := gzip.NewReader(tarGzFile)
	if err != nil {
		return err
	}
	defer gzr.Close()
	tarr := tar.NewReader(gzr)
	for {
		hdr, err := tarr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		filePath := filepath.Join(destPath, hdr.Name)
		if hdr.FileInfo().IsDir() {
			continue
		}
		file, err := createFile(filePath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(file, tarr); err != nil {
			file.Close()
			return err
		} else {
			file.Close()
		}
	}
	return nil
}

func createFile(path string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(path), 0744)
	if err != nil {
		return nil, err
	}
	return os.Create(path)
}

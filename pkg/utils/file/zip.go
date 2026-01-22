package file

import (
	"archive/zip"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

// isZipFile checks if the file is a zip file.
func isZipFile(path string) bool {
	return strings.HasSuffix(path, ".zip")
}

// ReadZipFile reads a zip file and returns a zip.ReadCloser and a close function.
func ReadZipFile(path string) (*zip.ReadCloser, func(), error) {
	if !isZipFile(path) {
		return nil, nil, errors.New("not a zip file")
	}

	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return nil, nil, err
	}

	closeFunc := func() {
		_ = zipReader.Close()
	}

	return zipReader, closeFunc, nil
}

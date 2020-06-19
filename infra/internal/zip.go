package internal

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func zipFiles(fPaths []string, fPathFunc func(string) (string, error)) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for _, fPath := range fPaths {
		if err := func() error {
			r, err := os.Open(fPath)
			if err != nil {
				return err
			}
			defer r.Close()

			zPath, err := fPathFunc(fPath)
			if err != nil {
				return err
			}
			w, err := zipWriter.Create(zPath)
			if err != nil {
				return err
			}

			_, err = io.Copy(w, r)
			if err != nil {
				return err
			}

			return nil
		}(); err != nil {
			return nil, err
		}
	}

	err := zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func unzipFiles(src string, dest string) error {
	var filenames []string

	zipReader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, zipFile := range zipReader.File {
		if err := func() error {
			if zipFile.FileInfo().IsDir() {
				return nil
			}

			fPath := filepath.Join(dest, zipFile.Name)

			// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
			if !strings.HasPrefix(fPath, fmt.Sprintf("%s%s", filepath.Clean(dest), string(os.PathSeparator))) {
				return fmt.Errorf("%s: illegal file path", fPath)
			}

			filenames = append(filenames, fPath)

			// Make File
			err := os.MkdirAll(filepath.Dir(fPath), os.ModePerm)
			if err != nil {
				return err
			}

			f, err := os.Create(fPath)
			if err != nil {
				return err
			}
			defer f.Close()

			zipFileReader, err := zipFile.Open()
			if err != nil {
				return err
			}
			defer zipFileReader.Close()

			_, err = io.Copy(f, zipFileReader)
			if err != nil {
				return err
			}

			return nil
		}(); err != nil {
			return err
		}

	}
	return nil
}

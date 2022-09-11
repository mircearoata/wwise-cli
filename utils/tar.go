package utils

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/ulikunitz/xz"
)

func ExtractTarXz(xzStream io.Reader, extractPath string) error {
	uncompressedStream, err := xz.NewReader(xzStream)
	if err != nil {
		return errors.Wrap(err, "failed to create gzip reader")
	}
	if err != nil {
		return errors.Wrap(err, "failed to create tar reader")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		switch {

		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		target := filepath.Join(extractPath, header.Name)

		parentDir := filepath.Dir(target)
		if _, err := os.Stat(parentDir); err != nil {
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				return err
			}
		}

		switch header.Typeflag {

		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tarReader); err != nil {
				return err
			}

			f.Close()
		}
	}
}

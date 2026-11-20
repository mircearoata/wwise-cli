package utils

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ulikunitz/xz/v2"
)

func ExtractTarXz(xzStream io.Reader, extractPath string) error {
	uncompressedStream, err := xz.NewReader(xzStream)
	if err != nil {
		return fmt.Errorf("failed to create xz reader: %w", err)
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
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
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

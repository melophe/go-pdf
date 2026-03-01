package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// generateZIP creates a ZIP archive containing the given image files.
func generateZIP(imagePaths []string, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	for _, imgPath := range imagePaths {
		if err := addFileToZip(zw, imgPath); err != nil {
			return err
		}
	}

	return nil
}

// addFileToZip adds a single file to the ZIP archive.
func addFileToZip(zw *zip.Writer, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(filePath)
	header.Method = zip.Deflate

	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, f)
	return err
}

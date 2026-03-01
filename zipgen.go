package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// generateZIP creates a ZIP archive containing the given image files.
func generateZIP(imagePaths []string, outputPath string) error {
	return generateZIPWithProgress(imagePaths, outputPath, nil)
}

// generateZIPWithProgress creates a ZIP archive with a progress callback.
func generateZIPWithProgress(imagePaths []string, outputPath string, onProgress func(current int)) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	for i, imgPath := range imagePaths {
		if err := addFileToZip(zw, imgPath); err != nil {
			return err
		}
		if onProgress != nil {
			onProgress(i + 1)
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

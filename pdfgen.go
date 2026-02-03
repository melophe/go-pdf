package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-pdf/fpdf"
)

const (
	// A4 portrait dimensions in millimeters.
	pageWidth  = 210.0
	pageHeight = 297.0
	// Margin around the image in millimeters.
	margin = 10.0
)

// generatePDF creates a PDF file from the given image paths and writes it to outputPath.
func generatePDF(imagePaths []string, outputPath string) error {
	pdf := fpdf.New("P", "mm", "A4", "")

	for _, imgPath := range imagePaths {
		if err := addImagePage(pdf, imgPath); err != nil {
			return fmt.Errorf("failed to add %s: %w", filepath.Base(imgPath), err)
		}
	}

	return pdf.OutputFileAndClose(outputPath)
}

// addImagePage adds a new page with the given image, auto-resized to fit within A4 margins.
func addImagePage(pdf *fpdf.Fpdf, imgPath string) error {
	// Get image dimensions in pixels.
	w, h, err := getImageSize(imgPath)
	if err != nil {
		return err
	}

	// Calculate the available area inside margins.
	availW := pageWidth - margin*2
	availH := pageHeight - margin*2

	// Scale image to fit within available area while keeping aspect ratio.
	scaleW := availW / float64(w)
	scaleH := availH / float64(h)
	scale := scaleW
	if scaleH < scaleW {
		scale = scaleH
	}
	imgW := float64(w) * scale
	imgH := float64(h) * scale

	// Center the image on the page.
	x := (pageWidth - imgW) / 2
	y := (pageHeight - imgH) / 2

	// Determine image type from extension.
	imgType := detectImageType(imgPath)

	pdf.AddPage()
	pdf.Image(imgPath, x, y, imgW, imgH, false, imgType, 0, "")

	return nil
}

// getImageSize returns the width and height in pixels of an image file.
func getImageSize(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

// detectImageType returns the fpdf image type string based on file extension.
func detectImageType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "JPG"
	case ".png":
		return "PNG"
	default:
		return ""
	}
}

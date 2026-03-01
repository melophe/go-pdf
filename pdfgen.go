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

// PageSizeMode controls how the page size is determined.
type PageSizeMode int

const (
	// PageSizeA4 uses fixed A4 portrait (210x297 mm) for every page.
	PageSizeA4 PageSizeMode = iota
	// PageSizeFitImage sets each page size to match the image dimensions.
	PageSizeFitImage
)

const (
	// A4 portrait dimensions in millimeters.
	a4Width  = 210.0
	a4Height = 297.0
	// Margin around the image in millimeters (used only in A4 mode).
	margin = 10.0
	// Pixels per inch assumed when converting pixel size to mm.
	dpi = 96.0
)

// generatePDF creates a PDF from image paths using the given page size mode.
func generatePDF(imagePaths []string, outputPath string, mode PageSizeMode) error {
	return generatePDFWithProgress(imagePaths, outputPath, mode, nil)
}

// generatePDFWithProgress creates a PDF with a progress callback.
func generatePDFWithProgress(imagePaths []string, outputPath string, mode PageSizeMode, onProgress func(current int)) error {
	pdf := fpdf.New("P", "mm", "A4", "")

	for i, imgPath := range imagePaths {
		if err := addImagePage(pdf, imgPath, mode); err != nil {
			return fmt.Errorf("failed to add %s: %w", filepath.Base(imgPath), err)
		}
		if onProgress != nil {
			onProgress(i + 1)
		}
	}

	return pdf.OutputFileAndClose(outputPath)
}

// addImagePage adds a single page with the image, using the chosen page size mode.
func addImagePage(pdf *fpdf.Fpdf, imgPath string, mode PageSizeMode) error {
	w, h, err := getImageSize(imgPath)
	if err != nil {
		return err
	}

	imgType := detectImageType(imgPath)

	switch mode {
	case PageSizeFitImage:
		return addFitToImagePage(pdf, imgPath, imgType, w, h)
	default:
		return addA4Page(pdf, imgPath, imgType, w, h)
	}
}

// addA4Page places the image centered on an A4 page, scaled to fit within margins.
func addA4Page(pdf *fpdf.Fpdf, imgPath, imgType string, w, h int) error {
	availW := a4Width - margin*2
	availH := a4Height - margin*2

	scale := fitScale(float64(w), float64(h), availW, availH)
	imgW := float64(w) * scale
	imgH := float64(h) * scale

	x := (a4Width - imgW) / 2
	y := (a4Height - imgH) / 2

	pdf.AddPage()
	pdf.Image(imgPath, x, y, imgW, imgH, false, imgType, 0, "")
	return nil
}

// addFitToImagePage sets the page size to the image dimensions (no margin, no scaling).
func addFitToImagePage(pdf *fpdf.Fpdf, imgPath, imgType string, w, h int) error {
	// Convert pixel dimensions to mm at the assumed DPI.
	pageW := pixelToMm(float64(w))
	pageH := pixelToMm(float64(h))

	pdf.AddPageFormat("", fpdf.SizeType{Wd: pageW, Ht: pageH})
	pdf.Image(imgPath, 0, 0, pageW, pageH, false, imgType, 0, "")
	return nil
}

// fitScale returns the scale factor to fit (srcW x srcH) into (maxW x maxH) while keeping aspect ratio.
func fitScale(srcW, srcH, maxW, maxH float64) float64 {
	scaleW := maxW / srcW
	scaleH := maxH / srcH
	if scaleH < scaleW {
		return scaleH
	}
	return scaleW
}

// pixelToMm converts pixels to millimeters at the package-level DPI.
func pixelToMm(px float64) float64 {
	return px / dpi * 25.4
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

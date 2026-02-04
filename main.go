package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// scanImages returns all JPG/PNG file paths in the given directory (non-recursive).
func scanImages(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}
	return paths, nil
}

func main() {
	a := app.New()
	w := a.NewWindow("go-pdf")

	countLabel := widget.NewLabel("No images selected")

	il := newImageList(func() {
		if il := countLabel; il != nil {
			countLabel.SetText(fmt.Sprintf("%d image(s) selected", 0))
		}
	})

	// Override onUpdate now that countLabel exists.
	il.onUpdate = func() {
		if il.count() == 0 {
			countLabel.SetText("No images selected")
		} else {
			countLabel.SetText(fmt.Sprintf("%d image(s) selected", il.count()))
		}
	}

	addBtn := widget.NewButton("Add Image", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()
			il.add(reader.URI().Path())
		}, w)

		fd.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".jpeg", ".png"}))
		fd.SetTitleText("Select Image")
		fd.Show()
	})

	statusLabel := widget.NewLabel("")

	addFolderBtn := widget.NewButton("Add Folder", func() {
		fd := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			imgs, scanErr := scanImages(uri.Path())
			if scanErr != nil {
				statusLabel.SetText("Error: " + scanErr.Error())
				return
			}
			for _, p := range imgs {
				il.add(p)
			}
		}, w)
		fd.SetTitleText("Select Folder")
		fd.Show()
	})

	convertBtn := widget.NewButton("Convert to PDF", func() {
		if il.count() == 0 {
			statusLabel.SetText("No images to convert.")
			return
		}

		sd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil || writer == nil {
				return
			}
			outputPath := writer.URI().Path()
			writer.Close()

			statusLabel.SetText("Converting...")
			if genErr := generatePDF(il.paths, outputPath); genErr != nil {
				statusLabel.SetText("Error: " + genErr.Error())
			} else {
				statusLabel.SetText("Done: " + outputPath)
			}
		}, w)

		sd.SetFileName("output.pdf")
		sd.SetTitleText("Save PDF")
		sd.Show()
	})

	w.SetContent(container.NewBorder(
		// top
		container.NewVBox(
			widget.NewLabel("Image to PDF Converter"),
			container.NewHBox(addBtn, addFolderBtn, convertBtn),
			countLabel,
		),
		// bottom
		statusLabel,
		// left
		nil,
		// right
		nil,
		// center (fills remaining space)
		il.widget,
	))

	w.Resize(fyne.NewSize(500, 400))
	w.ShowAndRun()
}

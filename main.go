package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

const prefKeyDefaultFolder = "defaultFolder"

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
	// Sort by modification time, newest first.
	sort.Slice(paths, func(i, j int) bool {
		fi, errI := os.Stat(paths[i])
		fj, errJ := os.Stat(paths[j])
		if errI != nil || errJ != nil {
			return false
		}
		return fi.ModTime().After(fj.ModTime())
	})

	return paths, nil
}

func main() {
	a := app.NewWithID("com.losts.go-pdf")
	w := a.NewWindow("go-pdf")

	countLabel := widget.NewLabel("No images selected")
	statusLabel := widget.NewLabel("")

	var il *imageList
	il = newImageList(func() {
		if il.count() == 0 {
			countLabel.SetText("No images selected")
		} else {
			countLabel.SetText(fmt.Sprintf("%d image(s) selected", il.count()))
		}
	})

	// Load images from default folder on startup.
	defaultFolder := a.Preferences().String(prefKeyDefaultFolder)
	if defaultFolder != "" {
		imgs, err := scanImages(defaultFolder)
		if err == nil {
			for _, p := range imgs {
				il.add(p)
			}
			statusLabel.SetText("Loaded from: " + defaultFolder)
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

	setDefaultBtn := widget.NewButton("Set Default", func() {
		fd := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			folder := uri.Path()
			a.Preferences().SetString(prefKeyDefaultFolder, folder)
			statusLabel.SetText("Default folder set: " + folder)
		}, w)
		fd.SetTitleText("Set Default Folder")
		fd.Show()
	})

	// Page size selector: index matches PageSizeMode order.
	pageSizeOptions := []string{"A4", "Fit to image"}
	pageSizeSelect := widget.NewSelect(pageSizeOptions, nil)
	pageSizeSelect.Selected = "A4"

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

			mode := PageSizeA4
			if pageSizeSelect.Selected == "Fit to image" {
				mode = PageSizeFitImage
			}

			statusLabel.SetText("Converting...")
			if genErr := generatePDF(il.paths, outputPath, mode); genErr != nil {
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
			container.NewHBox(addBtn, addFolderBtn, setDefaultBtn, convertBtn),
			container.NewHBox(widget.NewLabel("Page size:"), pageSizeSelect),
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

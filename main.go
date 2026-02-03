package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// imagePaths holds the list of selected image file paths.
var imagePaths []string

func main() {
	a := app.New()
	w := a.NewWindow("go-pdf")

	countLabel := widget.NewLabel("No images selected")

	addBtn := widget.NewButton("Add Image", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()
			imagePaths = append(imagePaths, reader.URI().Path())
			countLabel.SetText(fmt.Sprintf("%d image(s) selected", len(imagePaths)))
		}, w)

		fd.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".jpeg", ".png"}))
		fd.SetTitleText("Select Image")
		fd.Show()
	})

	w.SetContent(container.NewVBox(
		widget.NewLabel("Image to PDF Converter"),
		addBtn,
		countLabel,
	))

	w.Resize(fyne.NewSize(500, 400))
	w.ShowAndRun()
}

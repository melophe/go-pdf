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

	w.SetContent(container.NewBorder(
		// top
		container.NewVBox(
			widget.NewLabel("Image to PDF Converter"),
			container.NewHBox(addBtn),
			countLabel,
		),
		// bottom
		nil,
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

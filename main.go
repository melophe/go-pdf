package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("go-pdf")

	w.SetContent(container.NewCenter(
		widget.NewLabel("Image to PDF Converter"),
	))

	w.Resize(fyne.NewSize(500, 400))
	w.ShowAndRun()
}

package main

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// imageList manages the list of image paths and provides a Fyne List widget.
type imageList struct {
	paths  []string
	widget *widget.List
	// onUpdate is called whenever the list changes.
	onUpdate func()
}

func newImageList(onUpdate func()) *imageList {
	il := &imageList{onUpdate: onUpdate}

	il.widget = widget.NewList(
		// length returns the number of items.
		func() int { return len(il.paths) },
		// createItem returns a new row layout.
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("placeholder"),
				widget.NewButton("Remove", nil),
			)
		},
		// updateItem sets the label and button action for each row.
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			row := obj.(*fyne.Container)
			label := row.Objects[0].(*widget.Label)
			btn := row.Objects[1].(*widget.Button)

			label.SetText(filepath.Base(il.paths[id]))
			btn.OnTapped = func() {
				il.remove(int(id))
			}
		},
	)

	return il
}

func (il *imageList) add(path string) {
	il.paths = append(il.paths, path)
	il.widget.Refresh()
	il.onUpdate()
}

func (il *imageList) remove(index int) {
	il.paths = append(il.paths[:index], il.paths[index+1:]...)
	il.widget.Refresh()
	il.onUpdate()
}

func (il *imageList) count() int {
	return len(il.paths)
}

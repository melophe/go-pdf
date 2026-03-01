package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

var numRegex = regexp.MustCompile(`\d+`)

// naturalLess compares two strings in natural order (1, 2, 10, 20 instead of 1, 10, 2, 20).
func naturalLess(a, b string) bool {
	aNums := numRegex.FindAllString(a, -1)
	bNums := numRegex.FindAllString(b, -1)

	// If both have numbers, compare the first number.
	if len(aNums) > 0 && len(bNums) > 0 {
		aNum, _ := strconv.Atoi(aNums[0])
		bNum, _ := strconv.Atoi(bNums[0])
		if aNum != bNum {
			return aNum < bNum
		}
	}
	return a < b
}

const (
	prefKeyDefaultFolder = "defaultFolder"
	prefKeyOutputFolder  = "outputFolder"
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
	// Sort by filename in natural order (1, 2, 10 instead of 1, 10, 2).
	sort.Slice(paths, func(i, j int) bool {
		return naturalLess(filepath.Base(paths[i]), filepath.Base(paths[j]))
	})

	return paths, nil
}

func main() {
	a := app.NewWithID("com.losts.go-pdf")
	w := a.NewWindow("go-pdf")

	countLabel := widget.NewLabel("No images selected")
	statusLabel := widget.NewLabel("")
	progressBar := widget.NewProgressBar()

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
	pageSizeSelect.Selected = "Fit to image"

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

			// Remember output folder for next time.
			a.Preferences().SetString(prefKeyOutputFolder, filepath.Dir(outputPath))

			mode := PageSizeA4
			if pageSizeSelect.Selected == "Fit to image" {
				mode = PageSizeFitImage
			}

			// Run conversion in background goroutine.
			paths := make([]string, len(il.paths))
			copy(paths, il.paths)
			total := len(paths)
			zipPath := strings.TrimSuffix(outputPath, ".pdf") + ".zip"

			go func() {
				progressBar.SetValue(0)
				statusLabel.SetText("Converting...")

				// Track progress from both PDF and ZIP.
				var pdfProgress, zipProgress int
				updateProgress := func() {
					combined := float64(pdfProgress+zipProgress) / float64(total*2)
					progressBar.SetValue(combined)
					statusLabel.SetText(fmt.Sprintf("PDF: %d/%d, ZIP: %d/%d", pdfProgress, total, zipProgress, total))
				}

				// Run PDF and ZIP generation in parallel.
				var pdfErr, zipErr error
				done := make(chan struct{}, 2)

				go func() {
					pdfErr = generatePDFWithProgress(paths, outputPath, mode, func(current int) {
						pdfProgress = current
						updateProgress()
					})
					done <- struct{}{}
				}()

				go func() {
					zipErr = generateZIPWithProgress(paths, zipPath, func(current int) {
						zipProgress = current
						updateProgress()
					})
					done <- struct{}{}
				}()

				// Wait for both to complete.
				<-done
				<-done

				if pdfErr != nil {
					statusLabel.SetText("PDF error: " + pdfErr.Error())
					return
				}
				if zipErr != nil {
					statusLabel.SetText("ZIP error: " + zipErr.Error())
					return
				}

				progressBar.SetValue(1.0)
				statusLabel.SetText("Done: " + outputPath + " & .zip")
			}()
		}, w)

		// Set default filename from first image.
		firstImage := filepath.Base(il.paths[0])
		ext := filepath.Ext(firstImage)
		defaultName := strings.TrimSuffix(firstImage, ext) + ".pdf"
		sd.SetFileName(defaultName)

		// Set starting location to last used output folder.
		outputFolder := a.Preferences().String(prefKeyOutputFolder)
		if outputFolder != "" {
			uri, err := storage.ListerForURI(storage.NewFileURI(outputFolder))
			if err == nil {
				sd.SetLocation(uri)
			}
		}

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
			progressBar,
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

	// Handle drag and drop of image files.
	w.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		var dropped []string
		for _, uri := range uris {
			path := uri.Path()
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				dropped = append(dropped, path)
			}
		}
		// Sort dropped files in natural order.
		sort.Slice(dropped, func(i, j int) bool {
			return naturalLess(filepath.Base(dropped[i]), filepath.Base(dropped[j]))
		})
		for _, p := range dropped {
			il.add(p)
		}
		if len(dropped) > 0 {
			statusLabel.SetText(fmt.Sprintf("Dropped %d image(s)", len(dropped)))
		}
	})

	w.Resize(fyne.NewSize(500, 400))
	w.ShowAndRun()
}

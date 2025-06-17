package frontenddesktop

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"stacklatex/latex"
)

type weightedHBox struct {
    weights []float32
}

func (w *weightedHBox) Layout(objects []fyne.CanvasObject, size fyne.Size) {
    totalWeight := float32(0)
    for _, weight := range w.weights {
        totalWeight += weight
    }
    x := float32(0)
    for i, child := range objects {
        width := size.Width * (w.weights[i] / totalWeight)
        child.Resize(fyne.NewSize(width, size.Height))
        child.Move(fyne.NewPos(x, 0))
        x += width
    }
}

func (w *weightedHBox) MinSize(objects []fyne.CanvasObject) fyne.Size {
    minHeight := float32(0)
    minWidth := float32(0)
    for _, obj := range objects {
        min := obj.MinSize()
        minWidth += min.Width
        if min.Height > minHeight {
            minHeight = min.Height
        }
    }
    return fyne.NewSize(minWidth, minHeight)
}

type weightedVBox struct {
    weights []float32
}

func (w *weightedVBox) Layout(objects []fyne.CanvasObject, size fyne.Size) {
    totalWeight := float32(0)
    for _, weight := range w.weights {
        totalWeight += weight
    }
    x := float32(0)
    for i, child := range objects {
        height := size.Height * (w.weights[i] / totalWeight)
        child.Resize(fyne.NewSize(size.Width, height))
        child.Move(fyne.NewPos(0, x))
        x += height
    }
}

func (w *weightedVBox) MinSize(objects []fyne.CanvasObject) fyne.Size {
    minHeight := float32(0)
    minWidth := float32(0)
    for _, obj := range objects {
        min := obj.MinSize()
        minHeight += min.Height
        if min.Width > minWidth {
            minWidth = min.Width
        }
    }
    return fyne.NewSize(minWidth, minHeight)
}

func RunDesktopApp() {
	app := app.New()
	window := app.NewWindow("Hello")

	empty := container.NewWithoutLayout()

	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Enter LaTeX code here!")
	output := widget.NewMultiLineEntry()
	output.SetPlaceHolder("Output will appear here!")
	log := widget.NewLabel("")
    info := widget.NewLabel("")
	submit := widget.NewButton("Transform to \nSTACK-compatible LaTeX", func() {
		transformed := latex.TransformLatex(input.Text)
		if transformed.Success {
			output.SetText(transformed.Transformed)
			log.SetText(transformed.Log)
            info.SetText(transformed.Info)
		} else {
            output.SetText("")
			log.SetText(transformed.ErrorMessage)
            info.SetText("")
		}
	})
	copyButton := widget.NewButton("Copy output to clipboard", func() {
		clip := app.Clipboard()
		clip.SetContent(output.Text)
	})


	mid_content := container.New(&weightedVBox{weights: []float32{0.5, 0.1, 0.03, 0.1, 0.5}}, empty, submit, empty, copyButton, empty)
	input_col := container.New(&weightedVBox{weights: []float32{0.05,0.4,0.4}}, empty, input, log)
	output_col := container.New(&weightedVBox{weights: []float32{0.05,0.4,0.4}}, empty, output, info)
	content := container.New(&weightedHBox{weights: []float32{0.05, 6, 0.5, 3, 0.5, 6, 0.05}}, empty, input_col, empty, mid_content, empty, output_col, empty)

	window.SetContent(content)
	window.Resize(fyne.Size{Width: 1080, Height: 720})
	window.ShowAndRun()
}
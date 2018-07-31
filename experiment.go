package main

import (
	"image"
	"image/color"
	"math/rand"
	"time"

	"github.com/llgcode/draw2d/draw2dimg"
)

const (
	gridWidth   = 8
	gridHeight  = 8
	gridSpacing = 25
	lineLength  = 20
	randomLines = 5
)

func NewExperiment() *Experiment {
	e := &Experiment{
		preDelay:    500 * time.Millisecond,
		postDelay:   100 * time.Millisecond,
		displayTime: 90 * time.Millisecond,
		expect:      rand.Intn(2) + 1,
		question:    "Was there a red vertical line?",
	}

	grid := [gridWidth][gridHeight]int{}

	for i := 0; i < randomLines; i++ {
		x := rand.Intn(gridWidth)
		y := rand.Intn(gridHeight)
		grid[x][y] = 1
		x = rand.Intn(gridWidth)
		y = rand.Intn(gridHeight)
		grid[x][y] = 2
	}
	if e.expect == 1 {
		x := rand.Intn(gridWidth)
		y := rand.Intn(gridHeight)
		grid[x][y] = 3
	}

	// Draw the experiment
	image := image.NewRGBA(image.Rect(0, 0, gridWidth*gridSpacing, gridHeight*gridSpacing))
	e.image = image
	gc := draw2dimg.NewGraphicContext(image)

	gc.SetLineWidth(2)
	for x := 0; x < gridWidth; x++ {
		for y := 0; y < gridHeight; y++ {
			var xoff, yoff, xdiff, ydiff int
			gc.BeginPath()
			switch grid[x][y] {
			case 0:
				continue
			case 1:
				gc.SetStrokeColor(color.RGBA{0xff, 0x00, 0x00, 0xff})
				xoff = -lineLength / 2
				xdiff = lineLength / 2
			case 2:
				gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0xff, 0xff})
				yoff = -lineLength / 2
				ydiff = lineLength / 2
			case 3:
				gc.SetStrokeColor(color.RGBA{0xff, 0x00, 0x00, 0xff})
				yoff = -lineLength / 2
				ydiff = lineLength / 2
			}
			translate := func(grid, offset int) float64 {
				return float64(grid*gridSpacing+offset) + gridSpacing/2
			}
			gc.MoveTo(translate(x, xoff), translate(y, yoff))
			gc.LineTo(translate(x, xdiff), translate(y, ydiff))
			gc.Stroke()
		}
	}
	gc.Close()

	return e
}

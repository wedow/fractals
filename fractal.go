package main

import (
	"image"
	"image/color"
	"math"
	"math/cmplx"
	"github.com/skelterjohn/go.wde"
	_ "github.com/skelterjohn/go.wde/init"
	"fmt"
	"runtime"
)

// Utility function to convert a point on a Canvas to a
// complex number given a zoom level and the camplex value
// to be shown in the center of the Canvas
// A zoom of 1 means one pixel correspond to one unit in
// the complex plane
func toCmplx(x, y int, zoom float64, center complex128) complex128 {
	return center + complex(float64(x)/zoom, float64(y)/zoom)
}

// Perform iter iterations using the mandelbrot algorithm, and return
// the magnitude of the result
func mandelbrot(c complex128, iter int) float64 {
	z := complex(0, 0)
	for i := 0; i < iter; i++ {
		z = z*z + c
		if cmplx.Abs(z) > 1000 {
			return 1000
		}
	}
	return cmplx.Abs(z)
}

// Creates a function for converting a magnitude into a color
// based on a gradient image file
func createColorizer(filename string) func(float64) color.Color {
	gradient := CanvasFromFile(filename)
	limit := gradient.Bounds().Size().Y - 1
	return func(mag float64) color.Color {
		// Clamp magnitude to size of gradient
		m := int(math.Max(math.Min(300*mag, float64(limit)), 1))
		return gradient.At(0, m)
	}
}

func drawFractal(canvas *Canvas, zoom float64, center complex128, colorizer func(float64) color.Color) {
	size := canvas.Bounds().Size()
	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			c := toCmplx(x-size.X/2, y-size.Y/2, zoom, center)
			mag := mandelbrot(c, 50)
			color := colorizer(mag)
			canvas.Set(x, y, color)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	width, height := 400, 300
	canvas := NewCanvas(image.Rect(0, 0, width, height))
	zoom := 16000.0
	center := complex(-0.71, -0.25)
	colorizer := createColorizer("fractalGradients/gradient1.png")

	dw, err := wde.NewWindow(width, height)
	if err != nil {
		fmt.Println(err)
		return
	}
	dw.SetTitle("Fractals")
	dw.Show()

	drawFractal(canvas, zoom, center, colorizer)
	dw.Screen().CopyRGBA(&canvas.RGBA, canvas.Bounds())
	dw.FlushImage()

	events := dw.EventChan()
	done := make(chan bool)
	input := make(chan string, 100)
	go keyHandler(events, done, input)

	go func() {
		var i string
		for {
			select {
			case i = <-input:
				switch i {
				case "zoomIn":
					zoom *= 1.05
				case "zoomOut":
					zoom *= 0.95
				case "panUp":
					center -= complex(0, 10) * complex(1/zoom, 0)
				case "panDown":
					center += complex(0, 10) * complex(1/zoom, 0)
				case "panLeft":
					center -= complex(10, 0) * complex(1/zoom, 0)
				case "panRight":
					center += complex(10, 0) * complex(1/zoom, 0)
				}

				drawFractal(canvas, zoom, center, colorizer)
				dw.Screen().CopyRGBA(&canvas.RGBA, canvas.Bounds())
				dw.FlushImage()
			}
		}
	}()

	<-done
}


func keyHandler(events <-chan interface{}, done chan bool, input chan string) {
loop:
	for ei := range events {
		runtime.Gosched()
		switch e := ei.(type) {
		case wde.KeyUpEvent:
			for i:= 0; i < len(input); i++ {
				_ = <- input
			}
		case wde.KeyTypedEvent:
			switch e.Key {
			case "prior":
				input <- "zoomIn"
			case "next":
				input <- "zoomOut"
			case "up_arrow":
				input <- "panUp"
			case "down_arrow":
				input <- "panDown"
			case "left_arrow":
				input <- "panLeft"
			case "right_arrow":
				input <- "panRight"
			}
		case wde.CloseEvent:
			break loop
		}
	}
	done <- true
}
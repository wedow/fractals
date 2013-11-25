package main

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"os"
)

type Canvas struct {
	image.RGBA
}

func NewCanvas(r image.Rectangle) *Canvas {
	canvas := new(Canvas)
	canvas.RGBA = *image.NewRGBA(r)
	return canvas
}

func (c Canvas) Clone() *Canvas {
	clone := NewCanvas(c.Bounds())
	copy(clone.Pix, c.Pix)
	return clone
}

func CanvasFromFile(filename string) *Canvas {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	m, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	canvas := NewCanvas(m.Bounds())
	draw.Draw(canvas, m.Bounds(), m, image.ZP, draw.Src)
  return canvas
}

func (c Canvas) DrawGradient() {
	size := c.Bounds().Size()
	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			color := color.RGBA{
				uint8(255 * x / size.X),
				uint8(255 * y / size.Y),
				55,
				255}
			c.Set(x, y, color)
		}
	}
}

func (c Canvas) DrawLine(color color.RGBA, from Vector, to Vector) {
	delta := to.Sub(from)
	length := delta.Length()
	x_step, y_step := delta.X/length, delta.Y/length
	limit := int(length + 0.5)
	for i := 0; i < limit; i++ {
		x := from.X + float64(i)*x_step
		y := from.Y + float64(i)*y_step
		c.Set(int(x), int(y), color)
	}
}

func (c Canvas) DrawCircle(color color.RGBA, at Vector, radius int) {
	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			if x*x+y*y <= radius*radius {
				c.Set(int(at.X)+x, int(at.Y)+y, color)
			}
		}
	}
}

func (c Canvas) DrawRect(color color.RGBA, min Vector, max Vector) {
	for x := int(min.X); x <= int(max.X); x++ {
		for y := int(min.Y); y <= int(max.Y); y++ {
			c.Set(x, y, color)
		}
	}
}

func (c Canvas) DrawSpiral(color color.RGBA, from Vector) {
	dir := Vector{0, 2}
	last := from
	for i := 0; i < 10000; i++ {
		next := last.Add(dir)
		c.DrawLine(color, last, next)
		dir.Rotate(0.03)
		dir.Scale(0.999)
		last = next
	}
}

func (c Canvas) Blur(radius int, weight WeightFunction) {
	clone := c.Clone()
	size := c.Bounds().Size()
	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			color := c.BlurPixel(x, y, radius, weight)
			clone.Set(x, y, color)
		}
	}
	copy(c.Pix, clone.Pix)
}

func (c Canvas) BlurPixel(x int, y int, radius int, weight WeightFunction) color.Color {
	weightSum := float64(0)
	size := c.Bounds().Size()
	outR, outG, outB := float64(0), float64(0), float64(0)
	for i := x - radius; i < x+radius+1; i++ {
		if i < 0 || i > size.X {
			continue
		}
		for j := y - radius; j < y+radius+1; j++ {
			if j < 0 || j > size.Y {
				continue
			}
			weight := weight.Weight(i-x, j-y)
			r, g, b, _ := c.At(i, j).RGBA()
			outR += float64(r) * weight
			outG += float64(g) * weight
			outB += float64(b) * weight
			weightSum += weight
		}
	}
	// Need to divide by 0xFF as the RGBA() function returns color values as uint32
	// and we need uint8
	return color.RGBA{
		uint8(outR / (weightSum * 0xFF)),
		uint8(outG / (weightSum * 0xFF)),
		uint8(outB / (weightSum * 0xFF)),
		255}
}

// Blur weighting functions
type WeightFunction interface {
	Weight(x int, y int) float64
}

type WeightFunctionBox struct{}

func (w WeightFunctionBox) Weight(x int, y int) float64 { return 1.0 }

type WeightFunctionDist struct{}

func (w WeightFunctionDist) Weight(x int, y int) float64 {
	d := math.Hypot(float64(x), float64(y))
	return 1 / (1 + d)
}

type WeightFunctionMotion struct {
}

func (w WeightFunctionMotion) Weight(x int, y int) float64 {
	if y != 0 {
		return 0
	}
	if x < 0 {
		return 0
	}
	return 0.3 + 0.7/math.Sqrt(1+float64(x))
}

type WeightFunctionDouble struct {
	split int
}

func (w WeightFunctionDouble) Weight(x int, y int) float64 {
	if y == 0 && (x == w.split || x == -w.split) {
		return 1.0
	} else {
		return 0
	}
}

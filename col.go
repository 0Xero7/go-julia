package main

import (
	"bytes"
	"image"
	"image/color"
	"log"
	"math"
	"os"
)

type ColorOf interface {
	Get(l float64) color.RGBA
}

type SpectralColor struct{}

func (c SpectralColor) Get(arg float64) color.RGBA {
	l := 400 + (300)*arg

	var t float64
	var r float64 = 0
	var g float64 = 0
	var b float64 = 0

	if (l >= 400.0) && (l < 410.0) {
		t = (l - 400.0) / (410.0 - 400.0)
		r = +(0.33 * t) - (0.20 * t * t)
	} else if (l >= 410.0) && (l < 475.0) {
		t = (l - 410.0) / (475.0 - 410.0)
		r = 0.14 - (0.13 * t * t)
	} else if (l >= 545.0) && (l < 595.0) {
		t = (l - 545.0) / (595.0 - 545.0)
		r = +(1.98 * t) - (t * t)
	} else if (l >= 595.0) && (l < 650.0) {
		t = (l - 595.0) / (650.0 - 595.0)
		r = 0.98 + (0.06 * t) - (0.40 * t * t)
	} else if (l >= 650.0) && (l < 700.0) {
		t = (l - 650.0) / (700.0 - 650.0)
		r = 0.65 - (0.84 * t) + (0.20 * t * t)
	}

	if (l >= 415.0) && (l < 475.0) {
		t = (l - 415.0) / (475.0 - 415.0)
		g = +(0.80 * t * t)
	} else if (l >= 475.0) && (l < 590.0) {
		t = (l - 475.0) / (590.0 - 475.0)
		g = 0.8 + (0.76 * t) - (0.80 * t * t)
	} else if (l >= 585.0) && (l < 639.0) {
		t = (l - 585.0) / (639.0 - 585.0)
		g = 0.84 - (0.84 * t)
	}

	if (l >= 400.0) && (l < 475.0) {
		t = (l - 400.0) / (475.0 - 400.0)
		b = +(2.20 * t) - (1.50 * t * t)
	} else if (l >= 475.0) && (l < 560.0) {
		t = (l - 475.0) / (560.0 - 475.0)
		b = 0.7 - (t) + (0.30 * t * t)
	}

	return color.RGBA{uint8(255 * r), uint8(255 * g), uint8(255 * b), 255}
}

type Histogram struct {
	file  image.Image
	width int
}

func NewHistogram(path string) *Histogram {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}

	return &Histogram{
		file:  img,
		width: img.Bounds().Max.X,
	}
}

func (h Histogram) Get(l float64) color.RGBA {
	r, g, b, a := h.file.At(int(l*float64(h.width)), 0).RGBA()
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
}

type ColorRangeConverer interface {
	Get(l, r float64) float64
}

type ExponentialMappedModuloColorRangeConverer struct {
	S     float64
	Steps int
}

func (c ExponentialMappedModuloColorRangeConverer) Get(l, h float64) float64 {
	return math.Mod(math.Pow(math.Pow(float64(l)/float64(h), float64(c.S))*float64(c.Steps), 1.5), float64(c.Steps)) / float64(c.Steps)
}

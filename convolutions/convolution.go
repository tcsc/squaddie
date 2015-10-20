package main

import (
	"github.com/tcsc/squaddie/plugin"
	"image"
	"image/color"
)

type Matrix struct {
	width, height int
	values        []int
	divisor       int
}

func (m *Matrix) At(x, y int) int {
	return m.values[(y*m.width)+x]
}

func clamp(x int) uint8 {
	if x < 0 {
		return 0
	}

	if x > 255 {
		return 255
	}

	return uint8(x)
}

func nrgba(r, g, b, a int) color.NRGBA {
	return color.NRGBA{
		R: clamp(r),
		G: clamp(g),
		B: clamp(b),
		A: clamp(a),
	}
}

func convolve(img *plugin.MMapImage, matrix Matrix) {
	bounds := img.Bounds()
	height := bounds.Dy()
	width := bounds.Dx()
	out := image.NewNRGBA(img.Bounds())

	log.Info("Convolving with %d x %d image", width, height)
	mwOn2 := matrix.width / 2
	mhOn2 := matrix.height / 2
	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			var r, g, b int
			for y := 0; y < matrix.height; y++ {
				dy := y - mhOn2
				for x := 0; x < matrix.width; x++ {
					dx := x - mwOn2
					coef := matrix.At(x, y)
					p := img.NRGBAAt(px+dx, py+dy)
					r += (coef * int(p.R))
					g += (coef * int(p.G))
					b += (coef * int(p.B))
				}
			}

			out.SetNRGBA(px, py, nrgba(r, g, b, 255))
		}
	}

	// memcpy the result back over the input buffer
	copy(img.Pix(), out.Pix)
}

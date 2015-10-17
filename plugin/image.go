package plugin

import (
	"errors"
	"image"
	"image/color"
)

/// Defines an RGBA image backed by shared memory region
type MMapImage struct {
	bounds image.Rectangle
	stride int
	region region
}

// Creates a new Memory-Mapped image
func NewMMapImage(name string, rect image.Rectangle) (img *MMapImage, err error) {

	stride := rect.Dx() * 4
	cbImage := stride * rect.Dy()
	region, err := NewRegion(name, cbImage)
	if err != nil {
		return
	}

	img = &MMapImage{bounds: rect, stride: stride, region: region}
	return
}

func OpenMMapImage(name string, rect image.Rectangle) (img *MMapImage, err error) {

	region, err := OpenRegion(name)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			region.Close()
		}
	}()

	stride := rect.Dx() * 4
	cbImage := stride * rect.Dy()

	if len(region.bytes) != cbImage {
		err = errors.New("Image size mismatch")
		return
	}

	img = &MMapImage{bounds: rect, stride: stride, region: region}
	return
}

func (img *MMapImage) Close() {
	img.region.Close()
}

func (img *MMapImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (img *MMapImage) Bounds() image.Rectangle {
	return img.bounds
}

func (img *MMapImage) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(img.bounds)) {
		return color.RGBA{}
	}

	i := (y * img.stride) + x
	px := img.region.bytes
	return color.RGBA{
		px[i], px[i+1], px[i+2], px[i+3],
	}
}

func (img *MMapImage) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(img.bounds)) {
		return
	}

	i := (y * img.stride) + x
	px := img.region.bytes
	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	px[i+0] = rgba.R
	px[i+1] = rgba.G
	px[i+2] = rgba.B
	px[i+3] = rgba.A
}

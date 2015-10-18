package main

import (
	"github.com/tcsc/squaddie/plugin"
	"image"
	_ "image/jpeg"
	"os"
)

func loadImage(filename string, name string) (*plugin.MMapImage, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	src, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := src.Bounds()

	mmap, err := plugin.NewMMapImage(name, bounds)
	if err != nil {
		return nil, err
	}

	height := bounds.Dy()
	width := bounds.Dx()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			mmap.Set(x, y, src.At(x, y))
		}
	}

	return mmap, nil
}

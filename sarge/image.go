package main

import (
	"github.com/tcsc/squaddie/plugin"
	"image"
	"image/jpeg"
	"os"
)

func loadImage(filename string, name string) (*plugin.MMapImage, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	log.Debug("Decoding jpeg...")

	src, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := src.Bounds()

	log.Debug("Creating shared image")

	mmap, err := plugin.NewMMapImage(name, bounds)
	if err != nil {
		return nil, err
	}

	log.Debug("Converting image format")

	height := bounds.Dy()
	width := bounds.Dx()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			mmap.Set(x, y, src.At(x, y))
		}
	}

	log.Debug("Done!")

	return mmap, nil
}

func saveImage(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, nil)
}

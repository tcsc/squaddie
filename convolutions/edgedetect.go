package main

import (
	"github.com/tcsc/squaddie/plugin"
)

type EdgeDetect struct {
}

var callcount int32 = 0

func (self *EdgeDetect) Invoke(args plugin.InvokeArgs, reply *plugin.InvokeReply) error {
	log.Info("Edge-detect plugin invoked on memory region \"%s\"", args.Region)

	img, err := plugin.OpenMMapImage(args.Region, args.Bounds)
	if err != nil {
		log.Error("Failed to map image region: %s", err.Error())
		return err
	}
	defer img.Close()

	matrix := Matrix{width: 3, height: 3,
		values: []int{
			0, 1, 0,
			1, -4, 1,
			0, 1, 0,
		}}

	convolve(img, matrix)
	return nil
}

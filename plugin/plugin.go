package plugin

import (
	"image"
)

type InvokeArgs struct {
	Bounds image.Rectangle
	Region string
}

type InvokeReply struct {
}

type Plugin interface {
	Invoke(args InvokeArgs, reply *InvokeReply) error
}

type MappedImage struct {
}

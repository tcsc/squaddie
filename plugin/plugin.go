package plugin

import ()

type InvokeArgs struct {
}

type InvokeReply struct {
}

type Plugin interface {
	Invoke(args InvokeArgs, reply *InvokeReply) error
}

type MappedImage struct {
}

package main

import (
	"github.com/tcsc/squaddie/plugin"
	"time"
)

type EdgeDetect struct {
}

var callcount int32 = 0

func (self *EdgeDetect) Invoke(args plugin.InvokeArgs, reply *plugin.InvokeReply) error {
	select {
	case <-time.After(10 * time.Second):
		log.Info("Wait has elapsed")
	}
	reply = &plugin.InvokeReply{}
	log.Info("Leaving Invoke")
	return nil
}

package main

import (
	"fmt"
	pflag "github.com/ogier/pflag"
	logging "github.com/op/go-logging"
	"github.com/tcsc/squaddie/plugin"
	"os"
	"os/signal"
	"sync/atomic"
	"time"
)

var log = logging.MustGetLogger("main")

type EdgeDetect struct {
}

var callcount int32 = 0

func (self *EdgeDetect) Invoke(args plugin.InvokeArgs, reply *plugin.InvokeReply) error {
	count := atomic.AddInt32(&callcount, 1)
	log.Info("%d: Entering Invoke", count)

	select {
	case <-time.After(10 * time.Second):
		log.Info("%d: Wait has elapsed", count)
	}
	reply = &plugin.InvokeReply{}
	log.Info("%d: Leaving Invoke", count)
	return nil
}

func main() {
	os.Exit(run())
}

func run() int {
	log.Info("Starting edge detect service")
	args, err := plugin.ParseCommandLine(os.Args[1:])
	if err != nil {
		if err != pflag.ErrHelp {
			log.Error("Error: %s", err.Error())
		}
		return 1
	}

	log.Info("Trapping signals...")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	sarge, err := plugin.NewClient(args.Network, args.Path)
	if err != nil {
		log.Error("Failed to connect to RPC server: %s", err.Error())
		return 1
	}
	log.Info("Connected!")
	defer sarge.Close()

	ed := EdgeDetect{}
	svr, err := plugin.StartRpc(&ed, "unix", "/tmp/edge-detect.sock")
	if err != nil {
		log.Error("Failed to start RPC services: %s", err.Error())
		return 1
	}
	defer os.RemoveAll("/tmp/edge-detect.sock")

	log.Info("Registering Edge detection plugin")
	cookie, err := sarge.Register("Edge detection", "EdgeDetect.Invoke", svr.Addr())
	if err != nil {
		log.Info("Plugin registration failed: %s", err.Error())
		return 2
	}
	log.Info("Registered with cookie \"%s\"", cookie)

	log.Info("Waiting for kill signal")
	select {
	case sig := <-ch:
		fmt.Printf("Caught %d\n", sig)
	}

	print("Exiting\n")
	return 0
}

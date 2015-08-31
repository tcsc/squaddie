package main

import (
	"fmt"
	pflag "github.com/ogier/pflag"
	logging "github.com/op/go-logging"
	"github.com/tcsc/squaddie/plugin"
	"net/rpc"
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
	log.Info("Starting edge detect service")
	args, err := plugin.ParseCommandLine(os.Args[1:])
	if err != nil {
		if err != pflag.ErrHelp {
			log.Error("Error: %s", err.Error())
		}
		os.Exit(1)
	}

	log.Info("Trapping signals...")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	log.Info("Connecting to %s://%s", args.Network, args.Path)
	client, err := rpc.Dial(args.Network, args.Path)
	if err != nil {
		log.Error("Failed to connect to RPC server: %s", err.Error())
		os.Exit(1)
	}
	log.Info("Connected!")

	ed := EdgeDetect{}
	err = plugin.StartRpc(&ed, "unix", "/tmp/edge-detect.sock")
	if err != nil {
		log.Error("Failed to start RPC services: %s", err.Error())
		os.Exit(1)
	}
	defer os.RemoveAll("/tmp/edge-detect.sock")

	var cookie string
	pi := plugin.PluginInfo{
		Name:    "Edge Detect",
		Network: "unix",
		Path:    "/tmp/edge-detect.sock",
	}

	log.Info("Registering Edge detection plugin")
	err = client.Call("Registrar.RegisterPlugin", pi, &cookie)
	if err != nil {
		log.Info("Plugin registration failed: %s", err.Error())
		os.Exit(2)
	}
	log.Info("Registered with cookie \"%s\"", cookie)

	select {
	case sig := <-ch:
		fmt.Printf("Caught %d\n", sig)
	}

	print("Exiting\n")
}

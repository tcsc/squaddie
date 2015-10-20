package main

import (
	"fmt"
	pflag "github.com/ogier/pflag"
	logging "github.com/op/go-logging"
	"github.com/tcsc/squaddie/plugin"
	"os"
	"os/signal"
)

const url = "/tmp/squaddie-convolutions.sock"

var log = logging.MustGetLogger("main")

var format = logging.MustStringFormatter(
	"%{color}%{time:2006-01-02 15:04:05}%{color:reset} conv> %{message}")

func initLogging() {
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	formatted := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(formatted)
}

func run() int {
	initLogging()

	log.Info("Starting image convolution service")
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

	sarge, err := plugin.NewRegistrarClient(args.Network, args.Path)
	if err != nil {
		log.Error("Failed to connect to RPC server: %s", err.Error())
		return 1
	}
	log.Info("Connected!")
	defer sarge.Close()

	ed := EdgeDetect{}
	svr, err := plugin.StartRpc(&ed, "unix", url)
	if err != nil {
		log.Error("Failed to start RPC services: %s", err.Error())
		return 1
	}
	defer os.RemoveAll(url)

	log.Info("Registering convolution plugins")
	cookie, err := sarge.Register("Edge Detect", "EdgeDetect.Invoke", svr.Addr())
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

func main() {
	os.Exit(run())
}

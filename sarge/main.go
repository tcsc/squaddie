package main

import (
	pflag "github.com/ogier/pflag"
	logging "github.com/op/go-logging"
	"github.com/tcsc/squaddie/plugin"
	"net/rpc"
	"os"
	"os/signal"
)

const (
	url = "/tmp/sarge.sock"
)

var log = logging.MustGetLogger("main")

type Args struct {
	ImageFile string
}

func parseCommandLine(cmdLine []string) (Args, error) {
	var result Args
	flags := pflag.NewFlagSet("Sarge", pflag.ContinueOnError)
	flags.StringVarP(&result.ImageFile, "image", "i", "",
		"The image file to load")
	err := flags.Parse(cmdLine)
	return result, err
}

func StopRegistrar() {
	log.Info("Stopping registrar")
	err := os.Remove(url)
	if err != nil {
		log.Error("Failed to remove registrar socket: %s", err.Error())
	}
}

func OnRegistration(info *plugin.PluginInfo) {
	log.Info("Connecting to %s service at %s://%s", info.Name,
		info.Network, info.Path)
	_, err := rpc.Dial(info.Network, info.Path)
	if err != nil {
		log.Error("Failed to connect to RPC server: %s", err.Error())
		os.Exit(1)
	}
}

func run() int {
	// make sure our socket is removed when we exit
	defer os.Remove(url)

	args, err := parseCommandLine(os.Args[1:])
	if err != nil {
		if err != pflag.ErrHelp {
			log.Error("Error: %s", err.Error())
		}
		return 1
	}

	log.Info("Trapping signals")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	rpcServer, err := plugin.NewRpcServer("unix", url)
	if err != nil {
		log.Error("Failed to start RCP server")
		return 1
	}
	defer rpcServer.Close()
	log.Info("RPC server started on %s", rpcServer.Addr().String())

	log.Info("Adding registrar to RPC server")
	r := plugin.NewRegistrar()
	r.OnRegistration = OnRegistration
	err = rpcServer.Register(r)
	if err != nil {
		log.Error("Failed: %s", err.Error())
		return 1
	}

	plugin := NewPlugin("edge-detect", rpcServer.Addr())
	ch := plugin.Start()

	log.Info("Loading image %s", args.ImageFile)
	img, err := loadImage(args.ImageFile)
	if err != nil {
		log.Error("Failed loading image: %s", err.Error())
		return 3
	}

	log.Info("Image is a %T", img)

	for {
		select {
		case sig := <-signals:
			log.Info("Detected signal %d", sig)
			plugin.Stop()

		case err := <-ch:
			if err != nil {
				log.Error("Plugin execution failed: %s", err.Error())
				return 2
			} else {
				log.Info("Plugin exited cleanly")
				return 0
			}
		}
	}

	return 0
}

func main() {
	os.Exit(run())
}

package main

import (
	pflag "github.com/ogier/pflag"
	logging "github.com/op/go-logging"
	"github.com/tcsc/squaddie/plugin"
	"os"
	"os/signal"
)

const (
	url = "/tmp/squaddie-sarge.sock"
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
		log.Error("Failed to start RPC server: %s", err.Error())
		return 1
	}
	defer rpcServer.Close()
	log.Info("RPC server started on %s", rpcServer.Addr().String())

	// receive notifications abut plugin registrations on this channel
	regCh := make(chan plugin.Info, 1)

	log.Info("Creating registrar")
	registrar := plugin.NewRegistrar(func(info plugin.Info) {
		regCh <- info
	})

	log.Info("Adding registrar to RPC server")
	err = rpcServer.Register(registrar)
	if err != nil {
		log.Error("Failed: %s", err.Error())
		return 1
	}

	log.Info("Launching plugin server")
	pluginServer := NewPlugin("convolutions", rpcServer.Addr())
	ch := pluginServer.Start()

	log.Info("Waiting for plugin to register...")
	info := <-regCh

	log.Info("Connecting back to plugin service")
	client, err := plugin.NewClient(info)
	if err != nil {
		log.Error("Failed: %s", err.Error())
		return 4
	}

	log.Info("Loading image %s", args.ImageFile)
	img, err := loadImage(args.ImageFile, "squaddie-img-buffer")
	if err != nil {
		log.Error("Failed loading image: %s", err.Error())
		return 3
	}
	defer img.Close()

	log.Info("Image is %d x %d pixels", img.Bounds().Dx(), img.Bounds().Dy())

	log.Info("saving pre-process image...")
	err = saveImage(img, "pre-out.jpg")
	if err != nil {
		log.Error("Image save failed: %s", err)
	}

	log.Info("Invoking plugin")
	err = client.Invoke(img)
	if err != nil {
		log.Error("Plugin invocation failed: %s", err)
	}

	log.Info("saving post-process image...")
	err = saveImage(img, "post-out.jpg")
	if err != nil {
		log.Error("Image save failed: %s", err)
	}

	log.Info("Done. Waiting for kill signal")

	for {
		select {
		case sig := <-signals:
			log.Info("Detected signal %d", sig)
			pluginServer.Stop()

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

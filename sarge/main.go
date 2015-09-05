package main

import (
	"fmt"
	logging "github.com/op/go-logging"
	"github.com/tcsc/squaddie/plugin"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

const (
	url = "/tmp/sarge.sock"
)

var log = logging.MustGetLogger("main")

type Plugin struct {
	command *exec.Cmd
}

func NewPlugin(name string, registrar net.Addr) *Plugin {
	cmd := exec.Command(name,
		fmt.Sprintf("--network=%s", registrar.Network()),
		fmt.Sprintf("--path=%s", registrar.String()))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return &Plugin{command: cmd}
}

func (p *Plugin) Start() <-chan error {
	ch := make(chan error)
	go func() {
		log.Info("Running subprocess")
		ch <- p.command.Run()
		log.Info("Subprocess has returned")
	}()
	return ch
}

func (p *Plugin) Stop() error {
	log.Info("Killing the subprocess...")
	p.command.Process.Kill()
	return nil
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
	client, err := rpc.Dial(info.Network, info.Path)
	if err != nil {
		log.Error("Failed to connect to RPC server: %s", err.Error())
		os.Exit(1)
	}

	asyncArgs := plugin.InvokeArgs{}
	var asyncReply plugin.InvokeReply
	log.Info("Invoking async call...")
	call := client.Go("EdgeDetect.Invoke", asyncArgs, &asyncReply, nil)

	<-time.After(5 * time.Second)

	args := plugin.InvokeArgs{}
	var reply plugin.InvokeReply
	log.Info("Invoking sync call...")
	err = client.Call("EdgeDetect.Invoke", args, &reply)
	if err != nil {
		log.Error("Failed to call RPC method: %s", err.Error())
		os.Exit(1)
	}

	log.Info("Waiting got async call to finish")
	<-call.Done
	if call.Error != nil {
		log.Error("Failed to call RPC method: %s", call.Error.Error())
		os.Exit(1)
	}
	log.Info("All done")
}

func main() {
	os.Exit(run())
}

func run() int {
	// make sure our socket is removed when we exit
	defer os.Remove(url)

	log.Info("Binding signals")
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

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

func NewPlugin(name string, network, address string) *Plugin {
	cmd := exec.Command(name,
		fmt.Sprintf("--network=%s", network),
		fmt.Sprintf("--path=%s", address))
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

func StartRegistrar() (network, addr string, err error) {
	log.Info("Starting plugin registrar")

	r := plugin.NewRegistrar()
	r.OnRegistration = OnRegistration
	rpc.Register(r)
	listener, err := net.Listen("unix", url)
	if err != nil {
		log.Error("Failed to listen: %s", err.Error())
		return
	}

	go rpc.Accept(listener)

	network = listener.Addr().Network()
	addr = listener.Addr().String()
	return
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
	log.Info("Binding signals")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	net, addr, err := StartRegistrar()
	if err != nil {
		return
	}
	defer StopRegistrar()
	log.Info("Registrar started on %s://%s", net, addr)

	plugin := NewPlugin("edge-detect", net, addr)
	ch := plugin.Start()

	for {
		select {
		case sig := <-signals:
			log.Info("Detected signal %d", sig)
			plugin.Stop()

		case err := <-ch:
			if err != nil {
				log.Error("Plugin execution failed: %s", err.Error())
			} else {
				log.Info("Plugin exited cleanly")
			}
			return
		}
	}
}

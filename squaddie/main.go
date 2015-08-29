package main

import (
	logging "github.com/op/go-logging"
	"os"
	"os/exec"
	"os/signal"
)

var log = logging.MustGetLogger("main")

type Plugin struct {
	command *exec.Cmd
}

func NewPlugin(name string) *Plugin {
	return &Plugin{command: exec.Command(name)}
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

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	plugin := NewPlugin("edge-detect")
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

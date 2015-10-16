package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
)

type Plugin struct {
	name    string
	command *exec.Cmd
}

func NewPlugin(name string, registrar net.Addr) *Plugin {
	cmd := exec.Command(name,
		fmt.Sprintf("--network=%s", registrar.Network()),
		fmt.Sprintf("--path=%s", registrar.String()))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return &Plugin{name: name, command: cmd}
}

func (p *Plugin) Start() <-chan error {
	ch := make(chan error)
	go func() {
		log.Info("Running subprocess for plugin \"%s\"", p.name)
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

type pluginSession struct {
}

// func NewPluginSession() (*pluginSession, error) {

// }

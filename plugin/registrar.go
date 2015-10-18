package plugin

import (
	"errors"
	logging "github.com/op/go-logging"
)

var AlreadyRegistered = errors.New("Plugin already registered")

type Event func(Info)

type Registrar struct {
	plugins map[string]Info
	ch      chan regRequest
	onReg   Event
}

var regLog = logging.MustGetLogger("registrar")

type Info struct {
	Name    string
	Network string
	Path    string
	Service string
}

type regReply struct {
	err    error
	cookie string
}

type regRequest struct {
	Info
	replyCh chan<- regReply
}

func NewRegistrar(callback Event) *Registrar {
	ch := make(chan regRequest, 1)
	registrar := &Registrar{
		plugins: make(map[string]Info),
		ch:      ch,
		onReg:   callback,
	}
	go registrar.run()
	return registrar
}

func (r *Registrar) run() {
	regLog.Info("Entering registrar run loop")
	for {
		select {
		case request := <-r.ch:
			cookie, err := r.registerPlugin(request.Info)
			request.replyCh <- regReply{err: err, cookie: cookie}
		}
	}
	regLog.Info("Exiting registrar run loop")
}

// RegisterPlugin privides an RPC-compatible plugin registration
// end-point...
func (r *Registrar) RegisterPlugin(info Info, cookie *string) error {
	regLog.Info("Received registration request. Signalling registrar.")

	replyCh := make(chan regReply, 1)
	defer close(replyCh)
	r.ch <- regRequest{Info: info, replyCh: replyCh}

	regLog.Info("Waiting for response from registrar...")
	reply := <-replyCh

	regLog.Info("Plugin \"%s\" registered with cookie \"%s\"",
		info.Name, reply.cookie)
	(*cookie) = reply.cookie
	return reply.err
}

// private implementation. Should only be called from inside the message
// handler loop
func (r *Registrar) registerPlugin(info Info) (string, error) {
	regLog.Info("Registering plugin %s at %s://%s", info.Name, info.Network,
		info.Path)
	if _, ok := r.plugins[info.Name]; ok {
		regLog.Error("Already registered")
		return "", AlreadyRegistered
	}
	r.plugins[info.Name] = info
	cookie := "some-cookie"

	if r.onReg != nil {
		go r.onReg(info)
	}

	return cookie, nil
}

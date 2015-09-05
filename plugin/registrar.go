package plugin

import (
	"errors"
	logging "github.com/op/go-logging"
)

var AlreadyRegistered = errors.New("Plugin already registered")

type PluginEvent func(*PluginInfo)

type Registrar struct {
	plugins        map[string]PluginInfo
	OnRegistration PluginEvent
}

var regLog = logging.MustGetLogger("registrar")

type PluginInfo struct {
	Name    string
	Network string
	Path    string
	Service string
}

func NewRegistrar() *Registrar {
	return &Registrar{
		plugins: make(map[string]PluginInfo),
	}
}

// RegisterPlugin privides an RPC-compatible plugin
// registration end point.
//
// TODO: handle multiple simultaneous registrations
func (r *Registrar) RegisterPlugin(info PluginInfo, cookie *string) error {
	regLog.Info("Registering plugin %s at %s://%s", info.Name, info.Network,
		info.Path)
	if _, ok := r.plugins[info.Name]; ok {
		regLog.Error("Already registered")
		return AlreadyRegistered
	}
	r.plugins[info.Name] = info
	(*cookie) = "some-cookie"

	if r.OnRegistration != nil {
		go r.OnRegistration(&info)
	}

	return nil
}

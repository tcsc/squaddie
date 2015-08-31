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
}

func NewRegistrar() *Registrar {
	regLog.Info("Creating registrar")
	return &Registrar{
		plugins: make(map[string]PluginInfo),
	}
}

func (r *Registrar) RegisterPlugin(info PluginInfo, cookie *string) error {
	regLog.Info("Registering plugin %s at %s://%s", info.Name, info.Network,
		info.Path)
	if _, ok := r.plugins[info.Name]; ok {
		return AlreadyRegistered
	}
	r.plugins[info.Name] = info
	(*cookie) = "some-cookie"

	if r.OnRegistration != nil {
		go r.OnRegistration(&info)
	}

	return nil
}

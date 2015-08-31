package plugin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_RegisterPlugin(t *testing.T) {
	registrar := NewRegistrar()
	info := PluginInfo{
		Name: "Dummy Plugin",
		Url:  "unix:///var/run/squaddie/dummy.sock",
	}
	err := registrar.RegisterPlugin(info)
	assert.NoError(t, err)
}

func Test_ReregisteringPluginFails(t *testing.T) {
	registrar := NewRegistrar()
	info := PluginInfo{
		Name: "Dummy Plugin",
		Url:  "unix:///var/run/squaddie/dummy.sock",
	}
	err := registrar.RegisterPlugin(info)
	assert.NoError(t, err)
	err = registrar.RegisterPlugin(info)
	assert.Error(t, err)
}

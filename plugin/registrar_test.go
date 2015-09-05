package plugin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// ----------------------------------------------------------------------------
//
// ----------------------------------------------------------------------------

func Test_RegisterPlugin(t *testing.T) {
	registrar := NewRegistrar()

	info := PluginInfo{
		Name:    "Dummy Plugin",
		Network: "unix",
		Path:    "/var/run/squaddie/dummy.sock",
	}
	var cookie string
	err := registrar.RegisterPlugin(info, &cookie)
	assert.NoError(t, err)
}

func Test_ReregisteringPluginFails(t *testing.T) {
	registrar := NewRegistrar()

	info := PluginInfo{
		Name:    "Dummy Plugin",
		Network: "unix",
		Path:    "/var/run/squaddie/dummy.sock",
	}
	var cookie string
	err := registrar.RegisterPlugin(info, &cookie)
	assert.NoError(t, err)
	assert.NotEmpty(t, cookie)

	err = registrar.RegisterPlugin(info, &cookie)
	assert.Error(t, err)
}

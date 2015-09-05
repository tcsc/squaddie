package plugin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ParseCliReturnsHelp(t *testing.T) {
	_, err := ParseCommandLine([]string{
		"some-binary", "--help",
	})
	assert.Error(t, err)
}

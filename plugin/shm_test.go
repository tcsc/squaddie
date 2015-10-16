package plugin

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	//"fmt"
	"testing"
)

func Test_ShmOpen(t *testing.T) {
	region, err := shmCreate("/squaddie-shm-test", 1024)
	require.NoError(t, err, "shmCreate must not fail")
	defer region.Close()
	assert.Equal(t, 1024, len(region.bytes))
}

package plugin

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_NewRegion(t *testing.T) {
	region, err := NewRegion("/squaddie-shm-test", 1024)
	require.NoError(t, err, "NewRegion must not fail")
	defer region.Close()
	assert.Equal(t, 1024, len(region.bytes))
}

func Test_OpenRegionMapsAsExpected(t *testing.T) {
	region_name := "/squaddie-shm-test-2"
	srcRegion, err := NewRegion(region_name, 1024)
	require.NoError(t, err, "NewRegion must not fail")
	defer srcRegion.Close()
	assert.Equal(t, 1024, len(srcRegion.bytes))

	dstRegion, err := OpenRegion(region_name)
	require.NoError(t, err, "OpenRegion must not fail")
	defer dstRegion.Close()
	assert.Equal(t, 1024, len(dstRegion.bytes))

	srcRegion.bytes[0] = 0x00
	srcRegion.bytes[1] = 0x0F
	srcRegion.bytes[2] = 0xF0
	srcRegion.bytes[3] = 0xFF

	assert.Equal(t, byte(0x00), dstRegion.bytes[0])
	assert.Equal(t, byte(0x0F), dstRegion.bytes[1])
	assert.Equal(t, byte(0xF0), dstRegion.bytes[2])
	assert.Equal(t, byte(0xFF), dstRegion.bytes[3])
}

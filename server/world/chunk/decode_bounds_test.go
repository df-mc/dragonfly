package chunk

import (
	"bytes"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

func TestNetworkDecodeRejectsInvalidSubChunkBounds(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		count   int
	}{
		{name: "declared count", payload: bytes.Repeat([]byte{8, 0}, 9), count: 9},
		{name: "encoded index above range", payload: []byte{9, 0, 8}, count: 1},
		{name: "encoded index below range", payload: []byte{9, 0, 0xff}, count: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NetworkDecode(decodeTestRegistry{}, tt.payload, tt.count, cube.Range{0, 127}); err == nil {
				t.Fatal("expected malformed sub-chunk bounds to return an error")
			}
		})
	}
}

type decodeTestRegistry struct{}

func (decodeTestRegistry) BlockCount() int      { return 1 }
func (decodeTestRegistry) AirRuntimeID() uint32 { return 0 }
func (decodeTestRegistry) RuntimeIDToState(uint32) (string, map[string]any, bool) {
	return "minecraft:air", nil, true
}
func (decodeTestRegistry) StateToRuntimeID(string, map[string]any) (uint32, bool) { return 0, true }
func (decodeTestRegistry) FilteringBlock(uint32) uint8                            { return 0 }
func (decodeTestRegistry) LightBlock(uint32) uint8                                { return 0 }
func (decodeTestRegistry) RandomTickBlock(uint32) bool                            { return false }
func (decodeTestRegistry) NBTBlock(uint32) bool                                   { return false }
func (decodeTestRegistry) LiquidDisplacingBlock(uint32) bool                      { return false }
func (decodeTestRegistry) LiquidBlock(uint32) bool                                { return false }
func (decodeTestRegistry) HashToRuntimeID(uint32) (uint32, bool)                  { return 0, true }
func (decodeTestRegistry) RuntimeIDToHash(uint32) (uint32, bool)                  { return 0, true }

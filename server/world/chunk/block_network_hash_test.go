package chunk

import (
	"bytes"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

func TestSubChunkConvertBlockNetworkHashesToRuntimeIDs(t *testing.T) {
	br := networkHashTestRegistry{
		air:             0,
		hashToRuntimeID: map[uint32]uint32{100: 1, 200: 2},
	}
	sub := NewSubChunk(br.AirRuntimeID())
	storage := sub.Layer(0)
	storage.Set(1, 2, 3, 100)
	storage.Set(4, 5, 6, 999)

	sub.ConvertBlockNetworkHashesToRuntimeIDs(br)

	if got := sub.Block(1, 2, 3, 0); got != 1 {
		t.Fatalf("known hash converted to %d, want runtime ID 1", got)
	}
	if got := sub.Block(4, 5, 6, 0); got != 999 {
		t.Fatalf("unknown hash converted to %d, want original value 999", got)
	}
}

func TestChunkConvertBlockNetworkHashesToRuntimeIDs(t *testing.T) {
	br := networkHashTestRegistry{
		air:             0,
		hashToRuntimeID: map[uint32]uint32{100: 1},
	}
	c := New(br, cube.Range{0, 15})
	c.SetBlock(1, 2, 3, 0, 100)

	c.ConvertBlockNetworkHashesToRuntimeIDs()

	if got := c.Block(1, 2, 3, 0); got != 1 {
		t.Fatalf("known hash converted to %d, want runtime ID 1", got)
	}
}

func TestEncodeWithBlockNetworkHashesConvertsRuntimeIDsWithoutMutatingChunk(t *testing.T) {
	br := networkHashTestRegistry{
		air:             0,
		runtimeIDToHash: map[uint32]uint32{1: 100, 2: 200},
	}
	c := New(br, cube.Range{0, 15})
	c.SetBlock(1, 2, 3, 0, 1)
	c.SetBlock(4, 5, 6, 0, 999)

	data := EncodeWithBlockNetworkHashes(c)
	buf := bytes.NewBuffer(data.SubChunks[0])
	index := byte(0)
	decoded, err := decodeSubChunk(buf, c, &index, NetworkEncoding)
	if err != nil {
		t.Fatalf("decode encoded subchunk: %v", err)
	}

	if got := decoded.Block(1, 2, 3, 0); got != 100 {
		t.Fatalf("runtime ID encoded as %d, want network hash 100", got)
	}
	if got := decoded.Block(4, 5, 6, 0); got != 999 {
		t.Fatalf("unknown runtime ID encoded as %d, want original value 999", got)
	}
	if got := c.Block(1, 2, 3, 0); got != 1 {
		t.Fatalf("cached chunk mutated to %d, want original runtime ID 1", got)
	}
}

type networkHashTestRegistry struct {
	air             uint32
	hashToRuntimeID map[uint32]uint32
	runtimeIDToHash map[uint32]uint32
}

func (r networkHashTestRegistry) BlockCount() int { return 1000 }
func (r networkHashTestRegistry) AirRuntimeID() uint32 {
	return r.air
}
func (networkHashTestRegistry) RuntimeIDToState(uint32) (string, map[string]any, bool) {
	return "test:block", nil, true
}
func (networkHashTestRegistry) StateToRuntimeID(string, map[string]any) (uint32, bool) {
	return 0, true
}
func (networkHashTestRegistry) FilteringBlock(uint32) uint8       { return 0 }
func (networkHashTestRegistry) LightBlock(uint32) uint8           { return 0 }
func (networkHashTestRegistry) RandomTickBlock(uint32) bool       { return false }
func (networkHashTestRegistry) NBTBlock(uint32) bool              { return false }
func (networkHashTestRegistry) LiquidDisplacingBlock(uint32) bool { return false }
func (networkHashTestRegistry) LiquidBlock(uint32) bool           { return false }
func (r networkHashTestRegistry) HashToRuntimeID(hash uint32) (uint32, bool) {
	runtimeID, ok := r.hashToRuntimeID[hash]
	return runtimeID, ok
}
func (r networkHashTestRegistry) RuntimeIDToHash(runtimeID uint32) (uint32, bool) {
	hash, ok := r.runtimeIDToHash[runtimeID]
	return hash, ok
}

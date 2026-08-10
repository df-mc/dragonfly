package chunk

import (
	"math/rand"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// TestCompactPreservesLayerNumbers verifies that Compact does not renumber layers. Layer 0 is the block and layer 1 is
// the waterlogging layer, so dropping an all-air layer 0 would turn waterlogging into a solid liquid block.
func TestCompactPreservesLayerNumbers(t *testing.T) {
	tests := []struct {
		name  string
		layer uint8
	}{
		{name: "waterlogging layer above an air block layer", layer: 1},
		{name: "layer above several air layers", layer: 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(testRegistry{}, cube.Range{0, 15})
			c.SetBlock(5, 0, 5, tt.layer, 1)

			c.Compact()

			if got := c.Block(5, 0, 5, tt.layer); got != 1 {
				t.Errorf("Compact() block at layer %v = %v, want 1", tt.layer, got)
			}
			if got := c.Block(5, 0, 5, 0); got != 0 {
				t.Errorf("Compact() block at layer 0 = %v, want 0 (air): it was renumbered down from layer %v", got, tt.layer)
			}
		})
	}
}

// TestCompactDropsTrailingAirLayers verifies that Compact still drops trailing all-air storages. A layer past the last
// stored layer already reads as air, so those carry no information.
func TestCompactDropsTrailingAirLayers(t *testing.T) {
	c := New(testRegistry{}, cube.Range{0, 15})
	c.SetBlock(5, 0, 5, 0, 1)
	sub := c.sub[c.SubIndex(0)]
	sub.storages = append(sub.storages, emptyStorage(0), emptyStorage(0))

	c.Compact()

	if got := len(sub.storages); got != 1 {
		t.Fatalf("Compact() left %v storages, want 1", got)
	}
	if got := c.Block(5, 0, 5, 0); got != 1 {
		t.Fatalf("Block(layer 0) = %v, want 1", got)
	}
}

// TestEncodeStorageCountFitsInAByte verifies that a chunk grown to the highest addressable layer still encodes its
// storage count without overflowing the single byte it is written into. This is the limit MaxLayers exists to keep.
func TestEncodeStorageCountFitsInAByte(t *testing.T) {
	c := New(testRegistry{}, cube.Range{0, 15})
	c.SetBlock(0, 0, 0, MaxLayers-1, 1)

	// Byte 0 of a sub chunk payload is the version, byte 1 is the storage count.
	if got := Encode(c, NetworkEncoding).SubChunks[c.SubIndex(0)][1]; got != MaxLayers {
		t.Fatalf("encoded storage count = %v, want %v", got, MaxLayers)
	}
}

// testRegistry is a minimal BlockRegistry. Runtime ID 0 is air and any other runtime ID is an opaque non-air block.
type testRegistry struct{}

func (testRegistry) BlockCount() int { return 2 }

func (testRegistry) AirRuntimeID() uint32 { return 0 }

func (testRegistry) RuntimeIDToState(rid uint32) (string, map[string]any, bool) {
	if rid == 0 {
		return "minecraft:air", nil, true
	}
	return "minecraft:stone", nil, true
}

func (testRegistry) StateToRuntimeID(name string, _ map[string]any) (uint32, bool) {
	if name == "minecraft:air" {
		return 0, true
	}
	return 1, true
}

func (testRegistry) FilteringBlock(uint32) uint8 { return 0 }

func (testRegistry) LightBlock(uint32) uint8 { return 0 }

func (testRegistry) RandomTickBlock(uint32) bool { return false }

func (testRegistry) NBTBlock(uint32) bool { return false }

func (testRegistry) LiquidDisplacingBlock(uint32) bool { return false }

func (testRegistry) LiquidBlock(uint32) bool { return false }

func (testRegistry) HashToRuntimeID(hash uint32) (uint32, bool) { return hash, true }

func (testRegistry) RuntimeIDToHash(rid uint32) (uint32, bool) { return rid, true }

// TestCompactPreservesEveryBlock verifies over randomly populated chunks that compaction changes no block. Compaction
// used to drop all-air storages and close the gap, which moved every layer above the one dropped.
func TestCompactPreservesEveryBlock(t *testing.T) {
	type pos struct {
		x, z  uint8
		y     int16
		layer uint8
	}
	for seed := int64(0); seed < 30; seed++ {
		r := rand.New(rand.NewSource(seed))
		c := New(testRegistry{}, cube.Range{0, 255})

		want := map[pos]uint32{}
		for range 300 {
			p := pos{uint8(r.Intn(16)), uint8(r.Intn(16)), int16(r.Intn(256)), uint8(r.Intn(3))}
			rid := uint32(r.Intn(2))
			c.SetBlock(p.x, p.y, p.z, p.layer, rid)
			want[p] = rid
		}

		c.Compact()

		for p, rid := range want {
			if got := c.Block(p.x, p.y, p.z, p.layer); got != rid {
				t.Fatalf("seed %v: Compact() block at (%v,%v,%v) layer %v = %v, want %v", seed, p.x, p.y, p.z, p.layer, got, rid)
			}
		}
	}
}

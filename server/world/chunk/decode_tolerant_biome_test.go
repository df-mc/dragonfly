package chunk

import (
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// TestNetworkDecodeTolerantBiomes verifies that a chunk carrying non-conformant biome sections
// (an unsigned palette entry count instead of the spec's signed zigzag VarInt, plus fewer sections
// than the dimension height) still decodes into a usable chunk rather than being dropped. The
// payload is a real LevelChunk captured from pvp.inpvp.net (overworld, SubChunkCount 1).
func TestNetworkDecodeTolerantBiomes(t *testing.T) {
	raw, err := os.ReadFile("testdata/inpvp_biome_chunk.hex")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	data, err := hex.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("decode hex: %v", err)
	}

	reg := networkHashTestRegistry{air: 0}
	r := cube.Range{-64, 319}

	c, err := NetworkDecode(reg, data, 1, r)
	if err != nil {
		t.Fatalf("expected tolerant decode to succeed, got error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil chunk")
	}
	if len(c.Sub()) != (r.Height()>>4)+1 {
		t.Fatalf("unexpected sub-chunk count %d", len(c.Sub()))
	}
}

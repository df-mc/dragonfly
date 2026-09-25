package session

import (
	"bytes"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
)

func TestBlockHashNetworkBoundary(t *testing.T) {
	br := world.NewBlockRegistry()
	br.Finalize()

	stone := block.Stone{}
	rid := br.BlockRuntimeID(stone)
	hash, ok := br.RuntimeIDToHash(rid)
	if !ok || hash == rid {
		t.Fatalf("expected a distinct network hash for stone: runtime ID %d, hash %d, found %t", rid, hash, ok)
	}
	if got := blockNetworkID(br, stone); got != hash {
		t.Fatalf("network block ID = %d, want hash %d", got, hash)
	}
	if got, ok := blockFromNetworkID(br, hash); !ok || got != stone {
		t.Fatalf("block from hash = %#v, found %t, want stone", got, ok)
	}

	stack := item.NewStack(stone, 3)
	networkStack := stackFromItem(br, stack)
	if got := uint32(networkStack.BlockRuntimeID); got != hash {
		t.Fatalf("item block ID = %d, want hash %d", got, hash)
	}
	if got := stackToItem(br, networkStack); !got.Equal(stack) {
		t.Fatalf("item round trip = %#v, want %#v", got, stack)
	}

	r := cube.Range{0, 15}
	c := chunk.New(br, r)
	c.SetBlock(1, 2, 3, 0, rid)
	hashData := chunk.Encode(c, chunk.NetworkHashEncoding)
	runtimeData := chunk.Encode(c, chunk.NetworkEncoding)
	if bytes.Equal(hashData.SubChunks[0], runtimeData.SubChunks[0]) {
		t.Fatal("hash and runtime ID sub-chunk encodings are identical")
	}
	for _, tc := range []struct {
		name   string
		data   chunk.SerialisedData
		decode func(chunk.BlockRegistry, []byte, int, cube.Range) (*chunk.Chunk, error)
	}{
		{name: "hash", data: hashData, decode: chunk.NetworkHashDecode},
		{name: "runtime ID", data: runtimeData, decode: chunk.NetworkDecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wire := append(append([]byte(nil), tc.data.SubChunks[0]...), tc.data.Biomes...)
			decoded, err := tc.decode(br, wire, 1, r)
			if err != nil {
				t.Fatal(err)
			}
			if got := decoded.Block(1, 2, 3, 0); got != rid {
				t.Fatalf("internal block ID = %d, want runtime ID %d", got, rid)
			}
		})
	}
}

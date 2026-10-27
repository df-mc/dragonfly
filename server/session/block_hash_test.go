package session

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestBlockNetworkPackets(t *testing.T) {
	br := world.NewBlockRegistry()
	br.Finalize()
	stone := block.Stone{}
	rid := br.BlockRuntimeID(stone)
	hash, _ := br.RuntimeIDToHash(rid)
	for _, mode := range []chunk.BlockIDMode{chunk.SequentialBlockIDs, chunk.HashedBlockIDs} {
		hashes := mode == chunk.HashedBlockIDs
		t.Run(fmt.Sprintf("hashes=%t", hashes), func(t *testing.T) {
			s := &Session{
				br:      br,
				conf:    Config{BlockIDs: mode},
				packets: make(chan packet.Packet, 1),
			}
			id := rid
			if hashes {
				id = hash
			}
			s.ViewBlockUpdate(cube.Pos{1, 2, 3}, stone, 0)
			if got := (<-s.packets).(*packet.UpdateBlock).NewBlockRuntimeID; got != id {
				t.Errorf("UpdateBlock ID = %d, want %d", got, id)
			}
			for _, face := range cube.Faces() {
				s.ViewParticle(mgl64.Vec3{1, 2, 3}, particle.PunchBlock{Block: stone, Face: face})
				pk := (<-s.packets).(*packet.LevelEvent)
				event, data := int32(packet.LevelEventParticlesCrackBlock), int32(uint32(face)<<24|rid)
				if hashes {
					event, data = packet.LevelEventParticlesCrackBlockDown+int32(face), int32(hash)
				}
				if pk.EventType != event || pk.EventData != data {
					t.Errorf("face %d: crack event/data = %d/%d, want %d/%d", face, pk.EventType, pk.EventData, event, data)
				}
			}
		})
	}
}

func TestBlockNetworkBoundary(t *testing.T) {
	br := world.NewBlockRegistry()
	br.Finalize()

	stone := block.Stone{}
	rid := br.BlockRuntimeID(stone)
	hash, ok := br.RuntimeIDToHash(rid)
	if !ok || hash == rid || int32(hash) >= 0 {
		t.Fatalf("expected a distinct hash with the top bit set for stone: runtime ID %d, hash %d, found %t", rid, hash, ok)
	}
	r := cube.Range{-16, 15}
	c := chunk.New(br, r)
	c.SetBlock(1, -14, 3, 0, rid)
	c.SetBlock(4, 2, 5, 0, rid)
	c.SetBiome(1, -14, 3, 42)
	runtimeData := chunk.Encode(c, chunk.NetworkEncoding{BlockIDs: chunk.SequentialBlockIDs})

	for _, mode := range []chunk.BlockIDMode{chunk.SequentialBlockIDs, chunk.HashedBlockIDs} {
		hashes := mode == chunk.HashedBlockIDs
		t.Run(fmt.Sprintf("hashes=%t", hashes), func(t *testing.T) {
			s := &Session{br: br, conf: Config{BlockIDs: mode}}
			id := rid
			if hashes {
				id = hash
			}
			if got := br.NetworkID(rid, mode); got != id {
				t.Fatalf("network block ID = %d, want %d", got, id)
			}
			if got, ok := br.RuntimeID(id, mode); !ok || got != rid {
				t.Fatalf("runtime ID from network ID = %d, found %t, want %d", got, ok, rid)
			}

			stack := item.NewStack(stone, 3)
			networkStack := s.stackFromItem(stack)
			if got := uint32(networkStack.BlockRuntimeID); got != id {
				t.Fatalf("item block ID = %d, want %d", got, id)
			}
			if got := s.stackToItem(networkStack); !got.Equal(stack) || got.Count() != stack.Count() {
				t.Fatalf("item round trip = %#v, want %#v", got, stack)
			}

			data := chunk.Encode(c, chunk.NetworkEncoding{BlockIDs: mode})
			if hashes && bytes.Equal(data.SubChunks[0], runtimeData.SubChunks[0]) {
				t.Fatal("hash and sequential ID sub-chunk encodings are identical")
			}
			if !bytes.Equal(data.Biomes, runtimeData.Biomes) {
				t.Fatal("block network ID mode changed biome encoding")
			}
			wire := append(bytes.Join(data.SubChunks, nil), data.Biomes...)
			decoded, err := chunk.NetworkDecode(br, wire, len(data.SubChunks), r, mode)
			if err != nil {
				t.Fatal(err)
			}
			for _, pos := range []cube.Pos{{1, -14, 3}, {4, 2, 5}} {
				if got := decoded.Block(uint8(pos.X()), int16(pos.Y()), uint8(pos.Z()), 0); got != rid {
					t.Errorf("internal block ID at %v = %d, want runtime ID %d", pos, got, rid)
				}
			}
			if got := decoded.Biome(1, -14, 3); got != 42 {
				t.Errorf("decoded biome = %d, want 42", got)
			}
		})
	}
}

// missingHashRegistry writes a hash that is not registered for any block.
type missingHashRegistry struct{ world.BlockRegistry }

// NetworkID returns an unknown hash in hash mode and the original ID in sequential mode.
func (br missingHashRegistry) NetworkID(rid uint32, mode chunk.BlockIDMode) uint32 {
	if mode == chunk.HashedBlockIDs {
		for id := uint32(1); ; id++ {
			if _, ok := br.HashToRuntimeID(id); !ok {
				return id
			}
		}
	}
	return rid
}

func TestUnknownBlockHashDecodesAsAir(t *testing.T) {
	br := world.NewBlockRegistry()
	br.Finalize()
	unknown := missingHashRegistry{br}
	id := unknown.NetworkID(br.BlockRuntimeID(block.Stone{}), chunk.HashedBlockIDs)
	if _, ok := br.RuntimeID(id, chunk.HashedBlockIDs); ok {
		t.Fatal("unknown hash resolved to a runtime ID")
	}
	if _, ok := br.RuntimeID(uint32(br.BlockCount()), chunk.SequentialBlockIDs); ok {
		t.Fatal("out-of-range sequential ID resolved to a runtime ID")
	}

	r := cube.Range{0, 15}
	c := chunk.New(unknown, r)
	c.SetBlock(1, 2, 3, 0, br.BlockRuntimeID(block.Stone{}))
	data := chunk.Encode(c, chunk.NetworkEncoding{BlockIDs: chunk.HashedBlockIDs})
	wire := append(bytes.Join(data.SubChunks, nil), data.Biomes...)
	decoded, err := chunk.NetworkDecode(br, wire, 1, r, chunk.HashedBlockIDs)
	if err != nil {
		t.Fatal(err)
	}
	if got := decoded.Block(1, 2, 3, 0); got != br.AirRuntimeID() {
		t.Fatalf("unknown hash decoded as runtime ID %d, want air %d", got, br.AirRuntimeID())
	}

	s := &Session{br: br}
	stack := s.stackFromItem(item.NewStack(block.Stone{}, 3))
	stack.BlockRuntimeID = int32(id)
	if got := s.stackToItem(stack); !got.Empty() {
		t.Fatalf("unknown item block hash decoded as %#v, want air", got)
	}
}

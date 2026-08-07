package session

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestPrivateBlockAdvertisesFirstEmptySubChunk(t *testing.T) {
	for _, cacheEnabled := range []bool{false, true} {
		t.Run(cacheStateName(cacheEnabled), func(t *testing.T) {
			w := world.Config{Synchronous: true}.New()
			defer w.Close()

			s := newViewLayerTestSession(w, cacheEnabled)
			var advertised *packet.LevelChunk
			err := w.Do(func(tx *world.Tx) {
				s.chunkLoader.Load(tx, 1)
				drainViewLayerTestPackets(s.packets)

				pos := cube.Pos{0, int(w.Range()[0]), 0}
				s.viewLayer.ViewBlock(tx, pos, block.Stone{})
				for _, pk := range drainViewLayerTestPackets(s.packets) {
					if levelChunk, ok := pk.(*packet.LevelChunk); ok {
						advertised = levelChunk
					}
				}
				s.chunkLoader.Close(tx)
			}).Wait(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if advertised == nil {
				t.Fatal("expected a private sub-chunk height advertisement")
			}
			if advertised.HighestSubChunk != 1 {
				t.Fatalf("expected first sub-chunk to be advertised, got height %d", advertised.HighestSubChunk)
			}
			if advertised.SubChunkCount != protocol.SubChunkRequestModeLimited {
				t.Fatalf("expected limited sub-chunk request mode, got %d", advertised.SubChunkCount)
			}
			if advertised.CacheEnabled != cacheEnabled {
				t.Fatalf("expected cache enabled %t, got %t", cacheEnabled, advertised.CacheEnabled)
			}
			if cacheEnabled {
				if len(advertised.BlobHashes) != 1 {
					t.Fatalf("expected one cached biome blob, got %d", len(advertised.BlobHashes))
				}
				if _, ok := s.blobs[advertised.BlobHashes[0]]; !ok {
					t.Fatal("advertised biome blob was not tracked")
				}
				if !bytes.Equal(advertised.RawPayload, []byte{0}) {
					t.Fatalf("expected cached border-block payload, got %v", advertised.RawPayload)
				}
				return
			}
			if len(advertised.BlobHashes) != 0 {
				t.Fatalf("expected no blob hashes without client caching, got %v", advertised.BlobHashes)
			}
			if len(advertised.RawPayload) < 2 || advertised.RawPayload[len(advertised.RawPayload)-1] != 0 {
				t.Fatalf("expected encoded biomes followed by a border-block byte, got %v", advertised.RawPayload)
			}
		})
	}
}

func TestApplyViewLayerToChunkInitialisesNilBlockEntities(t *testing.T) {
	for _, populated := range []bool{false, true} {
		t.Run(mapStateName(populated), func(t *testing.T) {
			w := world.Config{Synchronous: true}.New()
			defer w.Close()

			s := newViewLayerTestSession(w, false)
			pos := cube.Pos{0, int(w.Range()[0]), 0}
			existingPos := pos.Add(cube.Pos{1, 0, 0})
			public := chunk.New(world.DefaultBlockRegistry, w.Range())
			var original map[cube.Pos]world.Block
			if populated {
				original = map[cube.Pos]world.Block{existingPos: block.NewChest()}
			}
			unchanged, unchangedEntities := s.applyViewLayerToChunk(world.ChunkPos{}, public, original)
			if unchanged != public || len(unchangedEntities) != len(original) {
				t.Fatal("expected no copy without an override")
			}

			var (
				visible       *chunk.Chunk
				blockEntities map[cube.Pos]world.Block
			)
			err := w.Do(func(tx *world.Tx) {
				s.viewLayer.ViewBlock(tx, pos, block.NewChest())
				visible, blockEntities = s.applyViewLayerToChunk(world.ChunkPos{}, public, original)
				s.chunkLoader.Close(tx)
			}).Wait(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if visible == public {
				t.Fatal("expected the public chunk to be copied before applying the override")
			}
			if _, ok := blockEntities[pos].(block.Chest); !ok {
				t.Fatalf("expected private chest block entity, got %T", blockEntities[pos])
			}
			if _, ok := original[pos]; ok {
				t.Fatal("private block entity was inserted into the public map")
			}
			if populated {
				if _, ok := blockEntities[existingPos].(block.Chest); !ok {
					t.Fatal("existing block entity was not retained in the visible copy")
				}
				delete(blockEntities, existingPos)
				if _, ok := original[existingPos]; !ok {
					t.Fatal("visible block-entity map aliases the public map")
				}
			}
			if got := public.Block(0, int16(pos[1]), 0, 0); got != world.DefaultBlockRegistry.AirRuntimeID() {
				t.Fatalf("public chunk mutated: got runtime ID %d", got)
			}
		})
	}
}

func TestViewPublicBlockRestoresPublicLiquid(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	s := newViewLayerTestSession(w, false)
	pos := cube.Pos{0, int(w.Range()[0]), 0}
	publicLiquid := block.Water{Still: true, Depth: 8}
	var updates []*packet.UpdateBlock
	err := w.Do(func(tx *world.Tx) {
		s.chunkLoader.Load(tx, 1)
		drainViewLayerTestPackets(s.packets)

		col, ok := s.loadedColumnAt(pos)
		if !ok {
			t.Fatal("expected loaded public column")
		}
		col.SetBlock(0, int16(pos[1]), 0, 0, s.br.BlockRuntimeID(block.Stone{}))
		col.SetBlock(0, int16(pos[1]), 0, 1, s.br.BlockRuntimeID(publicLiquid))

		s.viewLayer.ViewBlock(tx, pos, block.Stone{})
		drainViewLayerTestPackets(s.packets)
		s.viewLayer.ViewPublicBlock(tx, pos)
		for _, pk := range drainViewLayerTestPackets(s.packets) {
			if update, ok := pk.(*packet.UpdateBlock); ok {
				updates = append(updates, update)
			}
		}
		s.chunkLoader.Close(tx)
	}).Wait(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(updates) != 2 {
		t.Fatalf("expected solid and liquid updates, got %d", len(updates))
	}
	for layer, want := range []uint32{s.br.BlockRuntimeID(block.Stone{}), s.br.BlockRuntimeID(publicLiquid)} {
		update := updates[layer]
		if update.Position != (protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])}) {
			t.Fatalf("layer %d restored at wrong position: %v", layer, update.Position)
		}
		if update.Layer != uint32(layer) {
			t.Fatalf("expected layer %d, got %d", layer, update.Layer)
		}
		if update.NewBlockRuntimeID != want {
			t.Fatalf("layer %d: expected runtime ID %d, got %d", layer, want, update.NewBlockRuntimeID)
		}
		if update.Flags != packet.BlockUpdateNetwork {
			t.Fatalf("layer %d: expected network flag, got %d", layer, update.Flags)
		}
	}
}

func newViewLayerTestSession(w *world.World, cacheEnabled bool) *Session {
	s := &Session{
		conn:            viewLayerTestConn{cacheEnabled: cacheEnabled},
		packets:         make(chan packet.Packet, 16),
		blobs:           map[uint64][]byte{},
		closeBackground: make(chan struct{}),
		br:              world.DefaultBlockRegistry,
	}
	s.viewLayer = world.NewViewLayer(s)
	s.viewWorld.Store(w)
	s.chunkLoader = world.NewLoader(1, w, s)
	return s
}

func drainViewLayerTestPackets(packets <-chan packet.Packet) []packet.Packet {
	var drained []packet.Packet
	for len(packets) > 0 {
		drained = append(drained, <-packets)
	}
	return drained
}

func cacheStateName(enabled bool) string {
	if enabled {
		return "cache_enabled"
	}
	return "cache_disabled"
}

func mapStateName(populated bool) string {
	if populated {
		return "populated"
	}
	return "nil"
}

type viewLayerTestConn struct {
	cacheEnabled bool
}

func (viewLayerTestConn) Close() error                                               { return nil }
func (viewLayerTestConn) IdentityData() login.IdentityData                           { return login.IdentityData{} }
func (viewLayerTestConn) ClientData() login.ClientData                               { return login.ClientData{} }
func (c viewLayerTestConn) ClientCacheEnabled() bool                                 { return c.cacheEnabled }
func (viewLayerTestConn) ChunkRadius() int                                           { return 1 }
func (viewLayerTestConn) Latency() time.Duration                                     { return 0 }
func (viewLayerTestConn) Flush() error                                               { return nil }
func (viewLayerTestConn) RemoteAddr() net.Addr                                       { return viewLayerTestAddr("test") }
func (viewLayerTestConn) ReadPacket() (packet.Packet, error)                         { return nil, net.ErrClosed }
func (viewLayerTestConn) WritePacket(packet.Packet) error                            { return nil }
func (viewLayerTestConn) StartGameContext(context.Context, minecraft.GameData) error { return nil }

type viewLayerTestAddr string

func (a viewLayerTestAddr) Network() string { return string(a) }
func (a viewLayerTestAddr) String() string  { return string(a) }

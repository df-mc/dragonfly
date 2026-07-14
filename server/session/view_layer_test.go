package session

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/stretchr/testify/require"
)

func TestPrivateBlockAdvertisesFirstEmptySubChunk(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	s := &Session{
		closeBackground: make(chan struct{}),
		packets:         make(chan packet.Packet, 16),
		conn:            viewLayerTestConn{},
		br:              world.DefaultBlockRegistry,
	}
	s.viewLayer = world.NewViewLayer(s)
	s.viewWorld.Store(w)
	s.chunkLoader = world.NewLoader(1, w, s)

	var advertised bool
	err := w.Do(func(tx *world.Tx) {
		s.chunkLoader.Load(tx, 1)
		for len(s.packets) > 0 {
			<-s.packets
		}

		pos := cube.Pos{0, int(w.Range()[0]), 0}
		s.viewLayer.ViewBlock(w, pos, block.Stone{})

		for len(s.packets) > 0 {
			if _, ok := (<-s.packets).(*packet.LevelChunk); ok {
				advertised = true
			}
		}
		s.chunkLoader.Close(tx)
	}).Wait(context.Background())
	require.NoError(t, err)
	require.True(t, advertised)
}

type viewLayerTestConn struct{}

func (viewLayerTestConn) Close() error                                               { return nil }
func (viewLayerTestConn) IdentityData() login.IdentityData                           { return login.IdentityData{} }
func (viewLayerTestConn) ClientData() login.ClientData                               { return login.ClientData{} }
func (viewLayerTestConn) ClientCacheEnabled() bool                                   { return false }
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

package player

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/stretchr/testify/require"
)

func TestBreakViewedBlockRemovesPrivateOverrideWithoutMutatingWorld(t *testing.T) {
	withViewLayerTestPlayer(t, func(p *Player, tx *world.Tx) {
		pos := cube.Pos{0, 64, 0}
		tx.SetBlock(pos, block.Dirt{}, nil)
		p.ViewBlock(pos, block.Stone{})

		p.BreakViewedBlock(pos)

		_, ok := p.ViewLayer().Block(pos)
		require.False(t, ok, "expected private override to be removed")
		require.IsType(t, block.Dirt{}, tx.Block(pos))
	})
}

func TestBreakBlockIgnoresPrivateOverrideAndBreaksPublicBlock(t *testing.T) {
	withViewLayerTestPlayer(t, func(p *Player, tx *world.Tx) {
		pos := cube.Pos{0, 64, 0}
		tx.SetBlock(pos, block.Dirt{}, nil)
		p.ViewBlock(pos, block.Stone{})

		p.BreakBlock(pos)

		_, ok := p.ViewLayer().Block(pos)
		require.True(t, ok, "expected private override to remain")
		require.IsType(t, block.Air{}, tx.Block(pos))
	})
}

func TestFinishBreakingUsesStartedBreakMode(t *testing.T) {
	withViewLayerTestPlayer(t, func(p *Player, tx *world.Tx) {
		pos := cube.Pos{0, 64, 0}
		tx.SetBlock(pos, block.Dirt{}, nil)

		p.StartBreaking(pos, cube.FaceUp)
		p.ViewBlock(pos, block.Stone{})
		p.FinishBreaking()

		_, ok := p.ViewLayer().Block(pos)
		require.True(t, ok, "expected private override added after StartBreaking to remain")
		require.IsType(t, block.Air{}, tx.Block(pos))
	})
}

func TestUseItemOnPrivateBlockDoesNotMutatePublicWorld(t *testing.T) {
	withViewLayerTestPlayer(t, func(p *Player, tx *world.Tx) {
		pos := cube.Pos{0, 64, 0}
		tx.SetBlock(pos, block.Stone{}, nil)
		p.ViewBlock(pos, block.Lever{Facing: cube.FaceUp, Direction: cube.North})

		p.UseItemOnBlock(pos, cube.FaceUp, mgl64.Vec3{})

		require.IsType(t, block.Stone{}, tx.Block(pos))
		_, ok := p.ViewLayer().Block(pos)
		require.True(t, ok, "expected private override to remain")
	})
}

func withViewLayerTestPlayer(t *testing.T, f func(*Player, *world.Tx)) {
	t.Helper()

	s := session.Config{MaxChunkRadius: 1}.New(fakeConn{})
	w := world.New()

	<-w.Exec(func(worldTx *world.Tx) {
		data := &world.EntityData{}
		conf := Config{
			Session:  s,
			GameMode: world.GameModeCreative,
			Position: mgl64.Vec3{0.5, 64, 0.5},
		}
		conf.Apply(data)
		f(&Player{
			tx:         worldTx,
			handle:     world.NewEntity(Type, conf),
			data:       data,
			playerData: data.Data.(*playerData),
		}, worldTx)
	})
}

type fakeConn struct{}

func (fakeConn) Close() error                                               { return nil }
func (fakeConn) IdentityData() login.IdentityData                           { return login.IdentityData{DisplayName: "test"} }
func (fakeConn) ClientData() login.ClientData                               { return login.ClientData{} }
func (fakeConn) ClientCacheEnabled() bool                                   { return false }
func (fakeConn) ChunkRadius() int                                           { return 1 }
func (fakeConn) Latency() time.Duration                                     { return 0 }
func (fakeConn) Flush() error                                               { return nil }
func (fakeConn) RemoteAddr() net.Addr                                       { return fakeAddr("test") }
func (fakeConn) ReadPacket() (packet.Packet, error)                         { return nil, net.ErrClosed }
func (fakeConn) WritePacket(packet.Packet) error                            { return nil }
func (fakeConn) StartGameContext(context.Context, minecraft.GameData) error { return nil }

type fakeAddr string

func (a fakeAddr) Network() string { return string(a) }
func (a fakeAddr) String() string  { return string(a) }

package player

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/stretchr/testify/require"
)

func TestViewLayerBlockInteractions(t *testing.T) {
	tests := []struct {
		name                 string
		publicBlock          world.Block
		privateBlock         world.Block
		action               func(*testing.T, *Player, *world.Tx, cube.Pos)
		expectedPublicBlock  world.Block
		expectedPrivateBlock world.Block
	}{
		{
			name:                "break visible block removes private override without mutating world",
			publicBlock:         block.Dirt{},
			privateBlock:        block.Stone{},
			action:              func(_ *testing.T, p *Player, _ *world.Tx, pos cube.Pos) { p.BreakVisibleBlock(pos) },
			expectedPublicBlock: block.Dirt{},
		},
		{
			name:                 "break block ignores private override and breaks public block",
			publicBlock:          block.Dirt{},
			privateBlock:         block.Stone{},
			action:               func(_ *testing.T, p *Player, _ *world.Tx, pos cube.Pos) { p.BreakBlock(pos) },
			expectedPublicBlock:  block.Air{},
			expectedPrivateBlock: block.Stone{},
		},
		{
			name:        "finish breaking uses started break mode",
			publicBlock: block.Dirt{},
			action: func(_ *testing.T, p *Player, _ *world.Tx, pos cube.Pos) {
				p.StartBreaking(pos, cube.FaceUp)
				p.ViewBlock(pos, block.Stone{})
				p.FinishBreaking()
			},
			expectedPublicBlock:  block.Air{},
			expectedPrivateBlock: block.Stone{},
		},
		{
			name:        "finish breaking re-reads public block before break",
			publicBlock: block.Dirt{},
			action: func(t *testing.T, p *Player, tx *world.Tx, pos cube.Pos) {
				h := &blockBreakTestHandler{}
				p.Handle(h)
				p.StartBreaking(pos, cube.FaceUp)
				tx.SetBlock(pos, nil, nil)
				p.FinishBreaking()
				require.False(t, h.blockBreakCalled)
			},
			expectedPublicBlock: block.Air{},
		},
		{
			name:         "use item on private block does not mutate public world",
			publicBlock:  block.Stone{},
			privateBlock: block.Lever{Facing: cube.FaceUp, Direction: cube.North},
			action: func(_ *testing.T, p *Player, _ *world.Tx, pos cube.Pos) {
				p.UseItemOnBlock(pos, cube.FaceUp, mgl64.Vec3{})
			},
			expectedPublicBlock:  block.Stone{},
			expectedPrivateBlock: block.Lever{},
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withViewLayerTestPlayer(t, func(p *Player, tx *world.Tx) {
				pos := cube.Pos{i, 64, 0}
				tx.SetBlock(pos, tt.publicBlock, nil)
				if tt.privateBlock != nil {
					p.ViewBlock(pos, tt.privateBlock)
				}

				tt.action(t, p, tx, pos)

				require.IsType(t, tt.expectedPublicBlock, tx.Block(pos))
				privateBlock, ok := p.ViewLayer().Block(tx.World(), pos)
				if tt.expectedPrivateBlock == nil {
					require.False(t, ok)
					return
				}
				require.True(t, ok)
				require.IsType(t, tt.expectedPrivateBlock, privateBlock)
			})
		})
	}
}

func TestHandleBlockBreakReceivesPrivateBreakMode(t *testing.T) {
	withViewLayerTestPlayer(t, func(p *Player, tx *world.Tx) {
		pos := cube.Pos{0, 64, 0}
		tx.SetBlock(pos, block.Dirt{}, nil)
		p.ViewBlock(pos, block.Stone{})

		h := &blockBreakTestHandler{}
		p.Handle(h)

		p.BreakVisibleBlock(pos)
		p.ViewBlock(pos, block.Stone{})
		p.BreakBlock(pos)

		require.Equal(t, []bool{true, false}, h.private)
	})
}

func TestContinueBreakingReReadsTargetBlock(t *testing.T) {
	tests := []struct {
		name         string
		private      bool
		changeBlock  func(*Player, *world.Tx, cube.Pos)
		expectedType world.Block
	}{
		{
			name:    "public",
			private: false,
			changeBlock: func(_ *Player, tx *world.Tx, pos cube.Pos) {
				tx.SetBlock(pos, block.Obsidian{}, nil)
			},
			expectedType: block.Obsidian{},
		},
		{
			name:    "private",
			private: true,
			changeBlock: func(p *Player, _ *world.Tx, pos cube.Pos) {
				p.ViewBlock(pos, block.Obsidian{})
			},
			expectedType: block.Obsidian{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withViewLayerTestPlayer(t, func(p *Player, tx *world.Tx) {
				p.gameMode = world.GameModeSurvival
				pos := cube.Pos{0, 64, 0}
				tx.SetBlock(pos, block.Dirt{}, nil)
				if tt.private {
					p.ViewBlock(pos, block.Dirt{})
				}
				p.StartBreaking(pos, cube.FaceUp)

				tt.changeBlock(p, tx, pos)
				p.ContinueBreaking(cube.FaceUp)

				require.Equal(t, p.breakTime(tt.expectedType), p.lastBreakDuration)
			})
		})
	}
}

func TestViewerViewsPublicBlock(t *testing.T) {
	w := world.New()
	defer w.Close()

	pos := cube.Pos{0, 64, 0}
	viewer := &blockBreakTestViewer{viewLayer: world.NewViewLayer(nil)}
	require.True(t, viewerViewsPublicBlock(viewer, w, pos))

	viewer.viewLayer.ViewBlock(w, pos, block.Stone{})
	require.False(t, viewerViewsPublicBlock(viewer, w, pos))
	require.True(t, viewerViewsPublicBlock(world.NopViewer{}, w, pos))
}

func withViewLayerTestPlayer(t *testing.T, f func(*Player, *world.Tx)) {
	t.Helper()

	s := session.Config{MaxChunkRadius: 1}.New(fakeConn{})
	w := world.New()
	defer func() {
		_ = w.Close()
		s.CloseConnection()
	}()

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

type blockBreakTestHandler struct {
	NopHandler
	blockBreakCalled bool
	private          []bool
}

func (h *blockBreakTestHandler) HandleBlockBreak(_ *Context, _ cube.Pos, private bool, _ *[]item.Stack, _ *int) {
	h.blockBreakCalled = true
	h.private = append(h.private, private)
}

type fakeConn struct{}

type blockBreakTestViewer struct {
	world.NopViewer
	viewLayer *world.ViewLayer
}

func (v *blockBreakTestViewer) ViewLayer() *world.ViewLayer {
	return v.viewLayer
}

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

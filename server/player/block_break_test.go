package player

import (
	"context"
	"net"
	"reflect"
	"slices"
	"sync"
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
				if h.blockBreakCalled {
					t.Fatal("block-break handler called after the public target disappeared")
				}
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

				if got := tx.Block(pos); reflect.TypeOf(got) != reflect.TypeOf(tt.expectedPublicBlock) {
					t.Fatalf("expected public block type %T, got %T", tt.expectedPublicBlock, got)
				}
				privateBlock, ok := p.ViewLayer().Block(tx.World(), pos)
				if tt.expectedPrivateBlock == nil {
					if ok {
						t.Fatalf("expected no private block, got %T", privateBlock)
					}
					return
				}
				if !ok || reflect.TypeOf(privateBlock) != reflect.TypeOf(tt.expectedPrivateBlock) {
					t.Fatalf("expected private block type %T, got %T (present %t)", tt.expectedPrivateBlock, privateBlock, ok)
				}
			})
		})
	}
}

func TestPrivateBlockUseCorrectsPredictedPlacement(t *testing.T) {
	var (
		packets []packet.Packet
		waitErr error
		clicked = cube.Pos{0, 64, 0}
		face    = cube.FaceUp
	)
	withSpawnedViewLayerTestPlayerConn(t, func(p *Player, tx *world.Tx, conn *fakeConn) {
		tx.SetBlock(clicked, block.Dirt{}, nil)
		p.ViewBlock(clicked, block.Stone{})
		p.SetHeldItems(item.NewStack(block.Dirt{}, 1), item.Stack{})

		p.session().SendMessage("before private block use")
		_, waitErr = conn.packetsUntilText("before private block use")
		if waitErr != nil {
			return
		}

		p.UseItemOnBlock(clicked, face, mgl64.Vec3{})
		p.session().SendMessage("after private block use")
		packets, waitErr = conn.packetsUntilText("after private block use")
	})
	if waitErr != nil {
		t.Fatal(waitErr)
	}

	updated := map[cube.Pos]*packet.UpdateBlock{}
	for _, pk := range packets {
		if update, ok := pk.(*packet.UpdateBlock); ok && update.Layer == 0 {
			pos := cube.Pos{int(update.Position[0]), int(update.Position[1]), int(update.Position[2])}
			updated[pos] = update
		}
	}
	if len(updated) != 2 {
		t.Fatalf("expected two corrected positions, got %v", updated)
	}
	for pos, want := range map[cube.Pos]uint32{
		clicked:            world.DefaultBlockRegistry.BlockRuntimeID(block.Stone{}),
		clicked.Side(face): world.DefaultBlockRegistry.AirRuntimeID(),
	} {
		update, ok := updated[pos]
		if !ok {
			t.Fatalf("expected correction at %v, got %v", pos, updated)
		}
		if update.NewBlockRuntimeID != want {
			t.Fatalf("correction at %v: expected runtime ID %d, got %d", pos, want, update.NewBlockRuntimeID)
		}
		if update.Flags != packet.BlockUpdateNetwork {
			t.Fatalf("correction at %v: expected network flag, got %d", pos, update.Flags)
		}
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

		if want := []bool{true, false}; !slices.Equal(h.private, want) {
			t.Fatalf("expected private modes %v, got %v", want, h.private)
		}
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
				if p.blockBreakTarget == nil {
					t.Fatal("expected a break target after starting")
				}
				if got := p.blockBreakTarget.mode == privateBlockView; got != tt.private {
					t.Fatalf("expected private mode %t after starting, got %t", tt.private, got)
				}

				tt.changeBlock(p, tx, pos)
				p.ContinueBreaking(cube.FaceUp)

				if p.blockBreakTarget == nil {
					t.Fatal("expected break target to survive continuation")
				}
				if got := p.blockBreakTarget.mode == privateBlockView; got != tt.private {
					t.Fatalf("expected private mode %t after continuing, got %t", tt.private, got)
				}
				if want := p.breakTime(tt.expectedType); p.lastBreakDuration != want {
					t.Fatalf("expected break duration %v, got %v", want, p.lastBreakDuration)
				}
				p.FinishBreaking()
				if p.blockBreakTarget != nil {
					t.Fatal("expected break target to clear after finishing")
				}
				if tt.private {
					if got := tx.Block(pos); reflect.TypeOf(got) != reflect.TypeOf(block.Dirt{}) {
						t.Fatalf("expected public dirt to remain, got %T", got)
					}
					_, ok := p.ViewLayer().Block(tx.World(), pos)
					if ok {
						t.Fatal("expected private override to be removed")
					}
					return
				}
				if got := tx.Block(pos); reflect.TypeOf(got) != reflect.TypeOf(block.Air{}) {
					t.Fatalf("expected public block to be broken, got %T", got)
				}
			})
		})
	}
}

func TestPublicBlockAudienceUsesAffectedBlockViewers(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	playerViewer := &blockBreakTestViewer{}
	blockViewer := &blockBreakTestViewer{}
	playerPos := mgl64.Vec3{15.5, 64, 0.5}
	blockPos := cube.Pos{16, 64, 0}

	err := w.Do(func(tx *world.Tx) {
		playerLoader := world.NewLoader(1, w, playerViewer)
		blockLoader := world.NewLoader(1, w, blockViewer)
		playerLoader.Move(tx, playerPos)
		blockLoader.Move(tx, blockPos.Vec3Centre())
		playerLoader.Load(tx, 1)
		blockLoader.Load(tx, 1)
		defer playerLoader.Close(tx)
		defer blockLoader.Close(tx)

		p := &Player{tx: tx, data: &world.EntityData{Pos: playerPos}, playerData: &playerData{}}
		publicBlockAudience{p: p}.ViewBlockAction(blockPos, nil)
	}).Wait(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(playerViewer.blockActions) != 0 {
		t.Fatalf("player-chunk viewer received block actions: %v", playerViewer.blockActions)
	}
	if want := []cube.Pos{blockPos}; !slices.Equal(blockViewer.blockActions, want) {
		t.Fatalf("expected affected-block viewer actions %v, got %v", want, blockViewer.blockActions)
	}
}

func TestPublicBlockAudienceSoundUsesWorldHandler(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	h := &cancellingSoundHandler{}
	w.Handle(h)
	played := false
	err := w.Do(func(tx *world.Tx) {
		p := &Player{tx: tx, data: &world.EntityData{}, playerData: &playerData{}}
		publicBlockAudience{p: p}.PlaySound(mgl64.Vec3{}, recordingSound{played: &played})
	}).Wait(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if h.calls != 1 {
		t.Fatalf("expected one sound-handler call, got %d", h.calls)
	}
	if played {
		t.Fatal("cancelled public sound mutated the world")
	}
}

func TestPrivateBreakSoundHonoursWorldHandlerCancellation(t *testing.T) {
	var (
		packets []packet.Packet
		waitErr error
	)
	w := world.New()
	h := &cancellingSoundHandler{}
	w.Handle(h)
	defer w.Close()

	withViewLayerTestPlayerConnInWorld(t, w, func(p *Player, tx *world.Tx, conn *fakeConn) {
		pos := cube.Pos{0, 64, 0}
		tx.SetBlock(pos, block.Dirt{}, nil)
		p.ViewBlock(pos, block.Stone{})
		p.StartBreaking(pos, cube.FaceUp)

		p.session().SendMessage("before private break sound")
		_, waitErr = conn.packetsUntilText("before private break sound")
		if waitErr != nil {
			return
		}
		for range 5 {
			p.ContinueBreaking(cube.FaceUp)
		}
		p.session().SendMessage("after private break sound")
		packets, waitErr = conn.packetsUntilText("after private break sound")
	})
	if waitErr != nil {
		t.Fatal(waitErr)
	}
	if h.calls != 1 {
		t.Fatalf("expected one sound-handler call, got %d", h.calls)
	}
	for _, pk := range packets {
		_, soundDelivered := pk.(*packet.LevelSoundEvent)
		if soundDelivered {
			t.Fatal("cancelled private break sound was delivered")
		}
	}
}

func TestPrivateBlockSoundDoesNotPlayInPublicWorld(t *testing.T) {
	var (
		packets []packet.Packet
		waitErr error
		played  bool
	)
	w := world.New()
	h := &recordingSoundHandler{}
	w.Handle(h)
	defer w.Close()

	pos := mgl64.Vec3{1, 2, 3}
	withViewLayerTestPlayerConnInWorld(t, w, func(p *Player, _ *world.Tx, conn *fakeConn) {
		privateBlockAudience{p: p}.PlaySound(pos, recordingSound{played: &played})
		p.session().SendMessage("after private sound")
		packets, waitErr = conn.packetsUntilText("after private sound")
	})
	if waitErr != nil {
		t.Fatal(waitErr)
	}
	if want := []mgl64.Vec3{pos}; !slices.Equal(h.positions, want) {
		t.Fatalf("expected sound-handler positions %v, got %v", want, h.positions)
	}
	if played {
		t.Fatal("private sound mutated the public world")
	}
	delivered := false
	for _, pk := range packets {
		if _, ok := pk.(*packet.LevelSoundEvent); ok {
			delivered = true
		}
	}
	if !delivered {
		t.Fatal("private sound was not delivered to its session")
	}
}

func TestFireExtinguishSoundUsesFirePosition(t *testing.T) {
	w := world.New()
	h := &recordingSoundHandler{}
	w.Handle(h)
	defer w.Close()

	clicked, face := cube.Pos{0, 64, 0}, cube.FaceUp
	withViewLayerTestPlayerConnInWorld(t, w, func(p *Player, tx *world.Tx, _ *fakeConn) {
		tx.SetBlock(clicked, block.Stone{}, nil)
		tx.SetBlock(clicked.Side(face), block.Fire{}, nil)
		p.StartBreaking(clicked, face)
	})
	if want := []mgl64.Vec3{clicked.Side(face).Vec3()}; !slices.Equal(h.positions, want) {
		t.Fatalf("expected fire sound at %v, got %v", want, h.positions)
	}
}

func withViewLayerTestPlayer(t *testing.T, f func(*Player, *world.Tx)) {
	t.Helper()
	withViewLayerTestPlayerConn(t, func(p *Player, tx *world.Tx, _ *fakeConn) { f(p, tx) })
}

func withViewLayerTestPlayerConn(t *testing.T, f func(*Player, *world.Tx, *fakeConn)) {
	t.Helper()
	w := world.New()
	defer w.Close()
	withViewLayerTestPlayerConnInWorld(t, w, f)
}

func withSpawnedViewLayerTestPlayerConn(t *testing.T, f func(*Player, *world.Tx, *fakeConn)) {
	t.Helper()

	w := world.New()
	defer w.Close()
	conn := newFakeConn()
	s := session.Config{
		MaxChunkRadius: 1,
		HandleStop:     func(*world.Tx, session.Controllable) {},
	}.New(conn)
	defer s.CloseConnection()

	err := w.Do(func(tx *world.Tx) {
		conf := Config{
			Session:  s,
			GameMode: world.GameModeCreative,
			Position: mgl64.Vec3{0.5, 64, 0.5},
		}
		data := &world.EntityData{}
		conf.Apply(data)
		handle := world.NewEntity(Type, conf)
		s.SetHandle(handle, conf.Skin)
		p := tx.AddEntity(handle).(*Player)
		s.Spawn(p, tx)
		f(p, tx, conn)
		s.Close(tx, p)
	}).Wait(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func withViewLayerTestPlayerConnInWorld(t *testing.T, w *world.World, f func(*Player, *world.Tx, *fakeConn)) {
	t.Helper()

	conn := newFakeConn()
	s := session.Config{MaxChunkRadius: 1}.New(conn)
	defer func() {
		s.CloseConnection()
	}()

	err := w.Do(func(worldTx *world.Tx) {
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
		}, worldTx, conn)
	}).Wait(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

type blockBreakTestHandler struct {
	NopHandler
	blockBreakCalled bool
	private          []bool
}

type cancellingSoundHandler struct {
	world.NopHandler
	calls int
}

func (h *cancellingSoundHandler) HandleSound(ctx *world.Context, _ world.Sound, _ mgl64.Vec3) {
	h.calls++
	ctx.Cancel()
}

type recordingSoundHandler struct {
	world.NopHandler
	positions []mgl64.Vec3
}

func (h *recordingSoundHandler) HandleSound(_ *world.Context, _ world.Sound, pos mgl64.Vec3) {
	h.positions = append(h.positions, pos)
}

type recordingSound struct {
	played *bool
}

func (s recordingSound) Play(*world.World, mgl64.Vec3) {
	*s.played = true
}

func (h *blockBreakTestHandler) HandleBlockBreak(_ *Context, _ cube.Pos, private bool, _ *[]item.Stack, _ *int) {
	h.blockBreakCalled = true
	h.private = append(h.private, private)
}

type fakeConn struct {
	closeOnce sync.Once
	closed    chan struct{}
	packets   chan packet.Packet
}

func newFakeConn() *fakeConn {
	return &fakeConn{closed: make(chan struct{}), packets: make(chan packet.Packet, 1024)}
}

func (c *fakeConn) packetsUntilText(message string) ([]packet.Packet, error) {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	var packets []packet.Packet
	for {
		select {
		case pk := <-c.packets:
			if text, ok := pk.(*packet.Text); ok && text.Message == message {
				return packets, nil
			}
			packets = append(packets, pk)
		case <-timer.C:
			return nil, context.DeadlineExceeded
		}
	}
}

type blockBreakTestViewer struct {
	world.NopViewer
	blockActions []cube.Pos
}

func (v *blockBreakTestViewer) ViewBlockAction(pos cube.Pos, _ world.BlockAction) {
	v.blockActions = append(v.blockActions, pos)
}

func (c *fakeConn) Close() error {
	c.closeOnce.Do(func() { close(c.closed) })
	return nil
}
func (*fakeConn) IdentityData() login.IdentityData { return login.IdentityData{DisplayName: "test"} }
func (*fakeConn) ClientData() login.ClientData     { return login.ClientData{} }
func (*fakeConn) ClientCacheEnabled() bool         { return false }
func (*fakeConn) ChunkRadius() int                 { return 1 }
func (*fakeConn) Latency() time.Duration           { return 0 }
func (*fakeConn) Flush() error                     { return nil }
func (*fakeConn) RemoteAddr() net.Addr             { return fakeAddr("test") }
func (c *fakeConn) ReadPacket() (packet.Packet, error) {
	<-c.closed
	return nil, net.ErrClosed
}
func (c *fakeConn) WritePacket(pk packet.Packet) error {
	select {
	case <-c.closed:
		return net.ErrClosed
	case c.packets <- pk:
		return nil
	}
}
func (*fakeConn) StartGameContext(context.Context, minecraft.GameData) error { return nil }

type fakeAddr string

func (a fakeAddr) Network() string { return string(a) }
func (a fakeAddr) String() string  { return string(a) }

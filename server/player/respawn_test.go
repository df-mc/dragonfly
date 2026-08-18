package player

import (
	"context"
	"io"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestEndExitUsesRespawnAnchorWithoutConsumingCharge(t *testing.T) {
	provider := &respawnTestProvider{spawns: make(map[uuid.UUID]world.PlayerSpawn)}
	var overworld, nether, end *world.World
	overworld = world.Config{Provider: provider}.New()
	nether = world.Config{Provider: provider, Dim: world.Nether}.New()
	end = world.Config{Provider: provider, Dim: world.End, PortalDestination: func(dim world.Dimension) *world.World {
		if dim == world.End {
			return overworld
		}
		return nil
	}}.New()
	t.Cleanup(func() {
		_ = end.Close()
		_ = nether.Close()
		_ = overworld.Close()
	})

	id := uuid.New()
	anchorPos := cube.Pos{4, 64, 4}
	spawnPos := anchorPos.Add(cube.Pos{0, 0, -1})
	respawnTestDo(t, nether, func(tx *world.Tx) {
		tx.SetBlock(anchorPos, block.RespawnAnchor{Charges: 2}, nil)
		tx.SetBlock(spawnPos.Side(cube.FaceDown), block.Stone{}, nil)
	})
	nether.SetPlayerSpawn(id, anchorPos)

	dimensions := func(dim world.Dimension) *world.World {
		return map[world.Dimension]*world.World{
			world.Overworld: overworld,
			world.Nether:    nether,
			world.End:       end,
		}[dim]
	}
	handle := world.EntitySpawnOpts{ID: id}.New(Type, Config{
		UUID: id, Health: 20, MaxHealth: 20, WorldByDimension: dimensions,
	})
	respawnTestDo(t, end, func(tx *world.Tx) {
		tx.AddEntity(handle).(*Player).TravelThroughPortal(tx, world.End)
	})
	waitForRespawnTestPlayer(t, handle, nether)
	respawnTestDo(t, nether, func(tx *world.Tx) {
		entity, ok := handle.Entity(tx)
		if !ok {
			t.Fatal("player was not present in the Nether")
		}
		if got := entity.Position(); !got.ApproxEqual(spawnPos.Vec3Middle()) {
			t.Fatalf("End return position = %v, want %v", got, spawnPos.Vec3Middle())
		}
		if got := tx.Block(anchorPos).(block.RespawnAnchor).Charges; got != 2 {
			t.Fatalf("anchor charges after End return = %d, want 2", got)
		}
		tx.RemoveEntity(entity)
	})
	t.Cleanup(func() { _ = handle.Close() })
}

func TestRespawnDoesNotConsumeAnchorWhenArrivalPanics(t *testing.T) {
	provider := &respawnTestProvider{spawns: make(map[uuid.UUID]world.PlayerSpawn)}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var overworld, nether, end *world.World
	overworld = world.Config{Log: logger, Provider: provider}.New()
	nether = world.Config{Log: logger, Provider: provider, Dim: world.Nether}.New()
	end = world.Config{Log: logger, Provider: provider, Dim: world.End, PortalDestination: func(dim world.Dimension) *world.World {
		if dim == world.End {
			return overworld
		}
		return nil
	}}.New()
	t.Cleanup(func() {
		_ = end.Close()
		_ = nether.Close()
		_ = overworld.Close()
	})

	id := uuid.New()
	anchorPos := cube.Pos{4, 64, 4}
	spawnPos := anchorPos.Add(cube.Pos{0, 0, -1})
	respawnTestDo(t, nether, func(tx *world.Tx) {
		tx.SetBlock(anchorPos, block.RespawnAnchor{Charges: 2}, nil)
		tx.SetBlock(spawnPos.Side(cube.FaceDown), block.Stone{}, nil)
	})
	nether.SetPlayerSpawn(id, anchorPos)
	nether.Handle(panicEntitySpawnHandler{})

	s := session.Config{MaxChunkRadius: 2, HandleStop: func(*world.Tx, session.Controllable) {}}.New(respawnTestConn{})
	conf := Config{
		UUID: id, Health: 0, MaxHealth: 20, Session: s,
		WorldByDimension: func(dim world.Dimension) *world.World {
			return map[world.Dimension]*world.World{
				world.Overworld: overworld,
				world.Nether:    nether,
				world.End:       end,
			}[dim]
		},
	}
	handle := world.EntitySpawnOpts{ID: id}.New(Type, conf)
	s.SetHandle(handle, skin.Skin{})
	respawnTestDo(t, end, func(tx *world.Tx) {
		tx.AddEntity(handle).(*Player).Respawn()
	})
	waitForRespawnTestPlayer(t, handle, nether)
	respawnTestDo(t, nether, func(tx *world.Tx) {
		if got := tx.Block(anchorPos).(block.RespawnAnchor).Charges; got != 2 {
			t.Fatalf("anchor charges after failed arrival = %d, want 2", got)
		}
		if entity, ok := handle.Entity(tx); ok {
			tx.RemoveEntity(entity)
		}
	})
	t.Cleanup(func() {
		_ = handle.Close()
		s.CloseConnection()
	})
}

type panicEntitySpawnHandler struct{ world.NopHandler }

func (panicEntitySpawnHandler) HandleEntitySpawn(*world.Tx, world.Entity) {
	panic("test destination insertion failure")
}

func respawnTestDo(t *testing.T, w *world.World, f func(*world.Tx)) {
	t.Helper()
	if err := w.Do(f).Wait(context.Background()); err != nil {
		t.Fatalf("world task failed: %v", err)
	}
}

func waitForRespawnTestPlayer(t *testing.T, handle *world.EntityHandle, target *world.World) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		present := false
		respawnTestDo(t, target, func(tx *world.Tx) {
			_, present = handle.Entity(tx)
		})
		if present {
			return
		}
		time.Sleep(time.Millisecond * 10)
	}
	t.Fatal("timed out waiting for player to enter target world")
}

type respawnTestProvider struct {
	world.NopProvider
	mu     sync.Mutex
	spawns map[uuid.UUID]world.PlayerSpawn
}

func (p *respawnTestProvider) LoadPlayerSpawn(id uuid.UUID) (world.PlayerSpawn, bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	spawn, ok := p.spawns[id]
	return spawn, ok, nil
}

func (p *respawnTestProvider) SavePlayerSpawn(id uuid.UUID, spawn world.PlayerSpawn) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.spawns[id] = spawn
	return nil
}

type respawnTestConn struct{}

func (respawnTestConn) Close() error                                               { return nil }
func (respawnTestConn) IdentityData() login.IdentityData                           { return login.IdentityData{} }
func (respawnTestConn) ClientData() login.ClientData                               { return login.ClientData{} }
func (respawnTestConn) ClientCacheEnabled() bool                                   { return false }
func (respawnTestConn) ChunkRadius() int                                           { return 2 }
func (respawnTestConn) Latency() time.Duration                                     { return 0 }
func (respawnTestConn) Flush() error                                               { return nil }
func (respawnTestConn) RemoteAddr() net.Addr                                       { return &net.UDPAddr{} }
func (respawnTestConn) ReadPacket() (packet.Packet, error)                         { return nil, io.EOF }
func (respawnTestConn) WritePacket(packet.Packet) error                            { return nil }
func (respawnTestConn) StartGameContext(context.Context, minecraft.GameData) error { return nil }

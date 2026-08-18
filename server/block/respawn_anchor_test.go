package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
)

func TestRespawnAnchorExplosionRule(t *testing.T) {
	settings := &world.Settings{RespawnBlocksExplode: false}
	w := world.Config{Provider: world.NopProvider{Set: settings}, Synchronous: true}.New()
	t.Cleanup(func() { _ = w.Close() })
	h := &respawnAnchorExplosionHandler{}
	w.Handle(h)

	pos := cube.Pos{0, 64, 0}
	u := respawnAnchorUser{id: uuid.New()}
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(pos, RespawnAnchor{Charges: 1}, nil)
		if !(RespawnAnchor{Charges: 1}).Activate(pos, cube.FaceUp, tx, u, &item.UseContext{}) {
			t.Fatal("Activate() = false, want true")
		}
		if _, ok := tx.Block(pos).(Air); !ok {
			t.Fatalf("block after disabled explosion = %T, want Air", tx.Block(pos))
		}
		if h.explosions != 0 {
			t.Fatalf("explosions with RespawnBlocksExplode disabled = %d, want 0", h.explosions)
		}

		settings.Lock()
		settings.RespawnBlocksExplode = true
		settings.Unlock()
		tx.SetBlock(pos, RespawnAnchor{Charges: 1}, nil)
		(RespawnAnchor{Charges: 1}).Activate(pos, cube.FaceUp, tx, u, &item.UseContext{})
		if h.explosions != 1 {
			t.Fatalf("explosions with RespawnBlocksExplode enabled = %d, want 1", h.explosions)
		}
	})
}

func TestRespawnAnchorWaterSuppressesImpact(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = w.Close() })
	h := &respawnAnchorExplosionHandler{}
	w.Handle(h)

	pos := cube.Pos{0, 64, 0}
	runWorld(w, func(tx *world.Tx) {
		tx.SetLiquid(pos.Side(cube.FaceEast), Lava{Depth: 8, Still: true})
		if respawnAnchorTouchesWater(pos, tx) {
			t.Fatal("respawnAnchorTouchesWater() = true with only lava touching the anchor")
		}
		tx.SetLiquid(pos.Side(cube.FaceEast), Water{Depth: 8, Still: true})
		if !respawnAnchorTouchesWater(pos, tx) {
			t.Fatal("respawnAnchorTouchesWater() = false with water touching the anchor")
		}
		(RespawnAnchor{Charges: 1}).explode(pos, tx)
		if h.affectedBlocks != 0 {
			t.Fatalf("water-suppressed block impact = %d, want 0", h.affectedBlocks)
		}
		if h.explosions != 1 || h.sounds != 1 {
			t.Fatalf("water-suppressed explosion emitted %d events and %d sounds, want 1 and 1", h.explosions, h.sounds)
		}
	})
}

func TestRespawnAnchorChargeConsumedOnlyByRespawn(t *testing.T) {
	w := world.Config{Dim: world.Nether, Synchronous: true}.New()
	t.Cleanup(func() { _ = w.Close() })

	anchorPos := cube.Pos{0, 64, 0}
	spawnPos := anchorPos.Add(respawnAnchorSpawnOffsets[0])
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(anchorPos, RespawnAnchor{Charges: 2}, nil)
		tx.SetBlock(spawnPos.Side(cube.FaceDown), Stone{}, nil)

		anchor := tx.Block(anchorPos).(RespawnAnchor)
		if spawn, ok := anchor.SafeSpawn(anchorPos, tx); !ok || spawn != spawnPos {
			t.Fatalf("SafeSpawn() = %v, %t, want %v, true", spawn, ok, spawnPos)
		}
		if got := tx.Block(anchorPos).(RespawnAnchor).Charges; got != 2 {
			t.Fatalf("charges after SafeSpawn() = %d, want 2", got)
		}

		anchor.UseRespawn(anchorPos, tx)
		if got := tx.Block(anchorPos).(RespawnAnchor).Charges; got != 1 {
			t.Fatalf("charges after UseRespawn() = %d, want 1", got)
		}
	})
}

type respawnAnchorExplosionHandler struct {
	world.NopHandler
	explosions     int
	affectedBlocks int
	sounds         int
}

func (h *respawnAnchorExplosionHandler) HandleSound(*world.Context, world.Sound, mgl64.Vec3) {
	h.sounds++
}

func (h *respawnAnchorExplosionHandler) HandleExplosion(_ *world.Context, _ world.ExplosionSource, entities *[]world.Entity, blocks *[]cube.Pos, _ *float64, _ *bool) {
	h.explosions++
	h.affectedBlocks = len(*blocks)
}

type respawnAnchorUser struct {
	id uuid.UUID
}

func (respawnAnchorUser) Close() error                        { return nil }
func (respawnAnchorUser) H() *world.EntityHandle              { return nil }
func (respawnAnchorUser) Position() mgl64.Vec3                { return mgl64.Vec3{} }
func (respawnAnchorUser) Rotation() cube.Rotation             { return cube.Rotation{} }
func (respawnAnchorUser) HeldItems() (item.Stack, item.Stack) { return item.Stack{}, item.Stack{} }
func (respawnAnchorUser) SetHeldItems(item.Stack, item.Stack) {}
func (respawnAnchorUser) UsingItem() bool                     { return false }
func (respawnAnchorUser) ReleaseItem()                        {}
func (respawnAnchorUser) UseItem()                            {}
func (u respawnAnchorUser) UUID() uuid.UUID                   { return u.id }
func (respawnAnchorUser) Messaget(chat.Translation, ...any)   {}

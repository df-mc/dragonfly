package entity

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	_ "github.com/df-mc/dragonfly/server/world/biome"
	"github.com/go-gl/mathgl/mgl64"
)

// testAnglerState holds the state of a testAngler, persisted in its EntityData.
type testAnglerState struct {
	held item.Stack
	hook *world.EntityHandle
}

func (s *testAnglerState) Apply(data *world.EntityData) { data.Data = s }

type testAnglerType struct{}

func (testAnglerType) Open(_ *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &testAngler{handle: handle, data: data}
}
func (testAnglerType) EncodeEntity() string                            { return "dragonfly:test_angler" }
func (testAnglerType) BBox(world.Entity) cube.BBox                     { return cube.Box(-0.3, 0, -0.3, 0.3, 1.8, 0.3) }
func (testAnglerType) DecodeNBT(_ map[string]any, _ *world.EntityData) {}
func (testAnglerType) EncodeNBT(*world.EntityData) map[string]any      { return nil }

// testAngler is a minimal entity that is able to hold a fishing rod and cast a fishing hook.
type testAngler struct {
	handle *world.EntityHandle
	data   *world.EntityData
}

func (a *testAngler) H() *world.EntityHandle  { return a.handle }
func (a *testAngler) Close() error            { return nil }
func (a *testAngler) Position() mgl64.Vec3    { return a.data.Pos }
func (a *testAngler) Rotation() cube.Rotation { return a.data.Rot }
func (a *testAngler) state() *testAnglerState { return a.data.Data.(*testAnglerState) }
func (a *testAngler) FishingHook() *world.EntityHandle {
	return a.state().hook
}
func (a *testAngler) SetFishingHook(h *world.EntityHandle) { a.state().hook = h }
func (a *testAngler) HeldItems() (item.Stack, item.Stack) {
	return a.state().held, item.Stack{}
}

// fillWater fills a square pool of water with the given radius around the position passed, with the
// surface at the Y of the position.
func fillWater(tx *world.Tx, centre cube.Pos, radius, depth int) {
	for x := -radius; x <= radius; x++ {
		for z := -radius; z <= radius; z++ {
			for y := 0; y < depth; y++ {
				tx.SetBlock(centre.Add(cube.Pos{x, -y, z}), block.Water{Depth: 8, Still: true}, nil)
			}
			tx.SetBlock(centre.Add(cube.Pos{x, -depth, z}), block.Stone{}, nil)
		}
	}
}

// castHook spawns an angler holding a fishing rod at pos and casts a fishing hook with the velocity passed.
func castHook(tx *world.Tx, pos, vel mgl64.Vec3) (*testAngler, *Ent) {
	angler := tx.AddEntity(world.EntitySpawnOpts{Position: pos}.New(testAnglerType{}, &testAnglerState{
		held: item.NewStack(item.FishingRod{}, 1),
	})).(*testAngler)
	hook := tx.AddEntity(NewFishingHook(world.EntitySpawnOpts{Position: pos.Add(mgl64.Vec3{0, 1.62}), Velocity: vel}, angler, 0, 0))
	angler.SetFishingHook(hook.H())
	return angler, hook.(*Ent)
}

func TestFishingHookCatchesFish(t *testing.T) {
	w := world.Config{Entities: world.EntityRegistryConfig{Item: conf.Item}.New([]world.EntityType{FishingHookType, ItemType, ExperienceOrbType, testAnglerType{}})}.New()
	t.Cleanup(func() { _ = w.Close() })

	mustDo(t, w, func(tx *world.Tx) {
		surface := cube.Pos{0, 60, 5}
		fillWater(tx, surface, 4, 3)

		_, hook := castHook(tx, mgl64.Vec3{0.5, 61, 0.5}, mgl64.Vec3{0, 0.2, 0.6})
		b := hook.Behaviour().(*FishingHookBehaviour)

		bit := false
		for i := int64(0); i < 2000; i++ {
			hook.Tick(tx, i)
			if _, ok := hook.H().Entity(tx); !ok {
				t.Fatalf("hook was removed after %v ticks", i)
			}
			if b.state == fishingHookBobbing && i > 100 {
				// The hook should float close to the water surface.
				if y := hook.Position().Y(); y < 59.5 || y > 61.5 {
					t.Fatalf("bobbing hook at unexpected height %v", y)
				}
			}
			if b.Biting() {
				bit = true
				break
			}
		}
		if b.state != fishingHookBobbing {
			t.Fatalf("expected hook to be bobbing in water, got state %v at %v", b.state, hook.Position())
		}
		if !bit {
			t.Fatal("no fish bit the hook within 2000 ticks")
		}

		if damage := b.Reel(hook, tx); damage != 1 {
			t.Fatalf("expected reeling in a fish to deal 1 damage, got %v", damage)
		}
		items := 0
		for e := range tx.Entities() {
			if e.H().Type() == ItemType {
				items++
			}
		}
		if items != 1 {
			t.Fatalf("expected 1 caught item, got %v", items)
		}
		if _, ok := hook.H().Entity(tx); ok {
			t.Fatal("hook was not removed after reeling in")
		}
	})
}

func TestFishingHookStopsWithoutRod(t *testing.T) {
	w := world.Config{Entities: world.EntityRegistryConfig{}.New([]world.EntityType{FishingHookType, testAnglerType{}})}.New()
	t.Cleanup(func() { _ = w.Close() })

	mustDo(t, w, func(tx *world.Tx) {
		angler, hook := castHook(tx, mgl64.Vec3{0.5, 61, 0.5}, mgl64.Vec3{0, 0.2, 0.6})
		hook.Tick(tx, 0)
		if _, ok := hook.H().Entity(tx); !ok {
			t.Fatal("hook was removed while the angler held a fishing rod")
		}
		angler.state().held = item.NewStack(item.Stick{}, 1)
		hook.Tick(tx, 1)
		if _, ok := hook.H().Entity(tx); ok {
			t.Fatal("hook was not removed after the angler stopped holding a fishing rod")
		}
		if angler.FishingHook() != nil {
			t.Fatal("hook was not cleared from the angler")
		}
	})
}

func TestFishingHookOnGround(t *testing.T) {
	w := world.Config{Entities: world.EntityRegistryConfig{}.New([]world.EntityType{FishingHookType, testAnglerType{}})}.New()
	t.Cleanup(func() { _ = w.Close() })

	mustDo(t, w, func(tx *world.Tx) {
		for x := -3; x <= 3; x++ {
			for z := -3; z <= 10; z++ {
				tx.SetBlock(cube.Pos{x, 59, z}, block.Stone{}, nil)
			}
		}
		_, hook := castHook(tx, mgl64.Vec3{0.5, 60, 0.5}, mgl64.Vec3{0, 0.2, 0.6})
		for i := int64(0); i < 100; i++ {
			hook.Tick(tx, i)
		}
		b := hook.Behaviour().(*FishingHookBehaviour)
		if y := hook.Position().Y(); y < 59.99 || y > 60.01 {
			t.Fatalf("expected hook to rest on the ground at Y 60, got %v", y)
		}
		if damage := b.Reel(hook, tx); damage != 2 {
			t.Fatalf("expected reeling in a hook on the ground to deal 2 damage, got %v", damage)
		}
	})
}

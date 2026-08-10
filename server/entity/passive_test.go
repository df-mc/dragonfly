package entity

import (
	"math"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// testWorld returns a synchronous World that is closed when the test ends.
func testWorld(t *testing.T) *world.World {
	t.Helper()
	w := world.Config{Synchronous: true, Entities: DefaultRegistry}.New()
	t.Cleanup(func() { _ = w.Close() })
	return w
}

// TestFallingBlockFallDistance verifies that a falling block tracks how far it has fallen, which is what decides the
// damage it deals when it lands. A sum of velocity changes telescopes to the final velocity rather than the distance
// travelled, and is cancelled by the tick the block lands on.
func TestFallingBlockFallDistance(t *testing.T) {
	w := testWorld(t)
	const start, floor = 60.0, 1
	mustDo(t, w, func(tx *world.Tx) {
		tx.SetBlock(cube.Pos{0, floor, 0}, block.Stone{}, nil)
	})

	anvil := NewFallingBlock(world.EntitySpawnOpts{Position: mgl64.Vec3{0.5, start, 0.5}}, block.Anvil{})
	mustDo(t, w, func(tx *world.Tx) { tx.AddEntity(anvil) })

	var maxDist, landingDist float64
	landed := false
	for range 400 {
		w.AdvanceTick()
		mustDo(t, w, func(tx *world.Tx) {
			e, ok := anvil.Entity(tx)
			if !ok {
				return
			}
			b := e.(*Ent).Behaviour().(*FallingBlockBehaviour)
			maxDist = max(maxDist, b.passive.fallDistance)
			if b.passive.close && !landed {
				// solidify() ran this tick; this is exactly the value
				// damageEntities() used.
				landed, landingDist = true, b.passive.fallDistance
			}
		})
	}
	if !landed {
		t.Fatal("precondition failed: falling block never landed")
	}

	want := start - float64(floor+1)
	// dist = ceil(fallDistance - 1); anvil Damage() = (2, 40).
	dist := math.Ceil(landingDist - 1)
	gotDmg := 0.0
	if dist > 0 {
		gotDmg = math.Min(math.Floor(dist*2), 40)
	}
	wantDmg := math.Min(math.Floor(math.Ceil(want-1)*2), 40)
	t.Logf("fell %v blocks; max fallDistance during the fall = %v, fallDistance when solidify() ran = %v",
		want, maxDist, landingDist)
	if landingDist < want-2 {
		t.Fatalf("falling block fell %v blocks but fallDistance was %v when it landed: "+
			"dist = ceil(%v-1) = %v, so a falling anvil deals %v damage instead of %v "+
			"(and never breaks, since the break roll is skipped when dmg == 0). "+
			"PassiveBehaviour.Tick accumulates the velocity delta m.dvel[1] instead of the position delta m.dpos[1], "+
			"which telescopes to the final velocity and is cancelled out entirely by the landing tick",
			want, landingDist, landingDist, dist, gotDmg, wantDmg)
	}
}

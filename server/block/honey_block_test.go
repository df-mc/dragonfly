package block_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// honeyPos is the position the tests below place their honey block at.
var honeyPos = cube.Pos{0, 60, 0}

// fallOntoHoney runs f with a Player that fell from y=70 in the column at x and came to rest at restY,
// in a world holding a honey block at honeyPos.
func fallOntoHoney(t *testing.T, x, restY float64, f func(tx *world.Tx, p *player.Player)) {
	t.Helper()

	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	pos := mgl64.Vec3{x, 70, 0.5}
	handle := world.EntitySpawnOpts{Position: pos}.New(player.Type, player.Config{Position: pos})
	w.Do(func(tx *world.Tx) {
		tx.SetBlock(honeyPos, block.HoneyBlock{}, nil)
		tx.AddEntity(handle)

		e, ok := handle.Entity(tx)
		if !ok {
			t.Fatal("player was not added to the world")
		}
		p := e.(*player.Player)
		for p.Position()[1] > restY {
			p.Move(mgl64.Vec3{0, -min(2, p.Position()[1]-restY), 0}, 0, 0)
		}
		if p.FallDistance() == 0 {
			t.Fatal("player did not accumulate a fall distance to test with")
		}
		f(tx, p)
	})
}

// onGround reports whether the server considers p to be standing on a block, replicating the check
// Player.checkOnGround makes with a zero movement delta.
func onGround(tx *world.Tx, p *player.Player) bool {
	box := player.Type.BBox(p).Translate(p.Position()).Extend(mgl64.Vec3{0, -0.05})
	grown := box.Grow(1)
	epsilon := mgl64.Vec3{mgl64.Epsilon, mgl64.Epsilon, mgl64.Epsilon}
	low, high := cube.PosFromVec3(grown.Min().Add(epsilon)), cube.PosFromVec3(grown.Max().Sub(epsilon))

	for x := low[0]; x <= high[0]; x++ {
		for z := low[2]; z <= high[2]; z++ {
			for y := low[1]; y < high[1]; y++ {
				pos := cube.Pos{x, y, z}
				for _, bb := range tx.Block(pos).Model().BBox(pos, tx) {
					if bb.Translate(pos.Vec3()).IntersectsWith(box) {
						return true
					}
				}
			}
		}
	}
	return false
}

// TestHoneyBlockStandingOnIsOnGround verifies that a player resting on a honey block is on the ground as
// far as the server is concerned. Fall damage is only dealt from the OnGround branch of
// Player.updateFallState, so a block shorter than the height the client rests players at would leave the
// server thinking the player is airborne and never damage it, at any fall height.
func TestHoneyBlockStandingOnIsOnGround(t *testing.T) {
	fallOntoHoney(t, 0.5, 61, func(tx *world.Tx, p *player.Player) {
		if !onGround(tx, p) {
			t.Error("expected a player resting on a honey block to be on the ground")
		}
	})
}

// TestHoneyBlockEntityLand verifies that a honey block reduces fall damage by 80%, using the example the
// wiki gives: a fall that would normally deal 10 damage deals 2 on a honey block. A fall is damaged for
// every block past the first three, so the fall dealing 10 damage normally is one of 13 blocks.
func TestHoneyBlockEntityLand(t *testing.T) {
	fallOntoHoney(t, 0.5, 61, func(tx *world.Tx, p *player.Player) {
		distance := 13.0
		block.HoneyBlock{}.EntityLand(honeyPos, tx, p, &distance)

		if want, got := 2.0, distance-3; !mgl64.FloatEqualThreshold(got, want, 1e-9) {
			t.Errorf("expected a fall damaging for 10 to damage for %v on honey, got %v", want, got)
		}
	})
}

// TestHoneyBlockLandingOnTopStillHurts verifies that a player that came to rest on top of a honey block
// keeps its fall distance and is still damaged for the fall. Player.Move calls EntityInside before it
// works out fall damage, so resetting the fall distance here would make falling onto a honey block from
// any height harmless. The player rests a fraction below the top of the block, as positions come from
// the client as float32 with the eye height subtracted.
func TestHoneyBlockLandingOnTopStillHurts(t *testing.T) {
	fallOntoHoney(t, 0.5, 60.999999, func(tx *world.Tx, p *player.Player) {
		fall := p.FallDistance()

		block.HoneyBlock{}.EntityInside(honeyPos, tx, p)

		if p.FallDistance() != fall {
			t.Errorf("expected fall distance %v to be kept for damage, got %v", fall, p.FallDistance())
		}
	})
}

// TestHoneyBlockSlidingResetsFallDistance verifies that a player sliding down the side of a honey block
// has its fall distance reset and takes no fall damage. The client stops a player exactly against the
// side of the block, so it rests at the side of the block plus half its own width and overlaps it by
// nothing at all, which must still count as sliding down it.
func TestHoneyBlockSlidingResetsFallDistance(t *testing.T) {
	fallOntoHoney(t, 15.0/16+0.3, 60, func(tx *world.Tx, p *player.Player) {
		block.HoneyBlock{}.EntityInside(honeyPos, tx, p)

		if p.FallDistance() != 0 {
			t.Errorf("expected fall distance 0, got %v", p.FallDistance())
		}
	})
}

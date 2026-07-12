package portal_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/portal"
)

// ringFrame is one (position, facing) pair on an end portal ring.
type ringFrame struct {
	pos    cube.Pos
	facing cube.Direction
}

// buildEndPortalRing places a complete twelve-frame end portal ring around center, with every frame holding an eye
// and facing the centre.
func buildEndPortalRing(tx *world.Tx, center cube.Pos) {
	for _, fp := range endPortalRingFrames(center) {
		tx.SetBlock(fp.pos, block.EndPortalFrame{Facing: fp.facing, Eye: true}, nil)
	}
}

// endRingOffsets is a hand-written oracle of the twelve (offset from centre, facing) pairs of a valid ring,
// independent of the production ring geometry.
var endRingOffsets = []ringFrame{
	{cube.Pos{-1, 0, -2}, cube.South}, {cube.Pos{0, 0, -2}, cube.South}, {cube.Pos{1, 0, -2}, cube.South},
	{cube.Pos{2, 0, -1}, cube.West}, {cube.Pos{2, 0, 0}, cube.West}, {cube.Pos{2, 0, 1}, cube.West},
	{cube.Pos{1, 0, 2}, cube.North}, {cube.Pos{0, 0, 2}, cube.North}, {cube.Pos{-1, 0, 2}, cube.North},
	{cube.Pos{-2, 0, 1}, cube.East}, {cube.Pos{-2, 0, 0}, cube.East}, {cube.Pos{-2, 0, -1}, cube.East},
}

// endPortalRingFrames returns the twelve (frame position, facing) pairs of a valid ring around center.
func endPortalRingFrames(center cube.Pos) []ringFrame {
	frames := make([]ringFrame, len(endRingOffsets))
	for i, f := range endRingOffsets {
		frames[i] = ringFrame{pos: center.Add(f.pos), facing: f.facing}
	}
	return frames
}

// interiorPositions returns the 3x3 interior block positions around center.
func interiorPositions(center cube.Pos) []cube.Pos {
	out := make([]cube.Pos, 0, 9)
	for dx := -1; dx <= 1; dx++ {
		for dz := -1; dz <= 1; dz++ {
			out = append(out, center.Add(cube.Pos{dx, 0, dz}))
		}
	}
	return out
}

func TestActivateEndPortal(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	center := cube.Pos{8, 10, 8}
	mustDo(t, w, func(tx *world.Tx) {
		buildEndPortalRing(tx, center)

		first := endPortalRingFrames(center)[0]
		if !portal.ActivateEndPortal(tx, first.pos) {
			t.Fatal("ActivateEndPortal() = false on a complete ring, want true")
		}
		for _, p := range interiorPositions(center) {
			if _, ok := tx.Block(p).(block.EndPortal); !ok {
				t.Fatalf("interior block at %v = %T, want block.EndPortal", p, tx.Block(p))
			}
		}
	})
}

func TestActivateEndPortalMissingEye(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	center := cube.Pos{8, 10, 8}
	mustDo(t, w, func(tx *world.Tx) {
		buildEndPortalRing(tx, center)

		frames := endPortalRingFrames(center)
		broken := frames[5]
		tx.SetBlock(broken.pos, block.EndPortalFrame{Facing: broken.facing, Eye: false}, nil)

		if portal.ActivateEndPortal(tx, frames[0].pos) {
			t.Fatal("ActivateEndPortal() = true with one missing eye, want false")
		}
		for _, p := range interiorPositions(center) {
			if _, ok := tx.Block(p).(block.EndPortal); ok {
				t.Fatalf("interior block at %v became EndPortal despite incomplete ring", p)
			}
		}
	})
}

func TestActivateEndPortalWrongFacing(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	center := cube.Pos{8, 10, 8}
	mustDo(t, w, func(tx *world.Tx) {
		buildEndPortalRing(tx, center)

		frames := endPortalRingFrames(center)
		bad := frames[3]
		tx.SetBlock(bad.pos, block.EndPortalFrame{Facing: bad.facing.Opposite(), Eye: true}, nil)

		if portal.ActivateEndPortal(tx, frames[0].pos) {
			t.Fatal("ActivateEndPortal() = true with one mis-facing frame, want false")
		}
	})
}

func TestEnderEyeCompletesRing(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	center := cube.Pos{8, 10, 8}
	mustDo(t, w, func(tx *world.Tx) {
		frames := endPortalRingFrames(center)
		for i, fp := range frames {
			tx.SetBlock(fp.pos, block.EndPortalFrame{Facing: fp.facing, Eye: i != 0}, nil)
		}

		first := frames[0]
		ctx := &item.UseContext{}
		if !(item.EnderEye{}).UseOnBlock(first.pos, cube.FaceUp, cube.Pos{}.Vec3(), tx, nil, ctx) {
			t.Fatal("EnderEye.UseOnBlock() = false on the last empty frame, want true")
		}
		if ctx.CountSub != 1 {
			t.Fatalf("EnderEye.UseOnBlock() subtracted %d items, want 1", ctx.CountSub)
		}
		for _, p := range interiorPositions(center) {
			if _, ok := tx.Block(p).(block.EndPortal); !ok {
				t.Fatalf("interior block at %v = %T, want block.EndPortal", p, tx.Block(p))
			}
		}
	})
}

func TestEndPortalDespawnsOnFrameBreak(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	center := cube.Pos{8, 10, 8}
	mustDo(t, w, func(tx *world.Tx) {
		buildEndPortalRing(tx, center)
		first := endPortalRingFrames(center)[0]
		if !portal.ActivateEndPortal(tx, first.pos) {
			t.Fatal("ActivateEndPortal() = false on a complete ring, want true")
		}

		// Break one frame: the portal blocks should despawn once the ring is no longer complete.
		broken := endPortalRingFrames(center)[5]
		tx.SetBlock(broken.pos, nil, nil)

		// Drive a neighbour update on an interior end_portal block, mimicking what the world does after the removal.
		ep, ok := tx.Block(center).(block.EndPortal)
		if !ok {
			t.Fatalf("centre block = %T, want block.EndPortal", tx.Block(center))
		}
		ep.NeighbourUpdateTick(center, broken.pos, tx)

		for _, p := range interiorPositions(center) {
			if _, ok := tx.Block(p).(block.EndPortal); ok {
				t.Fatalf("interior at %v still EndPortal after frame break", p)
			}
		}
	})
}

func TestEndPortalKeptOnPortalBlockBreak(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	center := cube.Pos{8, 10, 8}
	mustDo(t, w, func(tx *world.Tx) {
		buildEndPortalRing(tx, center)
		if !portal.ActivateEndPortal(tx, endPortalRingFrames(center)[0].pos) {
			t.Fatal("ActivateEndPortal() = false on a complete ring, want true")
		}

		// Break the centre portal block: the ring is still complete, so the other portal blocks must stay.
		tx.SetBlock(center, nil, nil)
		neighbour := center.Add(cube.Pos{1, 0, 0})
		ep, ok := tx.Block(neighbour).(block.EndPortal)
		if !ok {
			t.Fatalf("neighbour block = %T, want block.EndPortal", tx.Block(neighbour))
		}
		ep.NeighbourUpdateTick(neighbour, center, tx)

		for _, p := range interiorPositions(center) {
			if p == center {
				continue
			}
			if _, ok := tx.Block(p).(block.EndPortal); !ok {
				t.Fatalf("interior at %v despawned after breaking a portal block", p)
			}
		}
	})
}

func TestGenerateEndSpawnPlatformClearsThreeLayers(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	mustDo(t, w, func(tx *world.Tx) {
		for y := 49; y <= 52; y++ {
			tx.SetBlock(cube.Pos{100, y, 0}, block.Obsidian{}, nil)
		}
		portal.GenerateEndSpawnPlatform(tx)

		for y := 49; y <= 51; y++ {
			if _, ok := tx.Block(cube.Pos{100, y, 0}).(block.Air); !ok {
				t.Fatalf("block at y=%d = %T, want block.Air", y, tx.Block(cube.Pos{100, y, 0}))
			}
		}
		if _, ok := tx.Block(cube.Pos{100, 52, 0}).(block.Obsidian); !ok {
			t.Fatalf("block at y=52 = %T, want block.Obsidian", tx.Block(cube.Pos{100, 52, 0}))
		}
	})
}

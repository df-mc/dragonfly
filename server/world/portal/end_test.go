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

// endPortalRingFrames returns the twelve (frame position, facing) pairs of a valid ring around center.
func endPortalRingFrames(center cube.Pos) []ringFrame {
	frames := make([]ringFrame, 0, 12)
	for _, side := range cube.Directions() {
		base := center.Side(side.Face()).Side(side.Face())
		tangent := side.RotateRight().Face()
		inward := side.Opposite()
		for i := -1; i <= 1; i++ {
			pos := base
			step, n := tangent, i
			if n < 0 {
				step, n = step.Opposite(), -n
			}
			for range n {
				pos = pos.Side(step)
			}
			frames = append(frames, ringFrame{pos: pos, facing: inward})
		}
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
	<-w.Exec(func(tx *world.Tx) {
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
	<-w.Exec(func(tx *world.Tx) {
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
	<-w.Exec(func(tx *world.Tx) {
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
	<-w.Exec(func(tx *world.Tx) {
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

func TestGenerateEndSpawnPlatformClearsThreeLayers(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	<-w.Exec(func(tx *world.Tx) {
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

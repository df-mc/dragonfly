package portal_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/portal"
)

// ringFrame is one (position, outward facing) pair on an end portal ring.
type ringFrame struct {
	pos    cube.Pos
	facing cube.Direction
}

// buildEndPortalRing places a complete twelve-frame end portal ring on the y plane of center, with every frame having
// an eye and facing toward the centre (the vanilla-valid configuration).
func buildEndPortalRing(tx *world.Tx, center cube.Pos) {
	for _, fp := range endPortalRingFrames(center) {
		tx.SetBlock(fp.pos, block.EndPortalFrame{Facing: fp.facing, Eye: true}, nil)
	}
}

// endPortalRingFrames returns the twelve canonical (frame position, facing) pairs around center for a 3x3 interior
// portal. Each frame's Facing points TOWARD the centre — the opposite of the cardinal side it sits on — matching
// vanilla Bedrock's requirement (cardinal_direction = opposite of the placing player's facing, so build-from-centre
// yields inward-facing frames).
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

// interiorPositions returns the 3x3 interior block positions on the y plane of center.
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

		// Activate from the first frame.
		first := endPortalRingFrames(center)[0]
		if !portal.ActivateEndPortal(tx, first.pos) {
			t.Fatal("ActivateEndPortal() = false on a complete ring, want true")
		}

		// Every interior position must now hold an end_portal block.
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

		// Remove the eye from one frame.
		frames := endPortalRingFrames(center)
		broken := frames[5]
		tx.SetBlock(broken.pos, block.EndPortalFrame{Facing: broken.facing, Eye: false}, nil)

		// Activation must fail; no interior blocks placed.
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

		// Rotate one frame to face the wrong direction.
		frames := endPortalRingFrames(center)
		bad := frames[3]
		tx.SetBlock(bad.pos, block.EndPortalFrame{Facing: bad.facing.Opposite(), Eye: true}, nil)

		if portal.ActivateEndPortal(tx, frames[0].pos) {
			t.Fatal("ActivateEndPortal() = true with one mis-facing frame, want false")
		}
	})
}

func TestActivateEndPortalIdempotent(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	center := cube.Pos{8, 10, 8}
	<-w.Exec(func(tx *world.Tx) {
		buildEndPortalRing(tx, center)

		first := endPortalRingFrames(center)[0]
		if !portal.ActivateEndPortal(tx, first.pos) {
			t.Fatal("ActivateEndPortal() = false on first call, want true")
		}
		// Second call must not error and must leave interior intact.
		if !portal.ActivateEndPortal(tx, first.pos) {
			t.Fatal("ActivateEndPortal() = false on second call, want true")
		}
		for _, p := range interiorPositions(center) {
			if _, ok := tx.Block(p).(block.EndPortal); !ok {
				t.Fatalf("interior block at %v lost EndPortal after re-activation", p)
			}
		}
	})
}

func TestActivateEndPortalRejectsOutwardFacing(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	center := cube.Pos{8, 10, 8}
	<-w.Exec(func(tx *world.Tx) {
		// Build a ring with every frame facing OUTWARD (away from the centre). This is the configuration produced when
		// frames are placed by a player standing OUTSIDE the future ring. Vanilla Bedrock rejects this configuration.
		for _, fp := range endPortalRingFrames(center) {
			tx.SetBlock(fp.pos, block.EndPortalFrame{Facing: fp.facing.Opposite(), Eye: true}, nil)
		}
		first := endPortalRingFrames(center)[0]
		if portal.ActivateEndPortal(tx, first.pos) {
			t.Fatal("ActivateEndPortal() = true on outward-facing ring, want false")
		}
	})
}

func TestActivateEndPortalCornerIgnored(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	center := cube.Pos{8, 10, 8}
	<-w.Exec(func(tx *world.Tx) {
		buildEndPortalRing(tx, center)

		// Place an unrelated frame at one of the four corner positions (which are NOT part of the ring).
		// Activation must succeed regardless.
		corner := center.Add(cube.Pos{2, 0, 2})
		tx.SetBlock(corner, block.EndPortalFrame{Facing: cube.South, Eye: false}, nil)

		first := endPortalRingFrames(center)[0]
		if !portal.ActivateEndPortal(tx, first.pos) {
			t.Fatal("ActivateEndPortal() = false despite irrelevant corner block, want true")
		}
	})
}

func TestEnderEyeFillsFrame(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	pos := cube.Pos{8, 10, 8}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(pos, block.EndPortalFrame{Facing: cube.North, Eye: false}, nil)

		ctx := &item.UseContext{}
		if !(item.EnderEye{}).UseOnBlock(pos, cube.FaceUp, cube.Pos{}.Vec3(), tx, nil, ctx) {
			t.Fatal("EnderEye.UseOnBlock() = false on empty frame, want true")
		}
		if ctx.CountSub != 1 {
			t.Fatalf("EnderEye.UseOnBlock() subtracted %d items, want 1", ctx.CountSub)
		}
		f, ok := tx.Block(pos).(block.EndPortalFrame)
		if !ok {
			t.Fatalf("block at frame pos = %T, want block.EndPortalFrame", tx.Block(pos))
		}
		if !f.Eye {
			t.Fatalf("frame Eye = false after UseOnBlock, want true")
		}
	})
}

func TestEnderEyeOnFilledFrameNoOp(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	pos := cube.Pos{8, 10, 8}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(pos, block.EndPortalFrame{Facing: cube.North, Eye: true}, nil)

		ctx := &item.UseContext{}
		if (item.EnderEye{}).UseOnBlock(pos, cube.FaceUp, cube.Pos{}.Vec3(), tx, nil, ctx) {
			t.Fatal("EnderEye.UseOnBlock() = true on already-filled frame, want false")
		}
		if ctx.CountSub != 0 {
			t.Fatalf("EnderEye.UseOnBlock() consumed %d items on no-op, want 0", ctx.CountSub)
		}
	})
}

func TestEnderEyeOnNonFrameNoOp(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	pos := cube.Pos{8, 10, 8}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(pos, block.Obsidian{}, nil)

		ctx := &item.UseContext{}
		if (item.EnderEye{}).UseOnBlock(pos, cube.FaceUp, cube.Pos{}.Vec3(), tx, nil, ctx) {
			t.Fatal("EnderEye.UseOnBlock() = true on non-frame, want false")
		}
	})
}

func TestEnderEyeCompletesRing(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	center := cube.Pos{8, 10, 8}
	<-w.Exec(func(tx *world.Tx) {
		// Build the ring with all but the first frame filled.
		frames := endPortalRingFrames(center)
		for i, fp := range frames {
			tx.SetBlock(fp.pos, block.EndPortalFrame{Facing: fp.facing, Eye: i != 0}, nil)
		}

		first := frames[0]
		ctx := &item.UseContext{}
		if !(item.EnderEye{}).UseOnBlock(first.pos, cube.FaceUp, cube.Pos{}.Vec3(), tx, nil, ctx) {
			t.Fatal("EnderEye.UseOnBlock() = false on the last empty frame, want true")
		}
		// Interior must now be end_portal blocks.
		for _, p := range interiorPositions(center) {
			if _, ok := tx.Block(p).(block.EndPortal); !ok {
				t.Fatalf("interior block at %v = %T, want block.EndPortal", p, tx.Block(p))
			}
		}
	})
}

func TestEndPortalFromPosNonFrame(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	<-w.Exec(func(tx *world.Tx) {
		_, ok := portal.EndPortalFromPos(tx, cube.Pos{0, 50, 0})
		if ok {
			t.Fatal("EndPortalFromPos() ok = true on air block, want false")
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

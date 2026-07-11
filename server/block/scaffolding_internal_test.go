package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// TestScaffoldingPlacementPosSideways verifies that clicking a side face (not the bottom face) of an existing
// scaffolding block attaches the new block directly against that face, allowing horizontal extension.
func TestScaffoldingPlacementPosSideways(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	base := cube.Pos{0, 0, 0}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(base, Scaffolding{}, nil)
	})

	<-w.Exec(func(tx *world.Tx) {
		for _, face := range []cube.Face{cube.FaceNorth, cube.FaceSouth, cube.FaceEast, cube.FaceWest, cube.FaceUp} {
			want := base.Side(face)
			got, ok := scaffoldingPlacementPos(base, face, tx, Scaffolding{})
			if !ok {
				t.Errorf("face %v: expected placement to succeed", face)
				continue
			}
			if got != want {
				t.Errorf("face %v: expected placement at %v, got %v", face, want, got)
			}
		}
	})
}

// TestScaffoldingPlacementPosBottomFaceTowers verifies that clicking the underside of a scaffolding block
// redirects placement to the top of that block's column instead of attaching below it.
func TestScaffoldingPlacementPosBottomFaceTowers(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(cube.Pos{0, 0, 0}, Scaffolding{}, nil)
		tx.SetBlock(cube.Pos{0, 1, 0}, Scaffolding{}, nil)
		tx.SetBlock(cube.Pos{0, 2, 0}, Scaffolding{}, nil)
	})

	<-w.Exec(func(tx *world.Tx) {
		got, ok := scaffoldingPlacementPos(cube.Pos{0, 0, 0}, cube.FaceDown, tx, Scaffolding{})
		if !ok {
			t.Fatal("expected placement to succeed")
		}
		if want := (cube.Pos{0, 3, 0}); got != want {
			t.Errorf("expected placement at the top of the column %v, got %v", want, got)
		}
	})
}

// TestScaffoldingPlacementPosSidewaysOccupiedTargetFails verifies that clicking a sideways face whose target
// cell is already occupied by scaffolding simply fails to place, rather than redirecting to the top of whatever
// column the occupying block belongs to. This is a regression test: an earlier version redirected this case
// unconditionally, which meant clicking a side-branch block facing back towards an unrelated, much taller
// column would unexpectedly place a block at the top of that tall column instead of failing.
func TestScaffoldingPlacementPosSidewaysOccupiedTargetFails(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	// A 50-block tall tower, plus an unrelated single branch block one step to the west of its base.
	<-w.Exec(func(tx *world.Tx) {
		for y := 0; y < 50; y++ {
			tx.SetBlock(cube.Pos{0, y, 0}, Scaffolding{}, nil)
		}
		tx.SetBlock(cube.Pos{-1, 0, 0}, Scaffolding{}, nil)
	})

	<-w.Exec(func(tx *world.Tx) {
		// Clicking the east face of the branch block points back at the tower, whose neighbouring cell is
		// already occupied. This must fail rather than place at the top of the 50-tall tower.
		if _, ok := scaffoldingPlacementPos(cube.Pos{-1, 0, 0}, cube.FaceEast, tx, Scaffolding{}); ok {
			t.Error("expected placement against an already-occupied cell to fail")
		}
	})
}

// TestScaffoldingPlacementPosTopFaceOfLowerBlockTowers verifies that clicking the top face of a scaffolding
// block that is NOT the topmost one in its column (e.g. missing the exact tip due to how thin the model's click
// target is) still redirects to the true top of that same column, rather than failing. This is what makes
// towering reliable even when the click does not land on the exact topmost block. Unlike the sideways case, this
// is safe because climbing straight up from the clicked block can only ever reach blocks in its own column.
func TestScaffoldingPlacementPosTopFaceOfLowerBlockTowers(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(cube.Pos{0, 0, 0}, Scaffolding{}, nil)
		tx.SetBlock(cube.Pos{0, 1, 0}, Scaffolding{}, nil)
		tx.SetBlock(cube.Pos{0, 2, 0}, Scaffolding{}, nil)
	})

	<-w.Exec(func(tx *world.Tx) {
		// Click the bottom block's top face, even though blocks 1 and 2 already sit above it.
		got, ok := scaffoldingPlacementPos(cube.Pos{0, 0, 0}, cube.FaceUp, tx, Scaffolding{})
		if !ok {
			t.Fatal("expected placement to succeed")
		}
		if want := (cube.Pos{0, 3, 0}); got != want {
			t.Errorf("expected placement at the top of the column %v, got %v", want, got)
		}
	})
}

// TestScaffoldingBuildUpFromBranchTip verifies that building upward off the tip of a horizontal branch (clicking
// the top face of the outermost branch block, as when standing on it) resolves to the cell directly above,
// inherits that block's stability, and persists across ticks rather than being reverted.
func TestScaffoldingBuildUpFromBranchTip(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(cube.Pos{0, 0, 0}, Stone{}, nil)
		tx.SetBlock(cube.Pos{0, 1, 0}, Scaffolding{}, nil)
		tx.SetBlock(cube.Pos{-1, 1, 0}, Scaffolding{Stability: 1}, nil)
		tx.SetBlock(cube.Pos{-2, 1, 0}, Scaffolding{Stability: 2}, nil)
	})
	w.AdvanceTick()

	tip := cube.Pos{-2, 1, 0}
	var above cube.Pos
	<-w.Exec(func(tx *world.Tx) {
		resolved, ok := scaffoldingPlacementPos(tip, cube.FaceUp, tx, Scaffolding{})
		if !ok {
			t.Fatal("expected build-up from the branch tip to resolve successfully")
		}
		above = resolved
		stability, stabilityCheck := scaffoldingStability(above, tx)
		tx.SetBlock(above, Scaffolding{Stability: stability, StabilityCheck: stabilityCheck}, nil)
	})
	for i := 0; i < 3; i++ {
		w.AdvanceTick()
	}

	<-w.Exec(func(tx *world.Tx) {
		s, ok := tx.Block(above).(Scaffolding)
		if !ok {
			t.Fatalf("expected scaffolding above the branch tip at %v, got %v", above, tx.Block(above))
		}
		if s.Stability != 2 {
			t.Errorf("expected the block above the branch tip to inherit stability 2, got %d", s.Stability)
		}
	})
}

// TestScaffoldingRecognisedAsLavaFlammableNeighbour verifies that Lava's own ignition check (neighboursLavaFlammable,
// used by Lava.RandomTick to decide whether to start a Fire block nearby) recognises Scaffolding as a valid
// target now that LavaFlammable is true. This is what makes lava ignite scaffolding through the normal mechanism
// instead of needing a bespoke instant-destroy path.
func TestScaffoldingRecognisedAsLavaFlammableNeighbour(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	air := cube.Pos{0, 0, 0}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(air.Side(cube.FaceEast), Scaffolding{}, nil)

		if !neighboursLavaFlammable(air, tx) {
			t.Error("expected the air cell next to scaffolding to be recognised as having a lava-flammable neighbour")
		}
	})
}

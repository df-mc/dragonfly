package block_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// advanceTicks advances the synchronous world passed by the number of ticks passed, allowing scheduled block
// updates (such as scaffolding stability checks) to propagate.
func advanceTicks(w *world.World, n int) {
	for i := 0; i < n; i++ {
		w.AdvanceTick()
	}
}

// TestScaffoldingStabilityExtends verifies that scaffolding placed on a solid block has a stability of 0 and that
// stability increases by one for every block it is extended horizontally.
func TestScaffoldingStabilityExtends(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	base := cube.Pos{0, 1, 0}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(cube.Pos{0, 0, 0}, block.Stone{}, nil)
		tx.SetBlock(base, block.Scaffolding{}, nil)
		tx.SetBlock(cube.Pos{1, 1, 0}, block.Scaffolding{}, nil)
		tx.SetBlock(cube.Pos{2, 1, 0}, block.Scaffolding{}, nil)
	})
	advanceTicks(w, 5)

	<-w.Exec(func(tx *world.Tx) {
		for pos, want := range map[cube.Pos]int{base: 0, {1, 1, 0}: 1, {2, 1, 0}: 2} {
			s, ok := tx.Block(pos).(block.Scaffolding)
			if !ok {
				t.Errorf("expected scaffolding at %v, got %v", pos, tx.Block(pos))
			} else if s.Stability != want {
				t.Errorf("expected stability %d at %v, got %d", want, pos, s.Stability)
			}
		}
	})
}

// TestScaffoldingColumnCollapses verifies that removing the bottom of a scaffolding column breaks every block
// above it, matching the vanilla behaviour where the whole column drops as items.
func TestScaffoldingColumnCollapses(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(cube.Pos{0, 0, 0}, block.Stone{}, nil)
		tx.SetBlock(cube.Pos{0, 1, 0}, block.Scaffolding{}, nil)
		tx.SetBlock(cube.Pos{0, 2, 0}, block.Scaffolding{}, nil)
		tx.SetBlock(cube.Pos{0, 3, 0}, block.Scaffolding{}, nil)
	})
	advanceTicks(w, 3)

	// Remove the bottom scaffolding block: the two blocks above it should lose their support and break.
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(cube.Pos{0, 1, 0}, block.Air{}, nil)
	})
	advanceTicks(w, 10)

	<-w.Exec(func(tx *world.Tx) {
		for _, pos := range []cube.Pos{{0, 2, 0}, {0, 3, 0}} {
			if b := tx.Block(pos); b != (block.Air{}) {
				t.Errorf("expected scaffolding at %v to collapse, got %v", pos, b)
			}
		}
	})
}

// TestScaffoldingUnsupportedFalls verifies that a scaffolding block that is already unsupported (stability 7)
// leaves its position and falls, as happens when scaffolding is extended one block too far.
func TestScaffoldingUnsupportedFalls(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	pos := cube.Pos{0, 5, 0}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(pos, block.Scaffolding{Stability: 7}, nil)
	})
	advanceTicks(w, 3)

	<-w.Exec(func(tx *world.Tx) {
		if b := tx.Block(pos); b != (block.Air{}) {
			t.Errorf("expected unsupported scaffolding to fall, got %v", b)
		}
	})
}

// TestScaffoldingMaxHorizontalReach verifies that scaffolding can extend exactly 6 blocks out from its support
// (stability 0 through 6) but that the 7th block placed beyond it (stability 7) does not persist as a block,
// matching the documented vanilla limit.
func TestScaffoldingMaxHorizontalReach(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(cube.Pos{0, 0, 0}, block.Stone{}, nil)
		for x := 0; x <= 7; x++ {
			tx.SetBlock(cube.Pos{x, 1, 0}, block.Scaffolding{}, nil)
		}
	})
	advanceTicks(w, 10)

	<-w.Exec(func(tx *world.Tx) {
		for x := 0; x <= 6; x++ {
			pos := cube.Pos{x, 1, 0}
			s, ok := tx.Block(pos).(block.Scaffolding)
			if !ok {
				t.Errorf("expected scaffolding at %v within the 6-block limit, got %v", pos, tx.Block(pos))
			} else if s.Stability != x {
				t.Errorf("expected stability %d at %v, got %d", x, pos, s.Stability)
			}
		}
		if b := tx.Block(cube.Pos{7, 1, 0}); b != (block.Air{}) {
			t.Errorf("expected scaffolding beyond the 6-block limit to fall, got %v", b)
		}
	})
}

// TestScaffoldingDestroyedNextToLava verifies that scaffolding placed next to lava is destroyed outright
// (dropping nothing), matching Bedrock's behaviour, rather than merely catching fire and burning down over time
// like its Flammability would otherwise cause.
func TestScaffoldingDestroyedNextToLava(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	pos := cube.Pos{0, 0, 0}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(pos, block.Scaffolding{}, nil)
		tx.SetBlock(pos.Side(cube.FaceEast), block.Lava{Depth: 8}, nil)
	})
	advanceTicks(w, 3)

	<-w.Exec(func(tx *world.Tx) {
		if b := tx.Block(pos); b != (block.Air{}) {
			t.Errorf("expected scaffolding next to lava to be destroyed, got %v", b)
		}
	})
}

// TestScaffoldingDestroyedNextToFire verifies that scaffolding placed next to fire is destroyed outright, the
// same as it is next to lava.
func TestScaffoldingDestroyedNextToFire(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	pos := cube.Pos{0, 0, 0}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(pos, block.Scaffolding{}, nil)
		tx.SetBlock(pos.Side(cube.FaceEast), block.Fire{}, nil)
	})
	advanceTicks(w, 3)

	<-w.Exec(func(tx *world.Tx) {
		if b := tx.Block(pos); b != (block.Air{}) {
			t.Errorf("expected scaffolding next to fire to be destroyed, got %v", b)
		}
	})
}

// TestScaffoldingCannotBePlacedInLava verifies that scaffolding cannot be placed directly into a lava-filled
// cell, matching Bedrock's behaviour, even though Lava is normally replaceable by most other blocks.
func TestScaffoldingCannotBePlacedInLava(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	pos := cube.Pos{0, 0, 0}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(pos, block.Lava{Depth: 8}, nil)
	})

	<-w.Exec(func(tx *world.Tx) {
		used := block.Scaffolding{}.UseOnBlock(pos, cube.FaceUp, mgl64.Vec3{}, tx, nil, &item.UseContext{})
		if used {
			t.Error("expected placement into lava to fail")
		}
		if b := tx.Block(pos); b != (block.Lava{Depth: 8}) {
			t.Errorf("expected lava to remain in place, got %v", b)
		}
	})
}

// TestScaffoldingNeverWaterlogs verifies that scaffolding placed into water displaces the water outright instead
// of coexisting with it as a waterlogged block. This is a deliberate departure from vanilla: a waterlogged
// climbable block cannot be climbed on Bedrock, which permanently traps a player at the point where a scaffolding
// column transitions from a waterlogged section to a dry one. Displacing the water instead guarantees the whole
// column always stays climbable.
func TestScaffoldingNeverWaterlogs(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	pos := cube.Pos{0, 0, 0}
	<-w.Exec(func(tx *world.Tx) {
		tx.SetLiquid(pos, block.Water{Depth: 8})
	})

	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(pos, block.Scaffolding{}, nil)
	})

	<-w.Exec(func(tx *world.Tx) {
		if _, ok := tx.Block(pos).(block.Scaffolding); !ok {
			t.Fatalf("expected scaffolding at %v, got %v", pos, tx.Block(pos))
		}
		if _, ok := tx.Liquid(pos); ok {
			t.Error("expected water to be displaced, but it is still present")
		}
	})
}

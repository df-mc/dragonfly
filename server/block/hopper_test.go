package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// TestHopperInsertIntoPairedChest verifies that a hopper inserts into the container it faces. Chest.Inventory resolves
// the other half of a double chest relative to the position it is given, so passing the hopper's own position paired
// the destination with whatever happened to be beside the hopper instead.
//
// Chest.pairPos only takes the Y coordinate from the position passed, so this only goes wrong for a hopper above or
// below the chest, not beside it.
func TestHopperInsertIntoPairedChest(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	hopperPos, destPos := cube.Pos{0, 64, 0}, cube.Pos{0, 63, 0}
	pairPos, besideHopper := cube.Pos{1, 63, 0}, cube.Pos{1, 64, 0}

	var stored, held int
	var beside world.Block
	runWorld(w, func(tx *world.Tx) {
		h := NewHopper()
		h.Facing = cube.FaceDown

		// The destination chest, paired with the chest next to it.
		dest := NewChest()
		dest.Facing, dest.paired = cube.South, true
		dest.pairX, dest.pairZ = pairPos[0], pairPos[2]

		pair := NewChest()
		pair.Facing, pair.paired = cube.South, true
		pair.pairX, pair.pairZ = destPos[0], destPos[2]

		// An unrelated chest beside the hopper, at the position the pair used to be looked for at.
		unrelated := NewChest()
		unrelated.Facing = cube.South

		tx.SetBlock(destPos, dest, nil)
		tx.SetBlock(pairPos, pair, nil)
		tx.SetBlock(besideHopper, unrelated, nil)
		tx.SetBlock(hopperPos, h, nil)
		if err := h.inventory.SetItem(0, item.NewStack(item.Diamond{}, 4)); err != nil {
			t.Fatalf("SetItem() = %v, want nil", err)
		}

		h.Tick(0, hopperPos, tx)

		// The chest is read back from the world: pairing replaces both halves with copies sharing one inventory.
		stored = stackCount(tx.Block(destPos).(Chest).Inventory(tx, destPos).Items())
		held, beside = stackCount(h.inventory.Items()), tx.Block(besideHopper)
	})

	if stored != 1 {
		t.Errorf("destination chest holds %v diamonds, want 1: the hopper holds %v of the 4 it started with", stored, held)
	}
	if c, ok := beside.(Chest); ok && c.paired {
		t.Errorf("unrelated chest beside the hopper was paired to (%v, %v), want it untouched", c.pairX, c.pairZ)
	}
}

// TestHopperInsertKeepsHopperBlock verifies that inserting does not replace the hopper itself. Chest.Inventory writes
// the chest back at the position it is given when it cannot resolve its pair, so a hopper that passed its own position
// was overwritten by its destination chest.
func TestHopperInsertKeepsHopperBlock(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	hopperPos, destPos := cube.Pos{0, 64, 0}, cube.Pos{0, 63, 0}

	var atHopper world.Block
	runWorld(w, func(tx *world.Tx) {
		h := NewHopper()
		h.Facing = cube.FaceDown

		// A chest as it is left after being decoded from disk: paired, but with the pair not resolved yet, and with a
		// partner that does not hold a chest.
		c := NewChest()
		c.Facing, c.paired = cube.South, true
		c.pairX, c.pairZ = 1, 0

		tx.SetBlock(destPos, c, nil)
		tx.SetBlock(hopperPos, h, nil)
		if err := h.inventory.SetItem(0, item.NewStack(item.Diamond{}, 4)); err != nil {
			t.Fatalf("SetItem() = %v, want nil", err)
		}

		// insertItem is called directly: Hopper.Tick writes the hopper back itself, which would mask the result.
		h.insertItem(hopperPos, tx)
		atHopper = tx.Block(hopperPos)
	})

	if _, ok := atHopper.(Hopper); !ok {
		t.Fatalf("block at %v = %T, want block.Hopper", hopperPos, atHopper)
	}
}

// stackCount returns the total number of items across the stacks passed.
func stackCount(stacks []item.Stack) int {
	var n int
	for _, s := range stacks {
		n += s.Count()
	}
	return n
}

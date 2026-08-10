package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// TestHopperIntoShulkerBoxKeepsRefusedItem verifies that a hopper does not destroy an item a shulker box refuses to
// hold. A shulker box cannot contain another shulker box, and AddItem used to report a refused write as stored, so the
// hopper removed the item from itself for a transfer that never happened.
func TestHopperIntoShulkerBoxKeepsRefusedItem(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	hopperPos, boxPos := cube.Pos{0, 64, 0}, cube.Pos{0, 63, 0}

	var held, stored int
	runWorld(w, func(tx *world.Tx) {
		h := NewHopper()
		h.Facing = cube.FaceDown

		box := NewShulkerBox()
		box.Facing = cube.FaceUp

		tx.SetBlock(boxPos, box, nil)
		tx.SetBlock(hopperPos, h, nil)
		if err := h.inventory.SetItem(0, item.NewStack(NewShulkerBox(), 1)); err != nil {
			t.Fatalf("SetItem() = %v, want nil", err)
		}

		h.Tick(0, hopperPos, tx)

		held, stored = len(h.inventory.Items()), len(box.Inventory(tx, boxPos).Items())
	})

	if held != 1 {
		t.Errorf("hopper holds %v stacks, want 1: the shulker box was destroyed", held)
	}
	if stored != 0 {
		t.Errorf("destination shulker box holds %v stacks, want 0", stored)
	}
}

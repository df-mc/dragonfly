package block

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// TestChestPairingKeepsInventories verifies that pairing and unpairing a chest keeps the inventories it already had.
// Anything already looking at a chest holds the inventory it was opened with, and replacing it with a copy leaves that
// holding a second set of the same items.
func TestChestPairingKeepsInventories(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	left, right := cube.Pos{0, 64, 0}, cube.Pos{1, 64, 0}
	w.Do(func(tx *world.Tx) {
		a, b := NewChest(), NewChest()
		a.Facing, b.Facing = cube.South, cube.South
		tx.SetBlock(left, a, nil)
		tx.SetBlock(right, b, nil)
		if err := a.inventory.SetItem(0, item.NewStack(item.Diamond{}, 64)); err != nil {
			t.Fatalf("SetItem() = %v, want nil", err)
		}
		opened := a.inventory

		paired, pair, ok := a.pair(tx, left, right)
		if !ok {
			t.Fatal("expected the two chests to pair")
		}
		if paired.inventory != opened && pair.inventory != opened {
			t.Error("pairing replaced the inventory of both chests, leaving an open window holding a copy")
		}

		unpaired, _, ok := paired.unpair(tx, left)
		if !ok {
			t.Fatal("expected the chests to unpair")
		}
		if unpaired.inventory != opened {
			t.Error("unpairing replaced the inventory of the chest, leaving an open window holding a copy")
		}
		if got, _ := unpaired.inventory.Item(0); got.Count() != 64 {
			t.Errorf("chest holds %v diamonds after pairing and unpairing, want 64", got.Count())
		}
	}).Wait(context.Background())
}

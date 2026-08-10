package session

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// TestHandleLoomCraftResultCount verifies that a loom produces as many banners as it consumes. The number of crafts is
// client controlled, so producing a result at the count of the whole input stack let a client turn one banner and one
// dye into a full stack.
func TestHandleLoomCraftResultCount(t *testing.T) {
	tests := []struct {
		name         string
		inputCount   int
		timesCrafted byte
	}{
		{name: "one craft out of a full stack", inputCount: 16, timesCrafted: 1},
		{name: "several crafts out of a full stack", inputCount: 16, timesCrafted: 4},
		{name: "every banner in the stack", inputCount: 16, timesCrafted: 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := world.Config{Synchronous: true}.New()
			defer w.Close()

			pos, ui := cube.Pos{0, 64, 0}, inventory.New(54, nil)
			if err := ui.SetItem(loomInputSlot, item.NewStack(block.Banner{Colour: item.ColourRed()}, tt.inputCount)); err != nil {
				t.Fatalf("SetItem() = %v, want nil", err)
			}
			if err := ui.SetItem(loomDyeSlot, item.NewStack(item.Dye{Colour: item.ColourBlue()}, tt.inputCount)); err != nil {
				t.Fatalf("SetItem() = %v, want nil", err)
			}

			s := &Session{ui: ui}
			s.openedPos.Store(&pos)
			s.openedWindow.Store(inventory.New(1, nil))
			s.containerOpened.Store(true)
			h := &ItemStackRequestHandler{
				changes:         map[byte]map[byte]changeInfo{},
				responseChanges: map[int32]map[*inventory.Inventory]map[byte]responseChange{},
			}

			var err error
			w.Do(func(tx *world.Tx) {
				tx.SetBlock(pos, block.Loom{}, nil)
				err = h.handleLoomCraft(&protocol.CraftLoomRecipeStackRequestAction{
					// The border pattern needs no banner pattern item in the third slot.
					Pattern:      "bo",
					TimesCrafted: tt.timesCrafted,
				}, s, tx)
			}).Wait(context.Background())
			if err != nil {
				t.Fatalf("handleLoomCraft() = %v, want nil", err)
			}

			result, err := ui.Item(craftingResult)
			if err != nil {
				t.Fatalf("Item() = %v, want nil", err)
			}
			if got := result.Count(); got != int(tt.timesCrafted) {
				input, _ := ui.Item(loomInputSlot)
				t.Errorf("result of %v crafts = %v banners, want %v: the input went %v to %v, so %v were created",
					tt.timesCrafted, got, tt.timesCrafted, tt.inputCount, input.Count(), got-int(tt.timesCrafted))
			}
		})
	}
}

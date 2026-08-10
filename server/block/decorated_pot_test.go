package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// TestHopperDecoratedPotMaxCount verifies that a hopper cannot insert into a decorated pot that already holds a full
// stack. The pot stores its contents in a plain field rather than an inventory, so nothing clamped the count for it,
// and breaking an overfilled pot dropped a stack larger than the item allows.
func TestHopperDecoratedPotMaxCount(t *testing.T) {
	tests := []struct {
		name  string
		start int
		want  int
	}{
		{name: "pot with room left", start: 63, want: 64},
		{name: "pot holding a full stack", start: 64, want: 64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := world.Config{Synchronous: true}.New()
			defer w.Close()

			hopperPos, potPos := cube.Pos{0, 64, 0}, cube.Pos{0, 63, 0}

			var stored item.Stack
			runWorld(w, func(tx *world.Tx) {
				h := NewHopper()
				h.Facing = cube.FaceDown

				tx.SetBlock(potPos, DecoratedPot{Item: item.NewStack(item.Diamond{}, tt.start)}, nil)
				tx.SetBlock(hopperPos, h, nil)
				if err := h.inventory.SetItem(0, item.NewStack(item.Diamond{}, 1)); err != nil {
					t.Fatalf("SetItem() = %v, want nil", err)
				}

				h.Tick(0, hopperPos, tx)

				stored = tx.Block(potPos).(DecoratedPot).Item
			})

			if got := stored.Count(); got != tt.want {
				t.Errorf("pot holds %v diamonds, want %v (maximum count is %v)", got, tt.want, stored.MaxCount())
			}
		})
	}
}

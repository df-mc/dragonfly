package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// TestClusterBlockDropsCarryNoState verifies that a block holding several of an item drops items that are worth one
// each. The dropped stack is built from the block, so a block that records how many it holds would hand out items
// that place as the whole cluster again, and each break would double what a player has.
func TestClusterBlockDropsCarryNoState(t *testing.T) {
	tests := []struct {
		name  string
		block Breakable
		want  int
		// single is the block the dropped item must place as.
		single world.Block
	}{
		{name: "double slab", block: Slab{Block: Stone{}, Double: true}, want: 2, single: Slab{Block: Stone{}}},
		{name: "four candles", block: Candle{AdditionalCandles: 3}, want: 4, single: Candle{}},
		{name: "four sea pickles", block: SeaPickle{AdditionalCount: 3}, want: 4, single: SeaPickle{}},
		{name: "four pink petals", block: PinkPetals{AdditionalCount: 3}, want: 4, single: PinkPetals{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drops := tt.block.BreakInfo().Drops(item.ToolNone{}, nil)
			if len(drops) != 1 {
				t.Fatalf("dropped %v stacks, want 1", len(drops))
			}
			if got := drops[0].Count(); got != tt.want {
				t.Errorf("dropped %v items, want %v", got, tt.want)
			}
			if got, ok := drops[0].Item().(world.Block); !ok || got != tt.single {
				t.Errorf("dropped item is %#v, want %#v: placing it back would rebuild the whole cluster", got, tt.single)
			}
		})
	}
}

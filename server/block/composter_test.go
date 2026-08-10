package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/item"
)

// TestComposterDropsEmpty verifies that a composter drops an empty one however full it was. The item is placed as the
// block it holds, so an item carrying the level would put a full composter back down and give its bone meal again.
func TestComposterDropsEmpty(t *testing.T) {
	for _, level := range []int{0, 4, 8} {
		drops := Composter{Level: level}.BreakInfo().Drops(item.ToolNone{}, nil)
		if len(drops) == 0 {
			t.Fatalf("composter at level %v dropped nothing", level)
		}
		got, ok := drops[0].Item().(Composter)
		if !ok {
			t.Fatalf("composter at level %v dropped %T, want a Composter", level, drops[0].Item())
		}
		if got.Level != 0 {
			t.Errorf("composter at level %v dropped one at level %v, want 0", level, got.Level)
		}
	}
}

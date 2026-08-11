package block_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// TestTintedGlassBreakInfo verifies the breaking properties of tinted glass match vanilla: hardness and blast
// resistance of 0.3, harvestable by hand, and no tool being more effective than an empty hand.
func TestTintedGlassBreakInfo(t *testing.T) {
	info := block.TintedGlass{}.BreakInfo()
	if info.Hardness != 0.3 {
		t.Errorf("expected hardness 0.3, got %v", info.Hardness)
	}
	if info.BlastResistance != 0.3 {
		t.Errorf("expected blast resistance 0.3, got %v", info.BlastResistance)
	}
	if !info.Harvestable(item.ToolNone{}) {
		t.Error("expected tinted glass to be harvestable by hand")
	}
	if info.Effective(item.ToolNone{}) {
		t.Error("expected no tool to be effective against tinted glass")
	}
}

// TestTintedGlassAlwaysDropsItself verifies that tinted glass drops itself when broken by hand, unlike regular
// glass which only drops with Silk Touch.
func TestTintedGlassAlwaysDropsItself(t *testing.T) {
	drops := block.TintedGlass{}.BreakInfo().Drops(item.ToolNone{}, nil)
	if len(drops) != 1 {
		t.Fatalf("expected exactly one drop, got %d", len(drops))
	}
	if _, ok := drops[0].Item().(block.TintedGlass); !ok {
		t.Errorf("expected tinted glass to drop itself, got %T", drops[0].Item())
	}

	glassDrops := block.Glass{}.BreakInfo().Drops(item.ToolNone{}, nil)
	if len(glassDrops) != 0 {
		t.Errorf("expected regular glass to drop nothing by hand, got %d drops", len(glassDrops))
	}
}

// TestTintedGlassEncode verifies the block and item encode to the vanilla identifier.
func TestTintedGlassEncode(t *testing.T) {
	name, _ := block.TintedGlass{}.EncodeBlock()
	if name != "minecraft:tinted_glass" {
		t.Errorf("expected block name minecraft:tinted_glass, got %q", name)
	}
	itemName, meta := block.TintedGlass{}.EncodeItem()
	if itemName != "minecraft:tinted_glass" || meta != 0 {
		t.Errorf("expected item minecraft:tinted_glass meta 0, got %q meta %d", itemName, meta)
	}
}

// TestTintedGlassSolidAndSuffocationImmune verifies that tinted glass uses a solid model yet reports itself as
// immune to suffocation, the property that keeps entities inside it from taking suffocation damage.
func TestTintedGlassSolidAndSuffocationImmune(t *testing.T) {
	g := block.TintedGlass{}
	if _, ok := g.Model().(model.Solid); !ok {
		t.Errorf("expected tinted glass to use a solid model, got %T", g.Model())
	}
	immune, ok := any(g).(block.NonSuffocating)
	if !ok {
		t.Fatal("expected tinted glass to implement block.NonSuffocating")
	}
	if !immune.SuffocationImmune() {
		t.Error("expected tinted glass to be suffocation immune")
	}
}

// TestTintedGlassBlocksLight verifies that, unlike regular glass, tinted glass fully blocks light: it registers as
// the highest light-blocking block in its column, whereas transparent glass does not.
func TestTintedGlassBlocksLight(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	tinted, plain := cube.Pos{0, 40, 0}, cube.Pos{4, 40, 0}
	w.Do(func(tx *world.Tx) {
		tx.SetBlock(tinted, block.TintedGlass{}, nil)
		tx.SetBlock(plain, block.Glass{}, nil)

		if got := tx.HighestLightBlocker(tinted[0], tinted[2]); got != tinted[1] {
			t.Errorf("expected tinted glass to block light at y=%d, highest blocker was y=%d", tinted[1], got)
		}
		if got := tx.HighestLightBlocker(plain[0], plain[2]); got == plain[1] {
			t.Errorf("expected transparent glass not to block light at y=%d, but it did", plain[1])
		}
	})
}

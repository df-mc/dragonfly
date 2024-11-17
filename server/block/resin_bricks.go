package block

import "github.com/df-mc/dragonfly/server/world"

type (
	// ResinBricks ...
	ResinBricks struct {
		solid
		bassDrum
	}

	// ChiseledResinBricks ...
	ChiseledResinBricks struct {
		solid
		bassDrum
	}
)

// BreakInfo ...
func (r ResinBricks) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(r)).withBlastResistance(30)
}

// BreakInfo ...
func (c ChiseledResinBricks) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(c)).withBlastResistance(30)
}

// EncodeItem ...
func (ResinBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:resin_bricks", 0
}

// EncodeItem ...
func (ChiseledResinBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:chiseled_resin_bricks", 0
}

// EncodeBlock ...
func (ResinBricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:resin_bricks", nil
}

// EncodeBlock ...
func (ChiseledResinBricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:chiseled_resin_bricks", nil
}

// allResinBricks ...
func allResinBricks() (resinBricks []world.Block) {
	resinBricks = append(resinBricks, ResinBricks{})
	resinBricks = append(resinBricks, ChiseledResinBricks{})
	return
}

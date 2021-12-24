package block

import "github.com/df-mc/dragonfly/server/world"

type StoneBricks struct {
	solid
	bassDrum

	// Type is the type of stone bricks of the block.
	Type StoneBricksType
}

// BreakInfo ...
func (c StoneBricks) BreakInfo() BreakInfo {
	return newBreakInfo(c.Type.Hardness(), pickaxeHarvestable, pickaxeEffective, oneOf(c))
}

// EncodeItem ...
func (c StoneBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:stonebrick", int16(c.Type.Uint8())
}

// EncodeBlock ...
func (c StoneBricks) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:stonebrick", map[string]interface{}{
		"stone_brick_type": c.Type.String(),
	}
}

// allStoneBricks returns a list of all stoneBricks block variants.
func allStoneBricks() (c []world.Block) {
	for _, t := range StoneBricksTypes() {
		c = append(c, StoneBricks{Type: t})
	}
	return
}

package block

import "github.com/df-mc/dragonfly/server/world"

// StoneBricks are materials found in structures such as strongholds, igloo basements, jungle temples, ocean ruins
// and ruined portals.
type StoneBricks struct {
	solid
	bassDrum

	// Type is the type of stone bricks of the block.
	Type StoneBricksType
}

// BreakInfo ...
func (c StoneBricks) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(c))
}

// EncodeItem ...
func (c StoneBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:stonebrick", int16(c.Type.Uint8())
}

// EncodeBlock ...
func (c StoneBricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:stonebrick", map[string]any{"stone_brick_type": c.Type.String()}
}

// allStoneBricks returns a list of all stoneBricks block variants.
func allStoneBricks() (c []world.Block) {
	for _, t := range StoneBricksTypes() {
		c = append(c, StoneBricks{Type: t})
	}
	return
}

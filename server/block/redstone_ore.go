package block

import "github.com/df-mc/dragonfly/server/item"

// RedstoneOre is a common ore.
type RedstoneOre struct {
	solid
	bassDrum

	// Type is the type of redstone ore.
	Type OreType
}

// BreakInfo ...
func (r RedstoneOre) BreakInfo() BreakInfo {
	return newBreakInfo(r.Type.Hardness(), pickaxeHarvestable, pickaxeEffective, oreDrops(item.Redstone{}, r)).withXPDropRange(0, 2).withBlastResistance(15)
}

// SmeltInfo ...
func (RedstoneOre) SmeltInfo() item.SmeltInfo {
	return newOreSmeltInfo(item.NewStack(item.Redstone{}, 1), 0.1)
}

// EncodeItem ...
func (r RedstoneOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + r.Type.Prefix() + "redstone_ore", 0
}

// EncodeBlock ...
func (r RedstoneOre) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + r.Type.Prefix() + "redstone_ore", nil

}

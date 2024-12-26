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
func (c RedstoneOre) BreakInfo() BreakInfo {
	i := newBreakInfo(c.Type.Hardness(), func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierIron.HarvestLevel
	}, pickaxeEffective, silkTouchOneOf(RedstoneWire{}, c)).withXPDropRange(1, 5)
	if c.Type == DeepslateOre() {
		i = i.withBlastResistance(9)
	}
	return i
}

// SmeltInfo ...
func (RedstoneOre) SmeltInfo() item.SmeltInfo {
	return newOreSmeltInfo(item.NewStack(RedstoneWire{}, 1), 0.7)
}

// EncodeItem ...
func (c RedstoneOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + c.Type.Prefix() + "redstone_ore", 0
}

// EncodeBlock ...
func (c RedstoneOre) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + c.Type.Prefix() + "redstone_ore", nil

}

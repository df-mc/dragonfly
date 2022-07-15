package block

import "github.com/df-mc/dragonfly/server/item"

// CoalOre is a common ore.
type CoalOre struct {
	solid
	bassDrum

	// Type is the type of coal ore.
	Type OreType
}

// BreakInfo ...
func (c CoalOre) BreakInfo() BreakInfo {
	return newBreakInfo(c.Type.Hardness(), pickaxeHarvestable, pickaxeEffective, silkTouchOneOf(item.Coal{}, c)).withXPDropRange(0, 2)
}

// EncodeItem ...
func (c CoalOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + c.Type.Prefix() + "coal_ore", 0
}

// EncodeBlock ...
func (c CoalOre) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + c.Type.Prefix() + "coal_ore", nil

}

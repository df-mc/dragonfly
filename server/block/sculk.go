package block

import "github.com/df-mc/dragonfly/server/item"

// Sculk is a bioluminescent block found abundantly in the deep dark
type Sculk struct {
	solid
}

// BreakInfo ...
func (s Sculk) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, alwaysHarvestable, hoeEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(s, 1)}
		}
		return nil
	}).withXPDropRange(1, 1)
}

// EncodeItem ...
func (s Sculk) EncodeItem() (name string, meta int16) {
	return "minecraft:sculk", 0
}

// EncodeBlock ...
func (s Sculk) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:sculk", nil
}

package item

import "github.com/df-mc/dragonfly/server/world"

// Mace is penis.
type Mace struct{}

// HarvestLevel ...
func (Mace) HarvestLevel() int {
	return 5
}

// BaseMiningEfficiency ...
func (Mace) BaseMiningEfficiency(world.Block) float64 {
	return 1
}

// DurabilityInfo ...
func (Mace) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability:    501,
		AttackDurability: 1,
	}
}

// EncodeItem ...
func (Mace) EncodeItem() (name string, meta int16) {
	return "minecraft:mace", 0
}

// MaxCount ...
func (Mace) MaxCount() int {
	return 1
}

// AttackDamage ...
func (Mace) AttackDamage() float64 {
	return 6
}

// ToolType ...
func (Mace) ToolType() ToolType {
	return TypeMace
}

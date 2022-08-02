package block

import "github.com/df-mc/dragonfly/server/item"

// DeepslateBricks are a brick variant of deepslate and can spawn in ancient cities.
type DeepslateBricks struct {
	solid
	bassDrum

	// Cracked specifies if the deepslate bricks is its cracked variant.
	Cracked bool
}

// BreakInfo ...
func (d DeepslateBricks) BreakInfo() BreakInfo {
	return newBreakInfo(3.5, pickaxeHarvestable, pickaxeEffective, oneOf(d)).withBlastResistance(18)
}

// SmeltInfo ...
func (d DeepslateBricks) SmeltInfo() item.SmeltInfo {
	if d.Cracked {
		return item.SmeltInfo{}
	}
	return newSmeltInfo(item.NewStack(DeepslateBricks{Cracked: true}, 1), 0.1)
}

// EncodeItem ...
func (d DeepslateBricks) EncodeItem() (name string, meta int16) {
	if d.Cracked {
		return "minecraft:cracked_deepslate_bricks", 0
	}
	return "minecraft:deepslate_bricks", 0
}

// EncodeBlock ...
func (d DeepslateBricks) EncodeBlock() (string, map[string]any) {
	if d.Cracked {
		return "minecraft:cracked_deepslate_bricks", nil
	}
	return "minecraft:deepslate_bricks", nil
}

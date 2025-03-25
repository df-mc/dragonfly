package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand/v2"
)

// RedstoneOre is a common ore.
type RedstoneOre struct {
	solid
	bassDrum

	// Type is the type of redstone ore.
	Type OreType
	// Lit returns if the redstone ore is lit.
	Lit bool
}

// BreakInfo ...
func (r RedstoneOre) BreakInfo() BreakInfo {
	i := newBreakInfo(r.Type.Hardness(), func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierIron.HarvestLevel
	}, pickaxeEffective, silkTouchOneOf(RedstoneWire{}, r)).withXPDropRange(1, 5)
	return i
}

// Activate ...
func (r RedstoneOre) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, _ item.User, _ *item.UseContext) bool {
	if !r.Lit {
		r.Lit = true
		tx.SetBlock(pos, r, nil)
	}
	return false
}

// RandomTick ...
func (r RedstoneOre) RandomTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if r.Lit {
		r.Lit = false
		tx.SetBlock(pos, r, nil)
	}
}

// LightEmissionLevel ...
func (r RedstoneOre) LightEmissionLevel() uint8 {
	if r.Lit {
		return 9
	}
	return 0
}

// SmeltInfo ...
func (RedstoneOre) SmeltInfo() item.SmeltInfo {
	return newOreSmeltInfo(item.NewStack(RedstoneWire{}, 1), 0.7)
}

// EncodeItem ...
func (r RedstoneOre) EncodeItem() (name string, meta int16) {
	if r.Lit {
		return "minecraft:lit_" + r.Type.Prefix() + "redstone_ore", 0
	}
	return "minecraft:" + r.Type.Prefix() + "redstone_ore", 0
}

// EncodeBlock ...
func (r RedstoneOre) EncodeBlock() (string, map[string]any) {
	if r.Lit {
		return "minecraft:lit_" + r.Type.Prefix() + "redstone_ore", nil
	}
	return "minecraft:" + r.Type.Prefix() + "redstone_ore", nil
}

package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
)

// Magma is a light-emitting Nether block that damages entities standing on it.
type Magma struct {
	solid
	bassDrum
}

// LightEmissionLevel ...
func (Magma) LightEmissionLevel() uint8 {
	return 3
}

// EntityStepOn ...
func (Magma) EntityStepOn(_ cube.Pos, _ *world.Tx, e world.Entity) {
	if fireProof, ok := e.(interface{ FireProof() bool }); ok && fireProof.FireProof() {
		return
	}
	// TODO: Check for Frost Walker once the enchantment is implemented in Dragonfly.
	if sneaking, ok := e.(interface{ Sneaking() bool }); ok && sneaking.Sneaking() {
		return
	}
	if l, ok := e.(livingEntity); ok {
		l.Hurt(1, MagmaDamageSource{})
	}
}

// BreakInfo ...
func (m Magma) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, pickaxeHarvestable, pickaxeEffective, oneOf(m)).withBlastResistance(30)
}

// EncodeItem ...
func (Magma) EncodeItem() (name string, meta int16) {
	return "minecraft:magma", 0
}

// EncodeBlock ...
func (Magma) EncodeBlock() (string, map[string]any) {
	return "minecraft:magma", nil
}

// MagmaDamageSource is used for damage caused by standing on a magma block.
type MagmaDamageSource struct{}

func (MagmaDamageSource) ReducedByResistance() bool { return true }
func (MagmaDamageSource) ReducedByArmour() bool     { return true }
func (MagmaDamageSource) Fire() bool                { return true }
func (MagmaDamageSource) AffectedByEnchantment(e item.EnchantmentType) bool {
	return e == enchantment.FireProtection
}
func (MagmaDamageSource) IgnoreTotem() bool { return false }

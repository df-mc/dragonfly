package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"math"
)

// AffectedDamageSource represents a world.DamageSource whose damage may be
// affected by an enchantment. A world.DamageSource does not need to implement
// AffectedDamageSource to let Protection affect the damage. This happens
// depending on the (world.DamageSource).ReducedByResistance() method.
type AffectedDamageSource interface {
	world.DamageSource
	// AffectedByEnchantment specifies if a world.DamageSource is affected by
	// the item.EnchantmentType passed.
	AffectedByEnchantment(e item.EnchantmentType) bool
}

// DamageModifier is an item.EnchantmentType that can reduce damage through a
// modifier if an AffectedDamageSource returns true for it.
type DamageModifier interface {
	Modifier() float64
}

// ProtectionFactor calculates the combined protection factor for a slice of
// item.Enchantment. The factor depends on the world.DamageSource passed and is
// in a range of [0, 0.8], where 0.8 means incoming damage would be reduced by
// 80%.
func ProtectionFactor(src world.DamageSource, enchantments []item.Enchantment) float64 {
	f := 0.0
	for _, e := range enchantments {
		t := e.Type()
		modifier, ok := t.(DamageModifier)
		if !ok {
			continue
		}
		reduced := false
		if _, ok := t.(Protection); ok && src.ReducedByResistance() {
			// Special case for Protection, because it applies to all damage
			// sources by default, except those not reduced by resistance.
			reduced = true
		} else if asrc, ok := src.(AffectedDamageSource); ok && asrc.AffectedByEnchantment(t) {
			reduced = true
		}

		if reduced {
			f += float64(e.Level()) * modifier.Modifier()
		}
	}
	return math.Min(f, 0.8)
}

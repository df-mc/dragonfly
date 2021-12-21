package item

import (
	"github.com/df-mc/dragonfly/server/item/armour"
	"github.com/df-mc/dragonfly/server/world"
)

// Helmet is a defensive item that may be worn in the head slot. It comes in several tiers, each with
// different defence points and armour toughness.
type Helmet struct {
	// Tier is the tier of the armour.
	Tier armour.Tier
}

// Use handles the using of a helmet to auto-equip it in an armour slot.
func (h Helmet) Use(_ *world.World, user User, ctx *UseContext) bool {
	ctx.SwapHeldWithArmour(0)
	return false
}

// MaxCount always returns 1.
func (h Helmet) MaxCount() int {
	return 1
}

// DefencePoints ...
func (h Helmet) DefencePoints() float64 {
	switch h.Tier {
	case armour.TierLeather:
		return 1
	case armour.TierGold, armour.TierChain, armour.TierIron:
		return 2
	case armour.TierDiamond, armour.TierNetherite:
		return 3
	}
	panic("invalid helmet tier")
}

// KnockBackResistance ...
func (h Helmet) KnockBackResistance() float64 {
	return h.Tier.KnockBackResistance
}

// DurabilityInfo ...
func (h Helmet) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: int(h.Tier.BaseDurability),
		BrokenItem:    simpleItem(Stack{}),
	}
}

// Helmet ...
func (h Helmet) Helmet() bool {
	return true
}

// EncodeItem ...
func (h Helmet) EncodeItem() (name string, meta int16) {
	return "minecraft:" + h.Tier.Name + "_helmet", 0
}

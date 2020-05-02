package item

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/armour"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
)

// Helmet is a defensive item that may be worn in the head slot. It comes in several tiers, each with
// different defence points and armour toughness.
type Helmet struct {
	// Tier is the tier of the armour.
	Tier armour.Tier
}

// Use handles the using of a helmet to auto-equip it in an armour slot.
func (h Helmet) Use(_ *world.World, user User, _ *UseContext) bool {
	if armoured, ok := user.(Armoured); ok {
		currentEquipped := armoured.Armour().Helmet()

		right, left := user.HeldItems()
		armoured.Armour().SetHelmet(right)
		user.SetHeldItems(currentEquipped, left)
	}
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
	case armour.TierGold:
		return 2
	case armour.TierChain:
		return 2
	case armour.TierIron:
		return 2
	case armour.TierDiamond:
		return 3
	}
	panic("invalid helmet tier")
}

// DurabilityInfo ...
func (h Helmet) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: int(h.Tier.BaseDurability),
		BrokenItem:    simpleItem(Stack{}),
	}
}

// EncodeItem ...
func (h Helmet) EncodeItem() (id int32, meta int16) {
	switch h.Tier {
	case armour.TierLeather:
		return 298, 0
	case armour.TierGold:
		return 314, 0
	case armour.TierChain:
		return 302, 0
	case armour.TierIron:
		return 306, 0
	case armour.TierDiamond:
		return 310, 0
	}
	panic("invalid helmet tier")
}

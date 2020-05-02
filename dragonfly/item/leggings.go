package item

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/armour"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
)

// Leggings are a defensive item that may be equipped in the leggings armour slot. It comes in multiple tiers
// like helmets, chestplates and boots.
type Leggings struct {
	// Tier is the tier of the leggings.
	Tier armour.Tier
}

// Use handles the auto-equipping of leggings in an armour slot by using the item.
func (l Leggings) Use(_ *world.World, user User, _ *UseContext) bool {
	if armoured, ok := user.(Armoured); ok {
		currentEquipped := armoured.Armour().Leggings()

		right, left := user.HeldItems()
		armoured.Armour().SetLeggings(right)
		user.SetHeldItems(currentEquipped, left)
	}
	return false
}

// MaxCount always returns 1.
func (l Leggings) MaxCount() int {
	return 1
}

// DefencePoints ...
func (l Leggings) DefencePoints() float64 {
	switch l.Tier {
	case armour.TierLeather:
		return 2
	case armour.TierGold:
		return 3
	case armour.TierChain:
		return 4
	case armour.TierIron:
		return 5
	case armour.TierDiamond:
		return 6
	}
	panic("invalid leggings tier")
}

// DurabilityInfo ...
func (l Leggings) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: int(l.Tier.BaseDurability + l.Tier.BaseDurability/2.5),
		BrokenItem:    simpleItem(Stack{}),
	}
}

// EncodeItem ...
func (l Leggings) EncodeItem() (id int32, meta int16) {
	switch l.Tier {
	case armour.TierLeather:
		return 300, 0
	case armour.TierGold:
		return 316, 0
	case armour.TierChain:
		return 304, 0
	case armour.TierIron:
		return 308, 0
	case armour.TierDiamond:
		return 312, 0
	}
	panic("invalid leggings tier")
}

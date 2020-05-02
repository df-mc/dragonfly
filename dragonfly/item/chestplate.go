package item

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/armour"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
)

// Chestplate is a defensive item that may be equipped in the chestplate slot. Generally, chestplates provide
// the most defence of all armour items.
type Chestplate struct {
	// Tier is the tier of the chestplate.
	Tier armour.Tier
}

// Use handles the using of a chestplate to auto-equip it in the designated armour slot.
func (c Chestplate) Use(_ *world.World, user User, _ *UseContext) bool {
	if armoured, ok := user.(Armoured); ok {
		currentEquipped := armoured.Armour().Chestplate()

		right, left := user.HeldItems()
		armoured.Armour().SetChestplate(right)
		user.SetHeldItems(currentEquipped, left)
	}
	return false
}

// MaxCount always returns 1.
func (c Chestplate) MaxCount() int {
	return 1
}

// DefencePoints ...
func (c Chestplate) DefencePoints() float64 {
	switch c.Tier {
	case armour.TierLeather:
		return 3
	case armour.TierGold:
		return 5
	case armour.TierChain:
		return 5
	case armour.TierIron:
		return 6
	case armour.TierDiamond:
		return 8
	}
	panic("invalid chestplate tier")
}

// DurabilityInfo ...
func (c Chestplate) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: int(c.Tier.BaseDurability + c.Tier.BaseDurability/2.2),
		BrokenItem:    simpleItem(Stack{}),
	}
}

// EncodeItem ...
func (c Chestplate) EncodeItem() (id int32, meta int16) {
	switch c.Tier {
	case armour.TierLeather:
		return 299, 0
	case armour.TierGold:
		return 315, 0
	case armour.TierChain:
		return 303, 0
	case armour.TierIron:
		return 307, 0
	case armour.TierDiamond:
		return 311, 0
	}
	panic("invalid chestplate tier")
}

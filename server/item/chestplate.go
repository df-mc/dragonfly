package item

import (
	"github.com/df-mc/dragonfly/server/item/armour"
	"github.com/df-mc/dragonfly/server/world"
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
	case armour.TierGold, armour.TierChain:
		return 5
	case armour.TierIron:
		return 6
	case armour.TierDiamond, armour.TierNetherite:
		return 8
	}
	panic("invalid chestplate tier")
}

// KnockBackResistance ...
func (c Chestplate) KnockBackResistance() float64 {
	return c.Tier.KnockBackResistance
}

// DurabilityInfo ...
func (c Chestplate) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: int(c.Tier.BaseDurability + c.Tier.BaseDurability/2.2),
		BrokenItem:    simpleItem(Stack{}),
	}
}

// EncodeItem ...
func (c Chestplate) EncodeItem() (name string, meta int16) {
	return "minecraft:" + c.Tier.Name + "_chestplate", 0
}

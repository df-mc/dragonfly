package item

import (
	"github.com/df-mc/dragonfly/dragonfly/item/armour"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// Leggings are a defensive item that may be equipped in the leggings armour slot. They come in several tiers,
// like leather, gold, chain, iron and diamond.
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
	case armour.TierDiamond, armour.TierNetherite:
		return 6
	}
	panic("invalid leggings tier")
}

// KnockBackResistance ...
func (l Leggings) KnockBackResistance() float64 {
	return l.Tier.KnockBackResistance
}

// DurabilityInfo ...
func (l Leggings) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: int(l.Tier.BaseDurability + l.Tier.BaseDurability/2.5),
		BrokenItem:    simpleItem(Stack{}),
	}
}

// EncodeItem ...
func (l Leggings) EncodeItem() (id int32, name string, meta int16) {
	name = "minecraft:" + l.Tier.Name + "_leggings"
	switch l.Tier {
	case armour.TierLeather:
		return 300, name, 0
	case armour.TierGold:
		return 316, name, 0
	case armour.TierChain:
		return 304, name, 0
	case armour.TierIron:
		return 308, name, 0
	case armour.TierDiamond:
		return 312, name, 0
	case armour.TierNetherite:
		return 750, name, 0
	}
	panic("invalid leggings tier")
}

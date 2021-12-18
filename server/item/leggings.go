package item

import (
	"github.com/df-mc/dragonfly/server/item/armour"
	"github.com/df-mc/dragonfly/server/world"
)

// Leggings are a defensive item that may be equipped in the leggings armour slot. They come in several tiers,
// like leather, gold, chain, iron and diamond.
type Leggings struct {
	// Tier is the tier of the leggings.
	Tier armour.Tier
}

// Use handles the auto-equipping of leggings in an armour slot by using the item.
func (l Leggings) Use(_ *world.World, user User, ctx *UseContext) bool {
	ctx.SwapHeldWithArmour(2)
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

// Leggings ...
func (l Leggings) Leggings() bool {
	return true
}

// DurabilityInfo ...
func (l Leggings) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: int(l.Tier.BaseDurability + l.Tier.BaseDurability/2.5),
		BrokenItem:    simpleItem(Stack{}),
	}
}

// EncodeItem ...
func (l Leggings) EncodeItem() (name string, meta int16) {
	return "minecraft:" + l.Tier.Name + "_leggings", 0
}

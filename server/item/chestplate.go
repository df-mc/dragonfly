package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Chestplate is a defensive item that may be equipped in the chestplate slot. Generally, chestplates provide
// the most defence of all armour items.
type Chestplate struct {
	// Tier is the tier of the chestplate.
	Tier ArmourTier
}

// Use handles the using of a chestplate to auto-equip it in the designated armour slot.
func (c Chestplate) Use(_ *world.World, _ User, ctx *UseContext) bool {
	ctx.SwapHeldWithArmour(1)
	return false
}

// MaxCount always returns 1.
func (c Chestplate) MaxCount() int {
	return 1
}

// DefencePoints ...
func (c Chestplate) DefencePoints() float64 {
	switch c.Tier.(type) {
	case ArmourTierLeather:
		return 3
	case ArmourTierGold, ArmourTierChain:
		return 5
	case ArmourTierIron:
		return 6
	case ArmourTierDiamond, ArmourTierNetherite:
		return 8
	}
	panic("invalid chestplate tier")
}

// Toughness ...
func (c Chestplate) Toughness() float64 {
	return c.Tier.Toughness()
}

// KnockBackResistance ...
func (c Chestplate) KnockBackResistance() float64 {
	return c.Tier.KnockBackResistance()
}

// DurabilityInfo ...
func (c Chestplate) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: int(c.Tier.BaseDurability() + c.Tier.BaseDurability()/2.2),
		BrokenItem:    simpleItem(Stack{}),
	}
}

// RepairableBy ...
func (c Chestplate) RepairableBy(i Stack) bool {
	return armourTierRepairable(c.Tier)(i)
}

// Chestplate ...
func (c Chestplate) Chestplate() bool {
	return true
}

// EncodeItem ...
func (c Chestplate) EncodeItem() (name string, meta int16) {
	return "minecraft:" + c.Tier.Name() + "_chestplate", 0
}

// DecodeNBT ...
func (c Chestplate) DecodeNBT(data map[string]any) any {
	if t, ok := c.Tier.(ArmourTierLeather); ok {
		if v, ok := data["customColor"].(int32); ok {
			t.Colour = rgbaFromInt32(v)
			c.Tier = t
		}
	}
	return c
}

// EncodeNBT ...
func (c Chestplate) EncodeNBT() map[string]any {
	if t, ok := c.Tier.(ArmourTierLeather); ok && t.Colour != (color.RGBA{}) {
		return map[string]any{"customColor": int32FromRGBA(t.Colour)}
	}
	return nil
}

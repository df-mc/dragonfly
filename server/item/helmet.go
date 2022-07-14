package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Helmet is a defensive item that may be worn in the head slot. It comes in several tiers, each with
// different defence points and armour toughness.
type Helmet struct {
	// Tier is the tier of the armour.
	Tier ArmourTier
}

// Use handles the using of a helmet to auto-equip it in an armour slot.
func (h Helmet) Use(_ *world.World, _ User, ctx *UseContext) bool {
	ctx.SwapHeldWithArmour(0)
	return false
}

// MaxCount always returns 1.
func (h Helmet) MaxCount() int {
	return 1
}

// DefencePoints ...
func (h Helmet) DefencePoints() float64 {
	switch h.Tier.(type) {
	case ArmourTierLeather:
		return 1
	case ArmourTierGold, ArmourTierChain, ArmourTierIron:
		return 2
	case ArmourTierDiamond, ArmourTierNetherite:
		return 3
	}
	panic("invalid helmet tier")
}

// KnockBackResistance ...
func (h Helmet) KnockBackResistance() float64 {
	return h.Tier.KnockBackResistance()
}

// Toughness ...
func (h Helmet) Toughness() float64 {
	return h.Tier.Toughness()
}

// DurabilityInfo ...
func (h Helmet) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: int(h.Tier.BaseDurability()),
		BrokenItem:    simpleItem(Stack{}),
	}
}

// RepairableBy ...
func (h Helmet) RepairableBy(i Stack) bool {
	return armourTierRepairable(h.Tier)(i)
}

// Helmet ...
func (h Helmet) Helmet() bool {
	return true
}

// EncodeItem ...
func (h Helmet) EncodeItem() (name string, meta int16) {
	return "minecraft:" + h.Tier.Name() + "_helmet", 0
}

// DecodeNBT ...
func (h Helmet) DecodeNBT(data map[string]any) any {
	if t, ok := h.Tier.(ArmourTierLeather); ok {
		if v, ok := data["customColor"].(int32); ok {
			t.Colour = rgbaFromInt32(v)
			h.Tier = t
		}
	}
	return h
}

// EncodeNBT ...
func (h Helmet) EncodeNBT() map[string]any {
	if t, ok := h.Tier.(ArmourTierLeather); ok && t.Colour != (color.RGBA{}) {
		return map[string]any{"customColor": int32FromRGBA(t.Colour)}
	}
	return nil
}

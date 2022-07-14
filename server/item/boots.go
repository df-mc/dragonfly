package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Boots are a defensive item that may be equipped in the boots armour slot. They come in several tiers, like
// leather, gold, chain, iron and diamond.
type Boots struct {
	// Tier is the tier of the boots.
	Tier ArmourTier
}

// Use handles the auto-equipping of boots in the armour slot when using it.
func (b Boots) Use(_ *world.World, _ User, ctx *UseContext) bool {
	ctx.SwapHeldWithArmour(3)
	return false
}

// MaxCount always returns 1.
func (b Boots) MaxCount() int {
	return 1
}

// DurabilityInfo ...
func (b Boots) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: int(b.Tier.BaseDurability() + b.Tier.BaseDurability()/5.5),
		BrokenItem:    simpleItem(Stack{}),
	}
}

// RepairableBy ...
func (b Boots) RepairableBy(i Stack) bool {
	return armourTierRepairable(b.Tier)(i)
}

// DefencePoints ...
func (b Boots) DefencePoints() float64 {
	switch b.Tier.(type) {
	case ArmourTierLeather, ArmourTierGold, ArmourTierChain:
		return 1
	case ArmourTierIron:
		return 2
	case ArmourTierDiamond, ArmourTierNetherite:
		return 3
	}
	panic("invalid boots tier")
}

// Toughness ...
func (b Boots) Toughness() float64 {
	return b.Tier.Toughness()
}

// KnockBackResistance ...
func (b Boots) KnockBackResistance() float64 {
	return b.Tier.KnockBackResistance()
}

// Boots ...
func (b Boots) Boots() bool {
	return true
}

// EncodeItem ...
func (b Boots) EncodeItem() (name string, meta int16) {
	return "minecraft:" + b.Tier.Name() + "_boots", 0
}

// DecodeNBT ...
func (b Boots) DecodeNBT(data map[string]any) any {
	if t, ok := b.Tier.(ArmourTierLeather); ok {
		if v, ok := data["customColor"].(int32); ok {
			t.Colour = rgbaFromInt32(v)
			b.Tier = t
		}
	}
	return b
}

// EncodeNBT ...
func (b Boots) EncodeNBT() map[string]any {
	if t, ok := b.Tier.(ArmourTierLeather); ok && t.Colour != (color.RGBA{}) {
		return map[string]any{"customColor": int32FromRGBA(t.Colour)}
	}
	return nil
}

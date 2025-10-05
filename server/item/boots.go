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
	// Trim specifies the trim of the armour.
	Trim ArmourTrim
}

// Use handles the auto-equipping of boots in the armour slot when using it.
func (b Boots) Use(_ *world.Tx, _ User, ctx *UseContext) bool {
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

// SmeltInfo ...
func (b Boots) SmeltInfo() SmeltInfo {
	switch b.Tier.(type) {
	case ArmourTierIron, ArmourTierChain:
		return newOreSmeltInfo(NewStack(IronNugget{}, 1), 0.1)
	case ArmourTierGold:
		return newOreSmeltInfo(NewStack(GoldNugget{}, 1), 0.1)
	case ArmourTierCopper:
		return newOreSmeltInfo(NewStack(CopperNugget{}, 1), 0.1)
	}
	return SmeltInfo{}
}

// RepairableBy ...
func (b Boots) RepairableBy(i Stack) bool {
	return armourTierRepairable(b.Tier)(i)
}

// DefencePoints ...
func (b Boots) DefencePoints() float64 {
	switch b.Tier.Name() {
	case "leather", "golden", "chainmail":
		return 1
	case "iron":
		return 2
	case "diamond", "netherite":
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

// EnchantmentValue ...
func (b Boots) EnchantmentValue() int {
	return b.Tier.EnchantmentValue()
}

// Boots ...
func (b Boots) Boots() bool {
	return true
}

// WithTrim ...
func (b Boots) WithTrim(trim ArmourTrim) world.Item {
	b.Trim = trim
	return b
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
	b.Trim = readTrim(data)
	return b
}

// EncodeNBT ...
func (b Boots) EncodeNBT() map[string]any {
	m := map[string]any{}
	if t, ok := b.Tier.(ArmourTierLeather); ok && t.Colour != (color.RGBA{}) {
		m["customColor"] = int32FromRGBA(t.Colour)
	}
	writeTrim(m, b.Trim)
	return m
}

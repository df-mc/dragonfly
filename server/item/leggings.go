package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Leggings are a defensive item that may be equipped in the leggings armour slot. They come in several tiers,
// like leather, gold, chain, iron and diamond.
type Leggings struct {
	// Tier is the tier of the leggings.
	Tier ArmourTier
	// Trim specifies the trim of the armour.
	Trim ArmourTrim
}

// Use handles the auto-equipping of leggings in an armour slot by using the item.
func (l Leggings) Use(_ *world.Tx, _ User, ctx *UseContext) bool {
	ctx.SwapHeldWithArmour(2)
	return false
}

// MaxCount always returns 1.
func (l Leggings) MaxCount() int {
	return 1
}

// DefencePoints ...
func (l Leggings) DefencePoints() float64 {
	switch l.Tier.Name() {
	case "leather":
		return 2
	case "copper", "golden":
		return 3
	case "chainmail":
		return 4
	case "iron":
		return 5
	case "diamond", "netherite":
		return 6
	}
	panic("invalid leggings tier")
}

// Toughness ...
func (l Leggings) Toughness() float64 {
	return l.Tier.Toughness()
}

// KnockBackResistance ...
func (l Leggings) KnockBackResistance() float64 {
	return l.Tier.KnockBackResistance()
}

// EnchantmentValue ...
func (l Leggings) EnchantmentValue() int {
	return l.Tier.EnchantmentValue()
}

// Leggings ...
func (l Leggings) Leggings() bool {
	return true
}

// DurabilityInfo ...
func (l Leggings) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: int(l.Tier.BaseDurability() + l.Tier.BaseDurability()/2.5),
		BrokenItem:    simpleItem(Stack{}),
	}
}

// RepairableBy ...
func (l Leggings) RepairableBy(i Stack) bool {
	return armourTierRepairable(l.Tier)(i)
}

// SmeltInfo ...
func (l Leggings) SmeltInfo() SmeltInfo {
	switch l.Tier.(type) {
	case ArmourTierIron, ArmourTierChain:
		return newOreSmeltInfo(NewStack(IronNugget{}, 1), 0.1)
	case ArmourTierGold:
		return newOreSmeltInfo(NewStack(GoldNugget{}, 1), 0.1)
	case ArmourTierCopper:
		return newOreSmeltInfo(NewStack(CopperNugget{}, 1), 0.1)
	}
	return SmeltInfo{}
}

// WithTrim ...
func (l Leggings) WithTrim(trim ArmourTrim) world.Item {
	l.Trim = trim
	return l
}

// EncodeItem ...
func (l Leggings) EncodeItem() (name string, meta int16) {
	return "minecraft:" + l.Tier.Name() + "_leggings", 0
}

// DecodeNBT ...
func (l Leggings) DecodeNBT(data map[string]any) any {
	if t, ok := l.Tier.(ArmourTierLeather); ok {
		if v, ok := data["customColor"].(int32); ok {
			t.Colour = rgbaFromInt32(v)
			l.Tier = t
		}
	}
	l.Trim = readTrim(data)
	return l
}

// EncodeNBT ...
func (l Leggings) EncodeNBT() map[string]any {
	m := map[string]any{}
	if t, ok := l.Tier.(ArmourTierLeather); ok && t.Colour != (color.RGBA{}) {
		m["customColor"] = int32FromRGBA(t.Colour)
	}
	writeTrim(m, l.Trim)
	return m
}

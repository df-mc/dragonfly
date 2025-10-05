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
	// Trim specifies the trim of the armour.
	Trim ArmourTrim
}

// Use handles the using of a helmet to auto-equip it in an armour slot.
func (h Helmet) Use(_ *world.Tx, _ User, ctx *UseContext) bool {
	ctx.SwapHeldWithArmour(0)
	return false
}

// MaxCount always returns 1.
func (h Helmet) MaxCount() int {
	return 1
}

// DefencePoints ...
func (h Helmet) DefencePoints() float64 {
	switch h.Tier.Name() {
	case "leather":
		return 1
	case "golden", "chainmail", "iron":
		return 2
	case "diamond", "netherite":
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

// EnchantmentValue ...
func (h Helmet) EnchantmentValue() int {
	return h.Tier.EnchantmentValue()
}

// DurabilityInfo ...
func (h Helmet) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: int(h.Tier.BaseDurability()),
		BrokenItem:    simpleItem(Stack{}),
	}
}

// SmeltInfo ...
func (h Helmet) SmeltInfo() SmeltInfo {
	switch h.Tier.(type) {
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
func (h Helmet) RepairableBy(i Stack) bool {
	return armourTierRepairable(h.Tier)(i)
}

// Helmet ...
func (h Helmet) Helmet() bool {
	return true
}

// WithTrim ...
func (h Helmet) WithTrim(trim ArmourTrim) world.Item {
	h.Trim = trim
	return h
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
	h.Trim = readTrim(data)
	return h
}

// EncodeNBT ...
func (h Helmet) EncodeNBT() map[string]any {
	m := map[string]any{}
	if t, ok := h.Tier.(ArmourTierLeather); ok && t.Colour != (color.RGBA{}) {
		m["customColor"] = int32FromRGBA(t.Colour)
	}
	writeTrim(m, h.Trim)
	return m
}

func readTrim(m map[string]any) ArmourTrim {
	if trim, ok := m["Trim"].(map[string]any); ok {
		material, _ := trim["Material"].(string)
		pattern, _ := trim["Pattern"].(string)
		template, ok := smithingTemplateFromString(pattern)
		trimMaterial, ok2 := trimMaterialFromString(material)
		if ok && ok2 {
			return ArmourTrim{Template: template, Material: trimMaterial}
		}
	}
	return ArmourTrim{}
}

func writeTrim(m map[string]any, t ArmourTrim) {
	if !t.Zero() {
		m["Trim"] = map[string]any{
			"Material": t.Material.TrimMaterial(),
			"Pattern":  t.Template.String(),
		}
	}
}

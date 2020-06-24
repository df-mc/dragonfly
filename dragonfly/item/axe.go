package item

import (
	"github.com/df-mc/dragonfly/dragonfly/internal/item_internal"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Axe is a tool generally used for mining wood-like blocks. It may also be used to break some plant-like
// blocks at a faster pace such as pumpkins.
type Axe struct {
	// Tier is the tier of the axe.
	Tier tool.Tier
}

// UseOnBlock handles the stripping of logs when a player clicks a log with an axe.
func (a Axe) UseOnBlock(pos world.BlockPos, _ world.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	if b := w.Block(pos); item_internal.IsUnstrippedLog(b) {
		strippedLog := item_internal.StripLog(b)
		w.SetBlock(pos, strippedLog)
		w.PlaySound(pos.Vec3(), sound.ItemUseOn{Block: strippedLog})

		ctx.DamageItem(1)
		return true
	}
	return false
}

// MaxCount always returns 1.
func (a Axe) MaxCount() int {
	return 1
}

// DurabilityInfo ...
func (a Axe) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability:    a.Tier.Durability,
		BrokenItem:       simpleItem(Stack{}),
		AttackDurability: 2,
		BreakDurability:  1,
	}
}

// AttackDamage ...
func (a Axe) AttackDamage() float64 {
	return a.Tier.BaseAttackDamage + 2
}

// ToolType ...
func (a Axe) ToolType() tool.Type {
	return tool.TypeAxe
}

// HarvestLevel ...
func (a Axe) HarvestLevel() int {
	return a.Tier.HarvestLevel
}

// BaseMiningEfficiency ...
func (a Axe) BaseMiningEfficiency(world.Block) float64 {
	return a.Tier.BaseMiningEfficiency
}

// EncodeItem ...
func (a Axe) EncodeItem() (id int32, meta int16) {
	switch a.Tier {
	case tool.TierWood:
		return 271, 0
	case tool.TierGold:
		return 286, 0
	case tool.TierStone:
		return 275, 0
	case tool.TierIron:
		return 258, 0
	case tool.TierDiamond:
		return 279, 0
	}
	panic("invalid axe tier")
}

package item

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/internal/item_internal"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/tool"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Shovel is a tool generally used for mining ground-like blocks, such as sand, gravel and dirt. Additionally,
// shovels may be used to turn grass into grass paths.
type Shovel struct {
	// Tier is the tier of the shovel.
	Tier tool.Tier
}

// UseOnBlock handles the creation of grass path blocks from grass blocks.
func (s Shovel) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	if grass := w.Block(pos); grass == item_internal.Grass {
		if face == world.Down {
			// Grass paths are not created when the bottom face is clicked.
			return false
		}
		if w.Block(pos.Add(world.BlockPos{0, 1})) != item_internal.Air {
			// Grass paths can only be created if air is above the grass block.
			return false
		}
		w.SetBlock(pos, item_internal.GrassPath)
		w.PlaySound(pos.Vec3(), sound.ItemUseOn{Block: item_internal.GrassPath})

		ctx.DamageItem(1)
		return true
	}
	return false
}

// MaxCount always returns 1.
func (s Shovel) MaxCount() int {
	return 1
}

// AttackDamage returns the attack damage of the shovel.
func (s Shovel) AttackDamage() float64 {
	return s.Tier.BaseAttackDamage
}

// ToolType returns the tool type for shovels.
func (s Shovel) ToolType() tool.Type {
	return tool.TypeShovel
}

// HarvestLevel ...
func (s Shovel) HarvestLevel() int {
	return s.Tier.HarvestLevel
}

// BaseMiningEfficiency ...
func (s Shovel) BaseMiningEfficiency(world.Block) float64 {
	return s.Tier.BaseMiningEfficiency
}

// DurabilityInfo ...
func (s Shovel) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability:    s.Tier.Durability,
		BrokenItem:       simpleItem(Stack{}),
		AttackDurability: 2,
		BreakDurability:  1,
	}
}

// EncodeItem ...
func (s Shovel) EncodeItem() (id int32, meta int16) {
	switch s.Tier {
	case tool.TierWood:
		return 269, 0
	case tool.TierGold:
		return 284, 0
	case tool.TierStone:
		return 273, 0
	case tool.TierIron:
		return 256, 0
	case tool.TierDiamond:
		return 277, 0
	}
	panic("invalid shovel tier")
}

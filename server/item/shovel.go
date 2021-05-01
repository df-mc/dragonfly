package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/item_internal"
	"github.com/df-mc/dragonfly/server/item/tool"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Shovel is a tool generally used for mining ground-like blocks, such as sand, gravel and dirt. Additionally,
// shovels may be used to turn grass into grass paths.
type Shovel struct {
	// Tier is the tier of the shovel.
	Tier tool.Tier
}

// UseOnBlock handles the creation of grass path blocks from grass blocks.
func (s Shovel) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	if b, ok := w.Block(pos).(shovellable); ok {
		if res, ok := b.Shovel(); ok {
			if face == cube.FaceDown {
				// Grass paths are not created when the bottom face is clicked.
				return false
			}
			if w.Block(pos.Add(cube.Pos{0, 1})) != item_internal.Air {
				// Grass paths can only be created if air is above the grass block.
				return false
			}
			w.PlaceBlock(pos, res)
			w.PlaySound(pos.Vec3(), sound.ItemUseOn{Block: res})

			ctx.DamageItem(1)
			return true
		}
	}
	return false
}

// shovellable represents a block that can be changed by using a shovel on it.
type shovellable interface {
	// Shovel returns a block that results from using a shovel on it, or false if it could not be changed using
	// a shovel.
	Shovel() (world.Block, bool)
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
func (s Shovel) EncodeItem() (id int32, name string, meta int16) {
	name = "minecraft:" + s.Tier.Name + "_shovel"
	switch s.Tier {
	case tool.TierWood:
		return 269, name, 0
	case tool.TierGold:
		return 284, name, 0
	case tool.TierStone:
		return 273, name, 0
	case tool.TierIron:
		return 256, name, 0
	case tool.TierDiamond:
		return 277, name, 0
	case tool.TierNetherite:
		return 744, name, 0
	}
	panic("invalid shovel tier")
}

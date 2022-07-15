package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Shovel is a tool generally used for mining ground-like blocks, such as sand, gravel and dirt. Additionally,
// shovels may be used to turn grass into dirt paths.
type Shovel struct {
	// Tier is the tier of the shovel.
	Tier ToolTier
}

// UseOnBlock handles the creation of dirt path blocks from dirt or grass blocks.
func (s Shovel) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, _ User, ctx *UseContext) bool {
	if b, ok := w.Block(pos).(shovellable); ok {
		if res, ok := b.Shovel(); ok {
			if face == cube.FaceDown {
				// Dirt paths are not created when the bottom face is clicked.
				return false
			}
			if w.Block(pos.Add(cube.Pos{0, 1})) != air() {
				// Dirt paths can only be created if air is above the grass block.
				return false
			}
			w.SetBlock(pos, res, nil)
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

// AttackDamage returns the attack damage to the shovel.
func (s Shovel) AttackDamage() float64 {
	return s.Tier.BaseAttackDamage
}

// ToolType returns the tool type for shovels.
func (s Shovel) ToolType() ToolType {
	return TypeShovel
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

// RepairableBy ...
func (s Shovel) RepairableBy(i Stack) bool {
	return toolTierRepairable(s.Tier)(i)
}

// EncodeItem ...
func (s Shovel) EncodeItem() (name string, meta int16) {
	return "minecraft:" + s.Tier.Name + "_shovel", 0
}

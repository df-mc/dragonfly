package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
)

// PowderSnow is loose snow that players sink into unless wearing leather boots.
type PowderSnow struct {
	empty
}

// PowderSnow identifies the block for bucket collection without an import cycle.
func (PowderSnow) PowderSnow() {}

// EntityInside cushions falls and extinguishes burning entities, melting the snow.
func (PowderSnow) EntityInside(_ cube.Pos, tx *world.Tx, e world.Entity) {
	if fall, ok := e.(fallDistanceEntity); ok {
		fall.ResetFallDistance()
	}
	if burning, ok := e.(flammableEntity); ok && burning.OnFireDuration() > 0 {
		box := e.H().Type().BBox(e).Translate(e.Position()).Grow(-0.0001)
		low, high := cube.PosFromVec3(box.Min()), cube.PosFromVec3(box.Max())
		for x := low[0]; x <= high[0]; x++ {
			for y := low[1]; y <= high[1]; y++ {
				for z := low[2]; z <= high[2]; z++ {
					p := cube.Pos{x, y, z}
					if _, snow := tx.Block(p).(PowderSnow); snow {
						tx.SetBlock(p, nil, nil)
					}
				}
			}
		}
		burning.Extinguish()
	}
}

// EntityBBox allows leather boots to support an entity on top of the snow.
// Sneaking lets the entity descend through it.
func (PowderSnow) EntityBBox(pos cube.Pos, _ *world.Tx, e world.Entity) []cube.BBox {
	if e.Position()[1] < float64(pos[1]+1)-0.00000012 {
		return nil
	}
	if sneaker, ok := e.(interface{ Sneaking() bool }); ok && sneaker.Sneaking() {
		return nil
	}
	if falling, ok := e.(interface{ FallDistance() float64 }); ok && falling.FallDistance() > 2.5 {
		return []cube.BBox{cube.Box(0, 0, 0, 1, 0.9, 1)}
	}
	if e.H().Type().EncodeEntity() == "minecraft:falling_block" {
		return []cube.BBox{cube.Box(0, 0, 0, 1, 1, 1)}
	}
	wearer, ok := e.(interface{ Armour() *inventory.Armour })
	if !ok {
		return nil
	}
	boots, ok := wearer.Armour().Boots().Item().(item.Boots)
	if !ok {
		return nil
	}
	if _, leather := boots.Tier.(item.ArmourTierLeather); !leather {
		return nil
	}
	return []cube.BBox{cube.Box(0, 0, 0, 1, 1, 1)}
}

// EntityLand prevents damage when landing in powder snow.
func (PowderSnow) EntityLand(_ cube.Pos, _ *world.Tx, _ world.Entity, distance *float64) {
	*distance = 0
}

// Pick returns the bucket used to place powder snow.
func (PowderSnow) Pick() item.Stack {
	return item.NewStack(item.Bucket{Content: item.PowderSnowBucketContent()}, 1)
}

// BreakInfo ...
func (PowderSnow) BreakInfo() BreakInfo {
	return newBreakInfo(0.25, alwaysHarvestable, nothingEffective, simpleDrops())
}

// EncodeBlock ...
func (PowderSnow) EncodeBlock() (string, map[string]any) {
	return "minecraft:powder_snow", nil
}

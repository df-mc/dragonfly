package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Cobweb is a block that drastically slows down the movement of most entities touching it.
type Cobweb struct {
	empty
	transparent
	sourceWaterDisplacer
}

// SwordMiningEfficiency returns the mining efficiency used by swords when breaking a cobweb.
func (Cobweb) SwordMiningEfficiency() float64 {
	return 15
}

// EntityInside ...
func (Cobweb) EntityInside(_ cube.Pos, _ *world.Tx, e world.Entity) {
	if fallEntity, ok := e.(fallDistanceEntity); ok {
		fallEntity.ResetFallDistance()
	}
	if v, ok := e.(velocityEntity); ok {
		vel := v.Velocity()
		vel[0] *= 0.25
		vel[1] *= 0.05
		vel[2] *= 0.25
		v.SetVelocity(vel)
	}
}

// SideClosed ...
func (Cobweb) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// HasLiquidDrops ...
func (c Cobweb) HasLiquidDrops() bool {
	return true
}

// BreakInfo ...
func (c Cobweb) BreakInfo() BreakInfo {
	return newBreakInfo(4, alwaysHarvestable, func(t item.Tool) bool {
		return t.ToolType() == item.TypeShears || t.ToolType() == item.TypeSword
	}, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if t.ToolType() == item.TypeShears || hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(c, 1)}
		}
		if t.ToolType() == item.TypeSword {
			// TODO: Drop string once item.String is implemented.
			return nil
		}
		return nil
	}).withBlastResistance(4)
}

// EncodeItem ...
func (Cobweb) EncodeItem() (name string, meta int16) {
	return "minecraft:web", 0
}

// EncodeBlock ...
func (Cobweb) EncodeBlock() (string, map[string]any) {
	return "minecraft:web", nil
}

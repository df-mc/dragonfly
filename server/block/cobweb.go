package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Cobweb is a non-solid block that drastically slows entities passing through it. It is broken
// quickly with a sword or shears and drops string when broken without silk touch.
type Cobweb struct {
	empty
	transparent
}

// Cobweb is implemented because the item package needs to identify this block but cannot implement the block package.
func (Cobweb) Cobweb() {}

// EntityInside slows the entity's velocity and resets its fall distance while it is inside the cobweb.
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

// BreakInfo ...
func (c Cobweb) BreakInfo() BreakInfo {
	return newBreakInfo(4, alwaysHarvestable, func(t item.Tool) bool {
		return swordEffective(t) || shearsEffective(t)
	}, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if t.ToolType() == item.TypeShears || hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(c, 1)}
		}
		if t.ToolType() == item.TypeSword {
			return []item.Stack{item.NewStack(String{}, 1)}
		}
		return nil
	}).withBlastResistance(4)
}

// HasLiquidDrops ...
func (Cobweb) HasLiquidDrops() bool {
	return true
}

// EncodeItem ...
func (Cobweb) EncodeItem() (name string, meta int16) {
	return "minecraft:web", 0
}

// EncodeBlock ...
func (Cobweb) EncodeBlock() (string, map[string]any) {
	return "minecraft:web", nil
}

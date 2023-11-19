package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Cobweb is a block that can slow down entity movement and negate fall damage.
type Cobweb struct {
	empty
}

// BreakInfo ...
func (c Cobweb) BreakInfo() BreakInfo {
	return newBreakInfo(
		4,
		alwaysHarvestable,
		func(t item.Tool) bool {
			return swordEffective(t) || shearsEffective(t)
		},
		func(t item.Tool, e []item.Enchantment) []item.Stack {
			if t.ToolType() == item.TypeShears || (t.ToolType() == item.TypeSword && hasSilkTouch(e)) {
				return oneOf(c)(t, e)
			}
			return oneOf(String{})(t, e)
		},
	)
}

// EntityInside ...
func (Cobweb) EntityInside(_ cube.Pos, _ *world.World, e world.Entity) {
	if fallEntity, ok := e.(fallDistanceEntity); ok {
		fallEntity.ResetFallDistance()
	}
}

// EncodeItem ...
func (c Cobweb) EncodeItem() (name string, meta int16) {
	return "minecraft:web", 0
}

// EncodeBlock ...
func (c Cobweb) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:web", nil
}

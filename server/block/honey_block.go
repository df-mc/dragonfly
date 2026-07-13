package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// HoneyBlock is a sticky, translucent block crafted from honey bottles. It reduces the fall damage of
// entities that land on it.
//
// The collision box uses the standard full-size Solid model rather than a 1/16-block-shorter custom shape.
// A shorter box was tried to match the exact vanilla hitbox, but it broke checkOnGround()'s grounded-state
// detection: entities landing on the block never registered as on ground, so fall damage silently never
// triggered at all. A full block gives up a barely-visible hitbox difference in exchange for fall damage
// actually working, which matters far more for parity.
type HoneyBlock struct {
	solid
	transparent
}

// EntityLand reduces the fall damage dealt to the entity by 80%. The reduction is applied to the damage
// itself (i.e. the fall distance beyond the safe 3-block threshold), not to the raw fall distance, matching
// the wiki-documented example of a fall that would deal 10 damage dealing 2 damage instead.
func (HoneyBlock) EntityLand(_ cube.Pos, _ *world.Tx, e world.Entity, distance *float64) {
	if _, ok := e.(fallDistanceEntity); ok {
		*distance = (*distance-3)*0.2 + 3
	}
}

// BreakInfo ...
func (h HoneyBlock) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(h))
}

// EncodeItem ...
func (HoneyBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:honey_block", 0
}

// EncodeBlock ...
func (HoneyBlock) EncodeBlock() (string, map[string]any) {
	return "minecraft:honey_block", nil
}

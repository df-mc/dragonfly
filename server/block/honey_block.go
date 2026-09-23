package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// HoneyBlock is a sticky, translucent block crafted from honey bottles. Entities landing on it take
// reduced fall damage, while entities touching its sides slide down them slowly.
type HoneyBlock struct {
	transparent
}

// Model ...
func (HoneyBlock) Model() world.BlockModel {
	return model.Honey{}
}

// EntityLand reduces the fall damage dealt to the entity by 80%.
func (HoneyBlock) EntityLand(_ cube.Pos, _ *world.Tx, e world.Entity, distance *float64) {
	if _, ok := e.(fallDistanceEntity); ok {
		*distance = (*distance-3)*0.2 + 3
	}
}

// EntityInside resets the fall distance of an entity sliding down the side of the block, so that sliding
// down one deals no fall damage, and plays the sliding sound. An entity resting on top of the block is
// left alone, so that it is still damaged for the distance it fell.
func (h HoneyBlock) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	fallEntity, ok := e.(fallDistanceEntity)
	if !ok || fallEntity.FallDistance() == 0 {
		return
	}
	sides := h.Model().BBox(pos, tx)[0].Translate(pos.Vec3()).GrowVec3(mgl64.Vec3{0.001, 0, 0.001})
	if !sides.IntersectsWith(e.H().Type().BBox(e).Translate(e.Position())) {
		return
	}
	fallEntity.ResetFallDistance()

	if pos[1] == cube.PosFromVec3(e.Position())[1] && rand.IntN(10) == 0 {
		tx.PlaySound(pos.Vec3Centre(), sound.Custom{Name: "land.honey_block", Volume: 0.14, Pitch: 1})
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

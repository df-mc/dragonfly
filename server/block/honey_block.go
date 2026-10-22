package block

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
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

// EntityInside slides an entity moving down against a side of the block, resetting its fall distance and
// playing the slide sound every 10 ticks.
func (HoneyBlock) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	v, ok := e.(velocityEntity)
	if !ok || v.Velocity()[1] >= 0 {
		return
	}
	ePos, centre := e.Position(), pos.Vec3Middle()
	if ePos[1] < float64(pos[1]) || ePos[1] > float64(pos[1])+0.9375 {
		// Only the block at the entity's feet plays the sound, so a taller wall plays it once.
		return
	}
	side := e.H().Type().BBox(e).Width()/2 + 0.43125
	if math.Abs(centre[0]-ePos[0]) <= side && math.Abs(centre[2]-ePos[2]) <= side {
		return
	}
	if f, ok := e.(fallDistanceEntity); ok {
		f.ResetFallDistance()
	}
	if tx.CurrentTick()%10 == 0 {
		tx.PlaySound(ePos, sound.BlockSlide{Block: HoneyBlock{}, EntityType: e.H().Type().EncodeEntity()})
	}
}

// Friction ...
func (HoneyBlock) Friction() float64 {
	return 0.8
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

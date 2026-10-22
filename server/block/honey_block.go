package block

import (
	"math"
	"time"

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

// EntityInside resets the fall distance of a player sliding down a side of the block and plays the slide sound.
func (h HoneyBlock) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if !h.sliding(pos, e) {
		return
	}
	if f, ok := e.(fallDistanceEntity); ok {
		f.ResetFallDistance()
	}
	h.playSlideSound(tx, e)
}

// SlideVelocity slows down an entity moved by the server that is inside the block, and plays the slide sound if
// it is sliding down a side. Like vanilla, every honey block the entity is inside slows it down.
func (h HoneyBlock) SlideVelocity(pos cube.Pos, tx *world.Tx, e world.Entity, vel mgl64.Vec3) mgl64.Vec3 {
	if h.sliding(pos, e) {
		h.playSlideSound(tx, e)
	}
	return mgl64.Vec3{vel[0] * 0.4, max(vel[1], -0.12), vel[2] * 0.4}
}

// sliding checks if e moved down in its last movement against a side of the block, with its feet no higher than
// 15 pixels above the bottom of the block.
func (HoneyBlock) sliding(pos cube.Pos, e world.Entity) bool {
	if v, ok := e.(velocityEntity); !ok || v.Velocity()[1] >= 0 {
		return false
	}
	ePos, centre := e.Position(), pos.Vec3Middle()
	if ePos[1] > float64(pos[1])+0.9375 {
		return false
	}
	side := e.H().Type().BBox(e).Width()/2 + 0.43125
	return math.Abs(centre[0]-ePos[0]) > side || math.Abs(centre[2]-ePos[2]) > side
}

// playSlideSound plays the slide sound every 10 ticks of the entity's age, or of the world for entities without one.
func (HoneyBlock) playSlideSound(tx *world.Tx, e world.Entity) {
	ticks := tx.CurrentTick()
	if a, ok := e.(interface{ Age() time.Duration }); ok {
		ticks = int64(a.Age() / (time.Second / 20))
	}
	if ticks%10 == 0 {
		tx.PlaySound(e.Position(), sound.BlockSlide{Block: HoneyBlock{}, EntityType: e.H().Type().EncodeEntity()})
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

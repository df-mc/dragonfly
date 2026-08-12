package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// HoneyBlock is a sticky, translucent block crafted from honey bottles. It reduces the fall damage of
// entities that land on it and lets entities pressed against its sides slide down without taking fall
// damage.
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

// EntityInside slows down entities sliding down the side of the block and resets their fall distance, so
// that sliding down it, like sliding down a ladder, does not deal fall damage. While an entity slides,
// the block's landing sound is played on one in every ten ticks.
func (HoneyBlock) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	v, ok := e.(velocityEntity)
	if !ok || v.Velocity()[1] > 0.08 {
		return
	}
	if f, ok := e.(interface{ Flying() bool }); ok && f.Flying() {
		return
	}
	vel := v.Velocity()
	if vel[1] < -0.13 {
		m := -0.05 / vel[1]
		vel[0] *= m
		vel[2] *= m
	}
	vel[1] = -0.05
	v.SetVelocity(vel)

	if fallEntity, ok := e.(fallDistanceEntity); ok {
		fallEntity.ResetFallDistance()
	}
	// An entity is taller than one block and so slides against several honey blocks at once, each of
	// which calls this method every tick. Only the block level the entity's feet are at plays the sound,
	// so that it is played at most once per tick rather than once per block slid against.
	if pos[1] != cube.PosFromVec3(e.Position())[1] {
		return
	}
	if rand.IntN(10) == 0 {
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

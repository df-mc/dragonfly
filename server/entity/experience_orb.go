package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

type ExperienceOrb struct {
	transform
	experienceAmount int16
	age              int

	c *MovementComputer
}

// Name ...
func (e *ExperienceOrb) Name() string {
	return "XP Orb"
}

// EncodeEntity ...
func (e *ExperienceOrb) EncodeEntity() string {
	return "minecraft:xp_orb"
}

// BBox ...
func (e *ExperienceOrb) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}
func NewExperienceOrb(experienceAmount int, pos mgl64.Vec3) *ExperienceOrb {
	if experienceAmount > math.MaxInt16 {
		experienceAmount = math.MaxInt16
	} else if experienceAmount < 1 {
		experienceAmount = 1
	}
	xp := &ExperienceOrb{experienceAmount: int16(experienceAmount), c: &MovementComputer{
		Gravity:           0.03,
		DragBeforeGravity: true,
		Drag:              0.02,
	}}
	xp.transform = newTransform(xp, pos)
	return xp
}
func (e *ExperienceOrb) Amount() int16 {
	return e.experienceAmount
}

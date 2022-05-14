package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

// ExperienceOrb is an entity carrying a varying amount of experience. These can be collected by nearby players, and
// are then added to the player's own experience.
type ExperienceOrb struct {
	transform
	xp int
	c  *MovementComputer
}

// NewExperienceOrb creates a new experience orb and returns it.
func NewExperienceOrb(xp int, pos mgl64.Vec3) *ExperienceOrb {
	o := &ExperienceOrb{
		xp: xp,
		c: &MovementComputer{
			Gravity:           0.04,
			Drag:              0.02,
			DragBeforeGravity: true,
		},
	}
	o.transform = newTransform(o, pos)
	return o
}

// Name ...
func (*ExperienceOrb) Name() string {
	return "Experience Orb"
}

// EncodeEntity ...
func (*ExperienceOrb) EncodeEntity() string {
	return "minecraft:experience_orb"
}

// BBox ...
func (e *ExperienceOrb) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Tick ...
func (e *ExperienceOrb) Tick(int64) {
	//TODO implement me
	panic("implement me")
}

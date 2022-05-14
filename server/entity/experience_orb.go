package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// ExperienceOrb is an entity that carries a varying amount of experience. These can be collected by nearby players, and
// are then added to the player's own experience.
type ExperienceOrb struct {
	transform
	age, xp int
	c       *MovementComputer
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
	return "minecraft:xp_orb"
}

// BBox ...
func (e *ExperienceOrb) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Tick ...
func (e *ExperienceOrb) Tick(w *world.World, current int64) {
	e.mu.Lock()
	m := e.c.TickMovement(e, e.pos, e.vel, 0, 0)
	e.pos, e.vel = m.pos, m.vel
	e.mu.Unlock()

	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		_ = e.Close()
		return
	}
	if e.age++; e.age > 6000 {
		_ = e.Close()
		return
	}

	// TODO: Follow players.
}

// DecodeNBT decodes the properties in a map to an Item and returns a new Item entity.
func (e *ExperienceOrb) DecodeNBT(data map[string]any) any {
	o := NewExperienceOrb(int(nbtconv.Map[int32](data, "Value")), nbtconv.MapVec3(data, "Pos"))
	o.SetVelocity(nbtconv.MapVec3(data, "Motion"))
	o.age = int(nbtconv.Map[int16](data, "Age"))
	return e
}

// EncodeNBT encodes the Item entity's properties as a map and returns it.
func (e *ExperienceOrb) EncodeNBT() map[string]any {
	return map[string]any{
		"Age":    int16(e.age),
		"Value":  int32(e.xp),
		"Pos":    nbtconv.Vec3ToFloat32Slice(e.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(e.Velocity()),
	}
}

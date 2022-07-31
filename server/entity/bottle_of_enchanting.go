package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// BottleOfEnchanting is a bottle that releases experience orbs when thrown.
type BottleOfEnchanting struct {
	transform
	age   int
	close bool

	owner world.Entity

	c *ProjectileComputer
}

// NewBottleOfEnchanting ...
func NewBottleOfEnchanting(pos mgl64.Vec3, owner world.Entity) *BottleOfEnchanting {
	b := &BottleOfEnchanting{
		owner: owner,
		c: &ProjectileComputer{&MovementComputer{
			Gravity:           0.07,
			Drag:              0.01,
			DragBeforeGravity: true,
		}},
	}
	b.transform = newTransform(b, pos)
	return b
}

// Name ...
func (b *BottleOfEnchanting) Name() string {
	return "Bottle o' Enchanting"
}

// EncodeEntity ...
func (b *BottleOfEnchanting) EncodeEntity() string {
	return "minecraft:xp_bottle"
}

// Glint returns true if the bottle should render with glint. It always returns true.
func (b *BottleOfEnchanting) Glint() bool {
	return true
}

// BBox ...
func (b *BottleOfEnchanting) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Tick ...
func (b *BottleOfEnchanting) Tick(w *world.World, current int64) {
	if b.close {
		_ = b.Close()
		return
	}
	b.mu.Lock()
	m, result := b.c.TickMovement(b, b.pos, b.vel, 0, 0, b.ignores)
	b.pos, b.vel = m.pos, m.vel
	b.mu.Unlock()

	b.age++
	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		b.close = true
		return
	}

	if result != nil {
		colour, _ := effect.ResultingColour(nil)
		w.AddParticle(m.pos, particle.Splash{Colour: colour})
		w.PlaySound(m.pos, sound.GlassBreak{})

		for _, orb := range NewExperienceOrbs(m.pos, rand.Intn(9)+3) {
			orb.SetVelocity(mgl64.Vec3{(rand.Float64()*0.2 - 0.1) * 2, rand.Float64() * 0.4, (rand.Float64()*0.2 - 0.1) * 2})
			w.AddEntity(orb)
		}

		b.close = true
	}
}

// ignores returns whether the BottleOfEnchanting should ignore collision with the entity passed.
func (b *BottleOfEnchanting) ignores(entity world.Entity) bool {
	_, ok := entity.(Living)
	return !ok || entity == b || (b.age < 5 && entity == b.owner)
}

// New creates a BottleOfEnchanting with the position, velocity, yaw, and pitch provided. It doesn't spawn the
// BottleOfEnchanting, only returns it.
func (b *BottleOfEnchanting) New(pos, vel mgl64.Vec3, owner world.Entity) world.Entity {
	bottle := NewBottleOfEnchanting(pos, owner)
	bottle.vel = vel
	return bottle
}

// Owner ...
func (b *BottleOfEnchanting) Owner() world.Entity {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.owner
}

// DecodeNBT decodes the properties in a map to a BottleOfEnchanting and returns a new BottleOfEnchanting entity.
func (b *BottleOfEnchanting) DecodeNBT(data map[string]any) any {
	return b.New(
		nbtconv.MapVec3(data, "Pos"),
		nbtconv.MapVec3(data, "Motion"),
		nil,
	)
}

// EncodeNBT encodes the BottleOfEnchanting entity's properties as a map and returns it.
func (b *BottleOfEnchanting) EncodeNBT() map[string]any {
	return map[string]any{
		"Pos":    nbtconv.Vec3ToFloat32Slice(b.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(b.Velocity()),
		"Yaw":    0.0,
		"Pitch":  0.0,
	}
}

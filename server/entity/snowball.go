package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/entity/physics/trace"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// Snowball is a throwable projectile which damages entities on impact.
type Snowball struct {
	transform
	yaw, pitch float64

	ticksLived int

	closeNextTick bool

	owner world.Entity

	c *ProjectileComputer
}

// NewSnowball ...
func NewSnowball(pos mgl64.Vec3, yaw, pitch float64, owner world.Entity) *Snowball {
	s := &Snowball{
		yaw:   yaw,
		pitch: pitch,
		c: &ProjectileComputer{&MovementComputer{
			Gravity:           0.03,
			Drag:              0.01,
			DragBeforeGravity: true,
		}},
		owner: owner,
	}
	s.transform = newTransform(s, pos)

	return s
}

// Name ...
func (s *Snowball) Name() string {
	return "Snowball"
}

// EncodeEntity ...
func (s *Snowball) EncodeEntity() string {
	return "minecraft:snowball"
}

// AABB ...
func (s *Snowball) AABB() physics.AABB {
	return physics.NewAABB(mgl64.Vec3{-0.125, 0, -0.125}, mgl64.Vec3{0.125, 0.25, 0.125})
}

// Rotation ...
func (s *Snowball) Rotation() (float64, float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.yaw, s.pitch
}

// Tick ...
func (s *Snowball) Tick(current int64) {
	if s.closeNextTick {
		_ = s.Close()
		return
	}
	s.mu.Lock()
	m, result := s.c.TickMovement(s, s.pos, s.vel, s.yaw, s.pitch, s.ignores)
	s.pos, s.vel, s.yaw, s.pitch = m.pos, m.vel, m.yaw, m.pitch
	s.mu.Unlock()

	s.ticksLived++
	m.Send()

	if m.pos[1] < cube.MinY && current%10 == 0 {
		s.closeNextTick = true
		return
	}

	if result != nil {
		w := s.World()
		for i := 0; i < 6; i++ {
			w.AddParticle(result.Position(), particle.SnowballPoof{})
		}

		if r, ok := result.(trace.EntityResult); ok {
			if l, ok := r.Entity().(Living); ok {
				if _, vulnerable := l.Hurt(0.0, damage.SourceEntityAttack{Attacker: s}); vulnerable {
					l.KnockBack(m.pos, 0.45, 0.3608)
				}
			}
		}

		s.closeNextTick = true
	}
}

// ignores returns whether the snowball should ignore collision with the entity passed.
func (s *Snowball) ignores(entity world.Entity) bool {
	_, ok := entity.(Living)
	return !ok || entity == s || (s.ticksLived < 5 && entity == s.owner)
}

// New creates a snowball with the position, velocity, yaw, and pitch provided. It doesn't spawn the snowball,
// only returns it.
func (s *Snowball) New(pos, vel mgl64.Vec3, yaw, pitch float64) world.Entity {
	snow := NewSnowball(pos, yaw, pitch, nil)
	snow.vel = vel
	return snow
}

// Owner ...
func (s *Snowball) Owner() world.Entity {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.owner
}

// Own ...
func (s *Snowball) Own(owner world.Entity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.owner = owner
}

// DecodeNBT decodes the properties in a map to a Snowball and returns a new Snowball entity.
func (s *Snowball) DecodeNBT(data map[string]interface{}) interface{} {
	return s.New(
		nbtconv.MapVec3(data, "Pos"),
		nbtconv.MapVec3(data, "Motion"),
		float64(nbtconv.MapFloat32(data, "Pitch")),
		float64(nbtconv.MapFloat32(data, "Yaw")),
	)
}

// EncodeNBT encodes the Snowball entity's properties as a map and returns it.
func (s *Snowball) EncodeNBT() map[string]interface{} {
	yaw, pitch := s.Rotation()
	return map[string]interface{}{
		"Pos":    nbtconv.Vec3ToFloat32Slice(s.Position()),
		"Yaw":    yaw,
		"Pitch":  pitch,
		"Motion": nbtconv.Vec3ToFloat32Slice(s.Velocity()),
		"Damage": 0.0,
	}
}

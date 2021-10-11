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
			DragBeforeGravity: true,
			Drag:              0.01,
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
	return physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{0.25, 0.25, 0.25})
}

// Rotation ...
func (s *Snowball) Rotation() (float64, float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.yaw, s.pitch
}

// Tick ...
func (s *Snowball) Tick(current int64) {
	var result trace.Result
	s.mu.Lock()
	if s.ticksLived < 5 {
		s.pos, s.vel, s.yaw, s.pitch, result = s.c.TickMovement(s, s.pos, s.vel, s.yaw, s.pitch, s.owner)
	} else {
		s.pos, s.vel, s.yaw, s.pitch, result = s.c.TickMovement(s, s.pos, s.vel, s.yaw, s.pitch)
	}
	pos := s.pos
	s.ticksLived++
	s.mu.Unlock()

	if pos[1] < cube.MinY && current%10 == 0 {
		_ = s.Close()
		return
	}

	if result != nil {
		w := s.World()
		for i := 0; i < 6; i++ {
			w.AddParticle(result.Position(), particle.SnowballPoof{})
		}

		if r, ok := result.(trace.EntityResult); ok {
			if l, ok := r.Entity().(Living); ok {
				l.Hurt(0.0, damage.SourceEntityAttack{Attacker: s})
				l.KnockBack(pos, 0.45, 0.3608)
			}
		}

		_ = s.Close()
	}
}

// Launch ...
func (s *Snowball) Launch(pos, vel mgl64.Vec3, yaw, pitch float64) world.Entity {
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
func (it *Snowball) DecodeNBT(data map[string]interface{}) interface{} {
	return it.Launch(
		nbtconv.MapVec3(data, "Pos"),
		nbtconv.MapVec3(data, "Motion"),
		float64(nbtconv.MapFloat32(data, "Pitch")),
		float64(nbtconv.MapFloat32(data, "Yaw")),
	)
}

// EncodeNBT encodes the Snowball entity's properties as a map and returns it.
func (it *Snowball) EncodeNBT() map[string]interface{} {
	yaw, pitch := it.Rotation()
	return map[string]interface{}{
		"Pos":    nbtconv.Vec3ToFloat32Slice(it.Position()),
		"Yaw":    yaw,
		"Pitch":  pitch,
		"Motion": nbtconv.Vec3ToFloat32Slice(it.Velocity()),
		"Damage": 0.0,
	}
}

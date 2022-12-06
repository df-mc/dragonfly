package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// Snowball is a throwable projectile which damages entities on impact.
type Snowball struct {
	transform
	close bool

	owner world.Entity

	c *ProjectileComputer
}

// NewSnowball ...
func NewSnowball(pos mgl64.Vec3, owner world.Entity) *Snowball {
	s := &Snowball{c: newProjectileComputer(0.01, 0.01), owner: owner}
	s.transform = newTransform(s, pos)

	return s
}

// Type returns SnowballType.
func (s *Snowball) Type() world.EntityType {
	return SnowballType{}
}

// Tick ...
func (s *Snowball) Tick(w *world.World, current int64) {
	if s.close {
		_ = s.Close()
		return
	}
	s.mu.Lock()
	m, result := s.c.TickMovement(s, s.pos, s.vel, 0, 0)
	s.pos, s.vel = m.pos, m.vel
	s.mu.Unlock()

	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		s.close = true
		return
	}

	if result != nil {
		for i := 0; i < 6; i++ {
			w.AddParticle(result.Position(), particle.SnowballPoof{})
		}

		if r, ok := result.(trace.EntityResult); ok {
			if l, ok := r.Entity().(Living); ok {
				if _, vulnerable := l.Hurt(0.0, ProjectileDamageSource{Projectile: s, Owner: s.Owner()}); vulnerable {
					l.KnockBack(m.pos, 0.45, 0.3608)
				}
			}
		}

		s.close = true
	}
}

// New creates a snowball with the position, velocity, yaw, and pitch provided. It doesn't spawn the snowball,
// only returns it.
func (s *Snowball) New(pos, vel mgl64.Vec3, owner world.Entity) world.Entity {
	snow := NewSnowball(pos, owner)
	snow.vel = vel
	return snow
}

// Explode ...
func (s *Snowball) Explode(explosionPos mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	s.mu.Lock()
	s.vel = s.vel.Add(s.pos.Sub(explosionPos).Normalize().Mul(impact))
	s.mu.Unlock()
}

// Owner ...
func (s *Snowball) Owner() world.Entity {
	return s.owner
}

// SnowballType is a world.EntityType implementation for Snowball.
type SnowballType struct{}

func (SnowballType) EncodeEntity() string { return "minecraft:snowball" }
func (SnowballType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (SnowballType) DecodeNBT(m map[string]any) world.Entity {
	s := NewSnowball(nbtconv.Vec3(m, "Pos"), nil)
	s.vel = nbtconv.Vec3(m, "Motion")
	return s
}

func (SnowballType) EncodeNBT(e world.Entity) map[string]any {
	s := e.(*Snowball)
	return map[string]any{
		"Pos":    nbtconv.Vec3ToFloat32Slice(s.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(s.Velocity()),
	}
}

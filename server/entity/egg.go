package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// Egg is an item that can be used to craft food items, or as a throwable entity to spawn chicks.
type Egg struct {
	transform
	age   int
	close bool

	owner world.Entity

	c *ProjectileComputer
}

// NewEgg ...
func NewEgg(pos mgl64.Vec3, owner world.Entity) *Egg {
	s := &Egg{
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
func (s *Egg) Name() string {
	return "Egg"
}

// EncodeEntity ...
func (s *Egg) EncodeEntity() string {
	return "minecraft:egg"
}

// BBox ...
func (s *Egg) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Tick ...
func (s *Egg) Tick(w *world.World, current int64) {
	if s.close {
		_ = s.Close()
		return
	}
	s.mu.Lock()
	m, result := s.c.TickMovement(s, s.pos, s.vel, 0, 0, s.ignores)
	s.pos, s.vel = m.pos, m.vel
	s.mu.Unlock()

	s.age++
	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		s.close = true
		return
	}

	if result != nil {
		for i := 0; i < 6; i++ {
			w.AddParticle(result.Position(), particle.EggSmash{})
		}

		if r, ok := result.(trace.EntityResult); ok {
			if l, ok := r.Entity().(Living); ok {
				if _, vulnerable := l.Hurt(0.0, damage.SourceProjectile{Projectile: s, Owner: s.Owner()}); vulnerable {
					l.KnockBack(m.pos, 0.45, 0.3608)
				}
			}
		}

		// TODO: Spawn chicken(s) 12.5% of the time?

		s.close = true
	}
}

// ignores returns whether the egg should ignore collision with the entity passed.
func (s *Egg) ignores(entity world.Entity) bool {
	_, ok := entity.(Living)
	return !ok || entity == s || (s.age < 5 && entity == s.owner)
}

// New creates a egg with the position, velocity, yaw, and pitch provided. It doesn't spawn the egg,
// only returns it.
func (s *Egg) New(pos, vel mgl64.Vec3, owner world.Entity) world.Entity {
	egg := NewEgg(pos, owner)
	egg.vel = vel
	return egg
}

// Explode ...
func (s *Egg) Explode(explosionPos mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	s.mu.Lock()
	s.vel = s.vel.Add(s.pos.Sub(explosionPos).Normalize().Mul(impact))
	s.mu.Unlock()
}

// Owner ...
func (s *Egg) Owner() world.Entity {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.owner
}

// DecodeNBT decodes the properties in a map to a Egg and returns a new Egg entity.
func (s *Egg) DecodeNBT(data map[string]any) any {
	return s.New(
		nbtconv.MapVec3(data, "Pos"),
		nbtconv.MapVec3(data, "Motion"),
		nil,
	)
}

// EncodeNBT encodes the Egg entity's properties as a map and returns it.
func (s *Egg) EncodeNBT() map[string]any {
	return map[string]any{
		"Pos":    nbtconv.Vec3ToFloat32Slice(s.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(s.Velocity()),
	}
}

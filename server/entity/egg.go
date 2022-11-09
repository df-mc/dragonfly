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

// Egg is an item that can be used to craft food items, or as a throwable entity to spawn chicks.
type Egg struct {
	transform
	close bool

	owner world.Entity

	c *ProjectileComputer
}

// NewEgg ...
func NewEgg(pos mgl64.Vec3, owner world.Entity) *Egg {
	s := &Egg{c: newProjectileComputer(0.03, 0.01), owner: owner}
	s.transform = newTransform(s, pos)

	return s
}

// Type returns EggType.
func (egg *Egg) Type() world.EntityType {
	return EggType{}
}

// Tick ...
func (egg *Egg) Tick(w *world.World, current int64) {
	if egg.close {
		_ = egg.Close()
		return
	}
	egg.mu.Lock()
	m, result := egg.c.TickMovement(egg, egg.pos, egg.vel, 0, 0)
	egg.pos, egg.vel = m.pos, m.vel
	egg.mu.Unlock()

	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		egg.close = true
		return
	}

	if result != nil {
		for i := 0; i < 6; i++ {
			w.AddParticle(result.Position(), particle.EggSmash{})
		}

		if r, ok := result.(trace.EntityResult); ok {
			if l, ok := r.Entity().(Living); ok {
				if _, vulnerable := l.Hurt(0.0, ProjectileDamageSource{Projectile: egg, Owner: egg.Owner()}); vulnerable {
					l.KnockBack(m.pos, 0.45, 0.3608)
				}
			}
		}

		// TODO: Spawn chicken(egg) 12.5% of the time?

		egg.close = true
	}
}

// New creates a egg with the position, velocity, yaw, and pitch provided. It doesn't spawn the egg,
// only returns it.
func (*Egg) New(pos, vel mgl64.Vec3, owner world.Entity) world.Entity {
	egg := NewEgg(pos, owner)
	egg.vel = vel
	return egg
}

// Explode ...
func (egg *Egg) Explode(src mgl64.Vec3, force float64, _ block.ExplosionConfig) {
	egg.mu.Lock()
	egg.vel = egg.vel.Add(egg.pos.Sub(src).Normalize().Mul(force))
	egg.mu.Unlock()
}

// Owner ...
func (egg *Egg) Owner() world.Entity {
	egg.mu.Lock()
	defer egg.mu.Unlock()
	return egg.owner
}

// EggType is a world.EntityType implementation for Egg.
type EggType struct{}

func (EggType) EncodeEntity() string { return "minecraft:egg" }
func (EggType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (EggType) DecodeNBT(m map[string]any) world.Entity {
	egg := NewEgg(nbtconv.Vec3(m, "Pos"), nil)
	egg.vel = nbtconv.Vec3(m, "Motion")
	return egg
}

func (EggType) EncodeNBT(e world.Entity) map[string]any {
	egg := e.(*Egg)
	return map[string]any{
		"Pos":    nbtconv.Vec3ToFloat32Slice(egg.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(egg.Velocity()),
	}
}

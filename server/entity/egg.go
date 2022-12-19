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
	e := &Egg{c: newProjectileComputer(0.03, 0.01), owner: owner}
	e.transform = newTransform(e, pos)
	return e
}

// Type returns EggType.
func (e *Egg) Type() world.EntityType {
	return EggType{}
}

// Tick ...
func (e *Egg) Tick(w *world.World, current int64) {
	if e.close {
		_ = e.Close()
		return
	}
	e.mu.Lock()
	pastVel := e.vel
	m, result := e.c.TickMovement(e, e.pos, e.vel, 0, 0)
	e.pos, e.vel = m.pos, m.vel
	e.mu.Unlock()

	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		e.close = true
		return
	}

	if result != nil {
		for i := 0; i < 6; i++ {
			w.AddParticle(result.Position(), particle.EggSmash{})
		}

		if r, ok := result.(trace.EntityResult); ok {
			if l, ok := r.Entity().(Living); ok {
				if _, vulnerable := l.Hurt(0.0, ProjectileDamageSource{Projectile: e, Owner: e.Owner()}); vulnerable {
					horizontalVel := pastVel
					horizontalVel[1] = 0
					l.KnockBack(l.Position().Sub(horizontalVel), 0.4, 0.4)
				}
			}
		}

		// TODO: Spawn chicken(e) 12.5% of the time?

		e.close = true
	}
}

// Explode ...
func (e *Egg) Explode(src mgl64.Vec3, force float64, _ block.ExplosionConfig) {
	e.mu.Lock()
	e.vel = e.vel.Add(e.pos.Sub(src).Normalize().Mul(force))
	e.mu.Unlock()
}

// Owner ...
func (e *Egg) Owner() world.Entity {
	return e.owner
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

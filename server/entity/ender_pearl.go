package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

type EnderPearlType struct{}

func (EnderPearlType) String() string {
	return "Ender Pearl"
}

func (EnderPearlType) EncodeEntity() string {
	return "minecraft:ender_pearl"
}

func (EnderPearlType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (EnderPearlType) DecodeNBT(data map[string]any) world.Entity {
	ep := NewEnderPearl(nbtconv.MapVec3(data, "Pos"), nil)
	ep.vel = nbtconv.MapVec3(data, "Motion")
	return ep
}

func (EnderPearlType) EncodeNBT(e world.Entity) map[string]any {
	ep := e.(*EnderPearl)
	return map[string]any{
		"Pos":    nbtconv.Vec3ToFloat32Slice(ep.Position()),
		"Motion": nbtconv.Vec3ToFloat32Slice(ep.Velocity()),
	}
}

// EnderPearl is a smooth, greenish-blue item used to teleport and to make an eye of ender.
type EnderPearl struct {
	transform
	age   int
	close bool

	owner world.Entity

	c *ProjectileComputer
}

// NewEnderPearl ...
func NewEnderPearl(pos mgl64.Vec3, owner world.Entity) *EnderPearl {
	e := &EnderPearl{
		c: &ProjectileComputer{&MovementComputer{
			Gravity:           0.03,
			Drag:              0.01,
			DragBeforeGravity: true,
		}},
		owner: owner,
	}
	e.transform = newTransform(e, pos)

	return e
}

func (e *EnderPearl) Type() world.EntityType {
	return EnderPearlType{}
}

// teleporter represents a living entity that can teleport.
type teleporter interface {
	// Teleport teleports the entity to the position given.
	Teleport(pos mgl64.Vec3)
	Living
}

// Tick ...
func (e *EnderPearl) Tick(w *world.World, current int64) {
	if e.close {
		_ = e.Close()
		return
	}
	e.mu.Lock()
	m, result := e.c.TickMovement(e, e.pos, e.vel, 0, 0, e.ignores)
	e.pos, e.vel = m.pos, m.vel

	owner := e.owner
	e.mu.Unlock()

	e.age++
	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		e.close = true
		return
	}

	if result != nil {
		if r, ok := result.(trace.EntityResult); ok {
			if l, ok := r.Entity().(Living); ok {
				if _, vulnerable := l.Hurt(0.0, ProjectileDamageSource{Projectile: e, Owner: owner}); vulnerable {
					l.KnockBack(m.pos, 0.45, 0.3608)
				}
			}
		}

		if owner != nil {
			if user, ok := owner.(teleporter); ok {
				w.PlaySound(user.Position(), sound.Teleport{})

				user.Teleport(m.pos)

				w.AddParticle(m.pos, particle.EndermanTeleportParticle{})
				w.PlaySound(m.pos, sound.Teleport{})

				user.Hurt(5, FallDamageSource{})
			}
		}

		e.close = true
	}
}

// ignores returns whether the ender pearl should ignore collision with the entity passed.
func (e *EnderPearl) ignores(entity world.Entity) bool {
	_, ok := entity.(Living)
	return !ok || entity == e || (e.age < 5 && entity == e.owner)
}

// New creates an ender pearl with the position, velocity, yaw, and pitch provided. It doesn't spawn the ender pearl,
// only returns it.
func (e *EnderPearl) New(pos, vel mgl64.Vec3, owner world.Entity) world.Entity {
	pearl := NewEnderPearl(pos, owner)
	pearl.vel = vel
	return pearl
}

// Explode ...
func (e *EnderPearl) Explode(explosionPos mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	e.mu.Lock()
	e.vel = e.vel.Add(e.pos.Sub(explosionPos).Normalize().Mul(impact))
	e.mu.Unlock()
}

// Owner ...
func (e *EnderPearl) Owner() world.Entity {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.owner
}

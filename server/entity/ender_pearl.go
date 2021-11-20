package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/entity/physics/trace"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// EnderPearl is a smooth, greenish-blue item used to teleport and to make an eye of ender.
type EnderPearl struct {
	transform
	yaw, pitch float64

	ticksLived int

	closeNextTick bool

	owner world.Entity

	c *ProjectileComputer
}

// NewEnderPearl ...
func NewEnderPearl(pos mgl64.Vec3, yaw, pitch float64, owner world.Entity) *EnderPearl {
	e := &EnderPearl{
		yaw:   yaw,
		pitch: pitch,
		c: &ProjectileComputer{&MovementComputer{
			Gravity:           0.03,
			DragBeforeGravity: true,
			Drag:              0.01,
		}},
		owner: owner,
	}
	e.transform = newTransform(e, pos)

	return e
}

// Name ...
func (e *EnderPearl) Name() string {
	return "Ender Pearl"
}

// EncodeEntity ...
func (e *EnderPearl) EncodeEntity() string {
	return "minecraft:ender_pearl"
}

// AABB ...
func (e *EnderPearl) AABB() physics.AABB {
	return physics.NewAABB(mgl64.Vec3{-0.125, 0, -0.125}, mgl64.Vec3{0.125, 0.25, 0.125})
}

// Rotation ...
func (e *EnderPearl) Rotation() (float64, float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.yaw, e.pitch
}

// teleporter represents a living entity that can teleport.
type teleporter interface {
	// Teleport teleports the entity to the position given.
	Teleport(pos mgl64.Vec3)
	Living
}

// Tick ...
func (e *EnderPearl) Tick(current int64) {
	if e.closeNextTick {
		_ = e.Close()
		return
	}
	var result trace.Result
	e.mu.Lock()
	if e.ticksLived < 5 {
		e.pos, e.vel, e.yaw, e.pitch, result = e.c.TickMovement(e, e.pos, e.vel, e.yaw, e.pitch, e.owner)
	} else {
		e.pos, e.vel, e.yaw, e.pitch, result = e.c.TickMovement(e, e.pos, e.vel, e.yaw, e.pitch)
	}
	pos := e.pos
	e.ticksLived++
	e.mu.Unlock()

	if pos[1] < cube.MinY && current%10 == 0 {
		e.closeNextTick = true
		return
	}

	if result != nil {
		w := e.World()
		if r, ok := result.(trace.EntityResult); ok {
			if l, ok := r.Entity().(Living); ok {
				l.Hurt(0.0, damage.SourceEntityAttack{Attacker: e})
				l.KnockBack(pos, 0.45, 0.3608)
			}
		}

		owner := e.Owner()
		if owner != nil {
			if user, ok := owner.(teleporter); ok {
				shooterPos := user.Position()
				w.PlaySound(shooterPos, sound.EndermanTeleport{})

				user.Teleport(pos)

				w.AddParticle(pos, particle.EndermanTeleportParticle{})
				w.PlaySound(pos, sound.EndermanTeleport{})

				user.Hurt(5, damage.SourceFall{})
			}
		}

		e.closeNextTick = true
	}
}

// Launch creates a EnderPearl with the position, velocity, yaw, and pitch provided. It doesn't spawn the EnderPearl,
// only returns it.
func (e *EnderPearl) Launch(pos, vel mgl64.Vec3, yaw, pitch float64) world.Entity {
	snow := NewEnderPearl(pos, yaw, pitch, nil)
	snow.vel = vel
	return snow
}

// Owner ...
func (e *EnderPearl) Owner() world.Entity {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.owner
}

// Own ...
func (e *EnderPearl) Own(owner world.Entity) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.owner = owner
}

// DecodeNBT decodes the properties in a map to a EnderPearl and returns a new EnderPearl entity.
func (e *EnderPearl) DecodeNBT(data map[string]interface{}) interface{} {
	return e.Launch(
		nbtconv.MapVec3(data, "Pos"),
		nbtconv.MapVec3(data, "Motion"),
		float64(nbtconv.MapFloat32(data, "Pitch")),
		float64(nbtconv.MapFloat32(data, "Yaw")),
	)
}

// EncodeNBT encodes the EnderPearl entity's properties as a map and returns it.
func (e *EnderPearl) EncodeNBT() map[string]interface{} {
	yaw, pitch := e.Rotation()
	return map[string]interface{}{
		"Pos":    nbtconv.Vec3ToFloat32Slice(e.Position()),
		"Yaw":    yaw,
		"Pitch":  pitch,
		"Motion": nbtconv.Vec3ToFloat32Slice(e.Velocity()),
		"Damage": 0.0,
	}
}

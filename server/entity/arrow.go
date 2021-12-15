package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/action"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/entity/physics/trace"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand"
	"time"
)

// Arrow is used as ammunition for bows, crossbows, and dispensers. Arrows can be modified to
// imbue status effects on players and mobs.
type Arrow struct {
	transform
	yaw, pitch float64
	baseDamage float64

	ticksLived, collisionTicks int

	collidedBlockPos cube.Pos
	collidedBlock    world.Block

	shakeNextTick, closeNextTick, critical bool

	owner                     world.Entity
	canPickup, creativePickup bool

	c *ProjectileComputer
}

// NewArrow ...
func NewArrow(pos mgl64.Vec3, yaw, pitch float64, owner world.Entity, critical, canPickup, creativePickup bool, baseDamage float64) *Arrow {
	s := &Arrow{
		yaw:   yaw,
		pitch: pitch,
		c: &ProjectileComputer{&MovementComputer{
			Gravity:           0.05,
			Drag:              0.01,
			DragBeforeGravity: true,
		}},
		baseDamage:     baseDamage,
		canPickup:      canPickup,
		creativePickup: creativePickup,
		critical:       critical,
		owner:          owner,
	}
	s.transform = newTransform(s, pos)

	return s
}

// Name ...
func (a *Arrow) Name() string {
	return "Arrow"
}

// EncodeEntity ...
func (a *Arrow) EncodeEntity() string {
	return "minecraft:arrow"
}

// Critical ...
func (a *Arrow) Critical() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.critical
}

// SetCritical ...
func (a *Arrow) SetCritical(critical bool) {
	a.mu.Lock()
	a.critical = critical
	a.mu.Unlock()

	for _, v := range a.World().Viewers(a.Position()) {
		v.ViewEntityState(a)
	}
}

// AABB ...
func (a *Arrow) AABB() physics.AABB {
	return physics.NewAABB(mgl64.Vec3{-0.125, 0, -0.125}, mgl64.Vec3{0.125, 0.25, 0.125})
}

// Rotation ...
func (a *Arrow) Rotation() (float64, float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.yaw, a.pitch
}

// Tick ...
func (a *Arrow) Tick(current int64) {
	if a.closeNextTick {
		_ = a.Close()
		return
	}
	if a.shakeNextTick {
		for _, v := range a.World().Viewers(a.Position()) {
			v.ViewEntityAction(a, action.ArrowShake{Duration: time.Millisecond * 350})
		}
		a.shakeNextTick = false
		return
	}

	w := a.World()
	if w.Block(a.collidedBlockPos) == a.collidedBlock {
		if a.collisionTicks > 5 {
			a.checkNearby()
		}
		a.collisionTicks++
		return
	}

	a.mu.Lock()
	m, result := a.c.TickMovement(a, a.pos, a.vel, a.yaw, a.pitch, a.ignores)
	a.pos, a.vel, a.yaw, a.pitch = m.pos, m.vel, m.yaw, m.pitch
	a.mu.Unlock()

	a.ticksLived++
	a.collisionTicks = 0
	a.collidedBlockPos, a.collidedBlock = cube.Pos{}, nil
	m.Send()

	if m.pos[1] < cube.MinY && current%10 == 0 {
		a.closeNextTick = true
		return
	}

	if result != nil {
		a.SetCritical(false)
		w.PlaySound(a.Position(), sound.ArrowHit{})

		if blockResult, ok := result.(trace.BlockResult); ok {
			a.collidedBlockPos = blockResult.BlockPosition()
			a.collidedBlock = w.Block(a.collidedBlockPos)
			a.shakeNextTick = true
		} else if entityResult, ok := result.(trace.EntityResult); ok {
			if living, ok := entityResult.Entity().(Living); ok {
				if !living.AttackImmune() {
					living.Hurt(a.damage(), damage.SourceProjectile{Owner: a.owner})
					living.KnockBack(a.Position(), 0.45, 0.3608)
				}
			}
			a.closeNextTick = true
		}
	}
}

// ignores returns whether the arrow should ignore collision with the entity passed.
func (a *Arrow) ignores(entity world.Entity) bool {
	_, ok := entity.(Living)
	return !ok || entity == a || (a.ticksLived < 5 && entity == a.owner)
}

// New creates an arrow with the position, velocity, yaw, and pitch provided. It doesn't spawn the arrow,
// only returns it.
func (a *Arrow) New(pos, vel mgl64.Vec3, yaw, pitch float64, critical, canPickup, creativePickup bool, baseDamage float64) world.Entity {
	arrow := NewArrow(pos, yaw, pitch, nil, critical, canPickup, creativePickup, baseDamage)
	arrow.vel = vel
	return arrow
}

// Owner ...
func (a *Arrow) Owner() world.Entity {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.owner
}

// Own ...
func (a *Arrow) Own(owner world.Entity) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.owner = owner
}

// DecodeNBT decodes the properties in a map to an Arrow and returns a new Arrow entity.
func (a *Arrow) DecodeNBT(data map[string]interface{}) interface{} {
	pickupValue := nbtconv.MapByte(data, "Pickup")
	return a.New(
		nbtconv.MapVec3(data, "Pos"),
		nbtconv.MapVec3(data, "Motion"),
		float64(nbtconv.MapFloat32(data, "Pitch")),
		float64(nbtconv.MapFloat32(data, "Yaw")),
		false, // Vanilla doesn't save this value, so we don't either.
		pickupValue > 0,
		pickupValue == 2,
		float64(nbtconv.MapFloat32(data, "Damage")),
	).(*Arrow)
}

// EncodeNBT encodes the Arrow entity's properties as a map and returns it.
func (a *Arrow) EncodeNBT() map[string]interface{} {
	var pickupValue byte
	if a.creativePickup {
		pickupValue = 2
	} else if a.canPickup {
		pickupValue = 1
	}

	yaw, pitch := a.Rotation()
	return map[string]interface{}{
		"Pos":    nbtconv.Vec3ToFloat32Slice(a.Position()),
		"Yaw":    yaw,
		"Pitch":  pitch,
		"Motion": nbtconv.Vec3ToFloat32Slice(a.Velocity()),
		"Damage": a.baseDamage,
		"Pickup": pickupValue,
	}
}

// checkNearby checks for nearby arrow collectors.
func (a *Arrow) checkNearby() {
	w := a.World()
	grown := a.AABB().GrowVec3(mgl64.Vec3{1, 0.5, 1}).Translate(a.Position())
	ignore := func(e world.Entity) bool {
		return e == a
	}
	for _, e := range a.World().EntitiesWithin(a.AABB().Translate(a.Position()).Grow(2), ignore) {
		if e.AABB().Translate(e.Position()).IntersectsWith(grown) {
			if collector, ok := e.(Collector); ok {
				isCreative := collector.GameMode() == world.GameModeCreative
				if !a.canPickup {
					return
				}

				for _, viewer := range w.Viewers(a.Position()) {
					viewer.ViewEntityAction(a, action.PickedUp{Collector: collector})
				}
				if !isCreative && !a.creativePickup {
					// A collector was within range to pick up the entity.
					_ = collector.Collect(item.NewStack(item.Arrow{}, 1))
				}
				a.closeNextTick = true
				return
			}
		}
	}
}

// damage returns the full damage the arrow should deal, accounting for the velocity.
func (a *Arrow) damage() float64 {
	base := math.Ceil(a.Velocity().Len() * a.baseDamage)
	if a.critical {
		return base + float64(rand.Intn(int(base/2+1)))
	}
	return base
}

package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/action"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/entity/physics/trace"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/potion"
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

	tip potion.Potion

	closeNextTick, critical bool

	owner                        world.Entity
	shotByPlayer, shotInCreative bool

	c *ProjectileComputer
}

// NewArrow ...
func NewArrow(pos mgl64.Vec3, yaw, pitch float64, owner world.Entity, critical, shotByPlayer, shotInCreative bool, baseDamage float64, tip potion.Potion) *Arrow {
	a := &Arrow{
		yaw:   yaw,
		pitch: pitch,
		c: &ProjectileComputer{&MovementComputer{
			Gravity:           0.05,
			Drag:              0.01,
			DragBeforeGravity: true,
		}},
		baseDamage:     baseDamage,
		shotByPlayer:   shotByPlayer,
		shotInCreative: shotInCreative,
		critical:       critical,
		owner:          owner,
		tip:            tip,
	}
	a.transform = newTransform(a, pos)
	return a
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

// Effects ...
func (a *Arrow) Effects() []effect.Effect {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.tip.Effects()
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

	if m.pos[1] < float64(a.World().Range()[0]) && current%10 == 0 {
		a.closeNextTick = true
		return
	}

	if result != nil {
		a.SetCritical(false)
		w.PlaySound(a.Position(), sound.ArrowHit{})

		if blockResult, ok := result.(trace.BlockResult); ok {
			a.collidedBlockPos = blockResult.BlockPosition()
			a.collidedBlock = w.Block(a.collidedBlockPos)

			for _, v := range a.World().Viewers(a.Position()) {
				v.ViewEntityAction(a, action.ArrowShake{Duration: time.Millisecond * 350})
			}
		} else if entityResult, ok := result.(trace.EntityResult); ok {
			if living, ok := entityResult.Entity().(Living); ok {
				if !living.AttackImmune() {
					living.Hurt(a.damage(), damage.SourceProjectile{Owner: a.owner})
					living.KnockBack(a.Position(), 0.45, 0.3608)
					for _, eff := range a.tip.Effects() {
						living.AddEffect(eff)
					}
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
func (a *Arrow) New(pos, vel mgl64.Vec3, yaw, pitch float64, critical, shotByPlayer, shotInCreative bool, baseDamage float64, tip potion.Potion) world.Entity {
	arrow := NewArrow(pos, yaw, pitch, nil, critical, shotByPlayer, shotInCreative, baseDamage, tip)
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
	return a.New(
		nbtconv.MapVec3(data, "Pos"),
		nbtconv.MapVec3(data, "Motion"),
		float64(nbtconv.MapFloat32(data, "Pitch")),
		float64(nbtconv.MapFloat32(data, "Yaw")),
		false, // Vanilla doesn't save this value, so we don't either.
		nbtconv.MapByte(data, "player") == 1,
		nbtconv.MapByte(data, "isCreative") == 1,
		float64(nbtconv.MapFloat32(data, "Damage")),
		potion.From(nbtconv.MapInt32(data, "Tip")),
	).(*Arrow)
}

// EncodeNBT encodes the Arrow entity's properties as a map and returns it.
func (a *Arrow) EncodeNBT() map[string]interface{} {
	yaw, pitch := a.Rotation()
	return map[string]interface{}{
		"Pos":        nbtconv.Vec3ToFloat32Slice(a.Position()),
		"Yaw":        yaw,
		"Pitch":      pitch,
		"Motion":     nbtconv.Vec3ToFloat32Slice(a.Velocity()),
		"Damage":     a.baseDamage,
		"Tip":        int32(a.tip.Uint8()),
		"player":     boolByte(a.shotByPlayer),
		"isCreative": boolByte(a.shotInCreative),
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
				if !a.shotByPlayer {
					return
				}

				for _, viewer := range w.Viewers(a.Position()) {
					viewer.ViewEntityAction(a, action.PickedUp{Collector: collector})
				}
				if !isCreative && !a.shotInCreative {
					// A collector was within range to pick up the entity.
					_ = collector.Collect(item.NewStack(item.Arrow{Tip: a.tip}, 1))
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

// boolByte returns 1 if the bool passed is true, or 0 if it is false.
func boolByte(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

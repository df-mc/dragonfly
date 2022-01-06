package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/action"
	"github.com/df-mc/dragonfly/server/entity/damage"
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

// Arrow is used as ammunition for bows, crossbows, and dispensers. Arrows can be modified to imbue status effects
// on players and mobs.
type Arrow struct {
	transform
	yaw, pitch float64
	baseDamage float64

	ticksLived, collisionTicks int

	collidedBlockPos cube.Pos
	collidedBlock    world.Block

	tip potion.Potion

	closeNextTick, critical bool

	owner                               world.Entity
	disallowPickup, obtainArrowOnPickup bool

	c *ProjectileComputer
}

// NewArrow ...
func NewArrow(pos mgl64.Vec3, yaw, pitch float64, owner world.Entity, critical, disallowPickup, obtainArrowOnPickup bool) *Arrow {
	return NewTippedArrow(pos, yaw, pitch, owner, critical, disallowPickup, obtainArrowOnPickup, potion.Potion{})
}

// NewTippedArrow ...
func NewTippedArrow(pos mgl64.Vec3, yaw, pitch float64, owner world.Entity, critical, disallowPickup, obtainArrowOnPickup bool, tip potion.Potion) *Arrow {
	a := &Arrow{
		yaw:   yaw,
		pitch: pitch,
		c: &ProjectileComputer{&MovementComputer{
			Gravity:           0.05,
			Drag:              0.01,
			DragBeforeGravity: true,
		}},
		baseDamage:          2.0,
		disallowPickup:      disallowPickup,
		obtainArrowOnPickup: obtainArrowOnPickup,
		critical:            critical,
		owner:               owner,
		tip:                 tip,
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

// CollisionPos returns the position of the block the arrow collided with. If the arrow has not collided with any
// blocks, it returns false for it's second parameter.
func (a *Arrow) CollisionPos() (cube.Pos, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.collidedBlockPos, a.collidedBlock != nil
}

// Critical returns the critical state of the arrow, which can result in more damage and extra particles while in air.
func (a *Arrow) Critical() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.critical
}

// SetCritical sets the critical state of the arrow to true, which can result in more damage and extra particles while
// in air.
func (a *Arrow) SetCritical(critical bool) {
	a.mu.Lock()
	a.critical = critical
	pos := a.pos
	a.mu.Unlock()

	for _, v := range a.World().Viewers(pos) {
		v.ViewEntityState(a)
	}
}

// BaseDamage returns the base damage the arrow will deal, before accounting for velocity.
func (a *Arrow) BaseDamage() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.baseDamage
}

// SetBaseDamage sets the base damage the arrow will deal, before accounting for velocity.
func (a *Arrow) SetBaseDamage(baseDamage float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.baseDamage = baseDamage
}

// Tip returns the potion effect at the tip of the arrow, applied on impact to an entity. This also causes the arrow
// to show effect particles until it is removed.
func (a *Arrow) Tip() potion.Potion {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.tip
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

	if a.collidedBlock != nil {
		now, _ := world.BlockRuntimeID(w.Block(a.collidedBlockPos))
		last, _ := world.BlockRuntimeID(a.collidedBlock)
		if now == last {
			if a.collisionTicks > 5 {
				a.checkNearby()
			}
			a.collisionTicks++
			return
		}
	}

	a.mu.Lock()
	m, result := a.c.TickMovement(a, a.pos, a.vel, a.yaw, a.pitch, a.ignores)
	a.pos, a.vel, a.yaw, a.pitch = m.pos, m.vel, m.yaw, m.pitch
	a.mu.Unlock()

	a.ticksLived++
	a.collisionTicks = 0
	a.collidedBlockPos, a.collidedBlock = cube.Pos{}, nil
	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		a.closeNextTick = true
		return
	}

	if result != nil {
		a.SetCritical(false)
		w.PlaySound(m.pos, sound.ArrowHit{})

		if blockResult, ok := result.(trace.BlockResult); ok {
			a.collidedBlockPos = blockResult.BlockPosition()
			a.collidedBlock = w.Block(a.collidedBlockPos)

			for _, v := range w.Viewers(m.pos) {
				v.ViewEntityAction(a, action.ArrowShake{Duration: time.Millisecond * 350})
			}
		} else if entityResult, ok := result.(trace.EntityResult); ok {
			if living, ok := entityResult.Entity().(Living); ok {
				if !living.AttackImmune() {
					living.Hurt(a.damage(), damage.SourceProjectile{Owner: a.owner})
					living.KnockBack(m.pos, 0.45, 0.3608)
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
func (a *Arrow) New(pos, vel mgl64.Vec3, yaw, pitch float64, critical, disallowPickup, obtainArrowOnPickup bool, tip potion.Potion) world.Entity {
	arrow := NewTippedArrow(pos, yaw, pitch, nil, critical, disallowPickup, obtainArrowOnPickup, tip)
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
	arr := a.New(
		nbtconv.MapVec3(data, "Pos"),
		nbtconv.MapVec3(data, "Motion"),
		float64(nbtconv.MapFloat32(data, "Pitch")),
		float64(nbtconv.MapFloat32(data, "Yaw")),
		false, // Vanilla doesn't save this value, so we don't either.
		nbtconv.MapByte(data, "player") == 1,
		nbtconv.MapByte(data, "isCreative") == 1,
		potion.From(nbtconv.MapInt32(data, "auxValue")-1),
	).(*Arrow)
	arr.baseDamage = float64(nbtconv.MapFloat32(data, "Damage"))
	arr.collidedBlockPos = nbtconv.MapPos(data, "StuckToBlockPos")
	arr.collidedBlock = a.World().Block(arr.collidedBlockPos)
	return arr
}

// EncodeNBT encodes the Arrow entity's properties as a map and returns it.
func (a *Arrow) EncodeNBT() map[string]interface{} {
	yaw, pitch := a.Rotation()
	nbt := map[string]interface{}{
		"Pos":        nbtconv.Vec3ToFloat32Slice(a.Position()),
		"Yaw":        yaw,
		"Pitch":      pitch,
		"Motion":     nbtconv.Vec3ToFloat32Slice(a.Velocity()),
		"Damage":     a.BaseDamage(),
		"auxValue":   int32(a.tip.Uint8() + 1),
		"player":     boolByte(!a.disallowPickup),
		"isCreative": boolByte(!a.obtainArrowOnPickup),
	}
	if collisionPos, ok := a.CollisionPos(); ok {
		nbt["StuckToBlockPos"] = nbtconv.PosToInt32Slice(collisionPos)
	}
	return nbt
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
				if a.disallowPickup {
					return
				}

				if a.obtainArrowOnPickup {
					// A collector was within range to pick up the entity.
					for _, viewer := range w.Viewers(a.Position()) {
						viewer.ViewEntityAction(a, action.PickedUp{Collector: collector})
					}
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
	base := math.Ceil(a.Velocity().Len() * a.BaseDamage())
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

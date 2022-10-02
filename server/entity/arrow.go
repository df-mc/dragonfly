package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
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

	age, ageCollided int
	close, critical  bool

	collisionPos cube.Pos
	collided     bool

	owner world.Entity
	tip   potion.Potion

	fireTicks  int64
	punchLevel int

	disallowPickup, obtainArrowOnPickup bool

	c *ProjectileComputer
}

// NewArrow creates a new Arrow and returns it. It is equivalent to calling NewTippedArrow with `potion.Potion{}` as
// tip.
func NewArrow(pos mgl64.Vec3, yaw, pitch float64, owner world.Entity) *Arrow {
	return NewTippedArrowWithDamage(pos, yaw, pitch, 2.0, owner, potion.Potion{})
}

// NewArrowWithDamage creates a new Arrow with the given base damage, and returns it. It is equivalent to calling
// NewTippedArrowWithDamage with `potion.Potion{}` as tip.
func NewArrowWithDamage(pos mgl64.Vec3, yaw, pitch, damage float64, owner world.Entity) *Arrow {
	return NewTippedArrowWithDamage(pos, yaw, pitch, damage, owner, potion.Potion{})
}

// NewTippedArrow creates a new Arrow with a potion effect added to an entity when hit.
func NewTippedArrow(pos mgl64.Vec3, yaw, pitch float64, owner world.Entity, tip potion.Potion) *Arrow {
	return NewTippedArrowWithDamage(pos, yaw, pitch, 2.0, owner, tip)
}

// NewTippedArrowWithDamage creates a new Arrow with a potion effect added to an entity when hit and, and returns it.
// It uses the given damage as the base damage.
func NewTippedArrowWithDamage(pos mgl64.Vec3, yaw, pitch, damage float64, owner world.Entity, tip potion.Potion) *Arrow {
	a := &Arrow{
		yaw:                 yaw,
		pitch:               pitch,
		baseDamage:          damage,
		owner:               owner,
		tip:                 tip,
		obtainArrowOnPickup: true,
		c: &ProjectileComputer{&MovementComputer{
			Gravity:           0.05,
			Drag:              0.01,
			DragBeforeGravity: true,
		}},
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
	return a.collisionPos, a.collided
}

// Critical returns the critical state of the arrow, which can result in more damage and extra particles while in air.
func (a *Arrow) Critical() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.critical
}

// SetCritical sets the critical state of the arrow to true, which can result in more damage and extra particles while
// in air.
func (a *Arrow) SetCritical() {
	a.setCritical(true)
}

// setCritical changes the critical state of the Arrow and sends the update to any viewers.
func (a *Arrow) setCritical(critical bool) {
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

// Tip returns the potion effect at the tip of the arrow, applied on impact to an entity. This also causes the arrow
// to show effect particles until it is removed.
func (a *Arrow) Tip() potion.Potion {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.tip
}

// DisallowPickup prevents the Arrow from being picked up by any entity.
func (a *Arrow) DisallowPickup() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.disallowPickup = true
}

// VanishOnPickup makes the Arrow vanish on pickup, giving the entity that picked it up no arrow item.
func (a *Arrow) VanishOnPickup() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.obtainArrowOnPickup = false
}

// FireProof ...
func (a *Arrow) FireProof() bool {
	return false
}

// OnFireDuration ...
func (a *Arrow) OnFireDuration() time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()
	return time.Duration(a.fireTicks) * time.Second / 20
}

// SetOnFire ...
func (a *Arrow) SetOnFire(duration time.Duration) {
	a.mu.Lock()
	a.fireTicks = int64(duration.Seconds() * 20)
	pos := a.pos
	a.mu.Unlock()

	for _, v := range a.World().Viewers(pos) {
		v.ViewEntityState(a)
	}
}

// Extinguish ...
func (a *Arrow) Extinguish() {
	a.SetOnFire(0)
}

// BBox ...
func (a *Arrow) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Rotation ...
func (a *Arrow) Rotation() (float64, float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.yaw, a.pitch
}

// Tick ...
func (a *Arrow) Tick(w *world.World, current int64) {
	if a.close {
		_ = a.Close()
		return
	}

	a.mu.Lock()
	if a.collided {
		boxes := w.Block(a.collisionPos).Model().BBox(a.collisionPos, w)
		box := a.BBox().Translate(a.pos)
		for _, bb := range boxes {
			if box.IntersectsWith(bb.Translate(a.collisionPos.Vec3()).Grow(0.05)) {
				if a.ageCollided > 5 && !a.disallowPickup {
					a.checkNearby(w)
				}
				a.ageCollided++
				a.mu.Unlock()
				return
			}
		}
	}

	pastVel := a.vel
	m, result := a.c.TickMovement(a, a.pos, a.vel, a.yaw, a.pitch, a.ignores)
	a.pos, a.vel, a.yaw, a.pitch = m.pos, m.vel, m.yaw, m.pitch
	a.collisionPos, a.collided = cube.Pos{}, false
	a.mu.Unlock()

	a.age++
	a.ageCollided = 0
	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 || a.age > 1200 {
		a.close = true
		return
	}

	if result != nil {
		if res, ok := result.(trace.BlockResult); ok {
			a.mu.Lock()
			a.collisionPos, a.collided = res.BlockPosition(), true
			if t, ok := w.Block(a.collisionPos).(block.TNT); ok && a.fireTicks > 0 {
				t.Ignite(a.collisionPos, w)
			}
			a.mu.Unlock()

			for _, v := range w.Viewers(m.pos) {
				v.ViewEntityAction(a, ArrowShakeAction{Duration: time.Millisecond * 350})
			}
		} else if res, ok := result.(trace.EntityResult); ok {
			if living, ok := res.Entity().(Living); ok {
				living.Hurt(a.damage(pastVel), damage.SourceProjectile{Projectile: a, Owner: a.owner})
				living.KnockBack(m.pos, 0.45, 0.3608)
				for _, eff := range a.tip.Effects() {
					living.AddEffect(eff)
				}
				if flammable, ok := living.(Flammable); ok && a.OnFireDuration() > 0 {
					flammable.SetOnFire(time.Second * 5)
				}
				if a.punchLevel > 0 {
					horizontalVel := pastVel
					horizontalVel[1] = 0
					if speed := horizontalVel.Len(); speed > 0 {
						multiplier := (enchantment.Punch{}).PunchMultiplier(a.punchLevel, speed)
						living.SetVelocity(living.Velocity().Add(mgl64.Vec3{pastVel[0] * multiplier, 0.1, pastVel[2] * multiplier}))
					}
				}
				a.close = true
			}
		}

		a.setCritical(false)
		w.PlaySound(m.pos, sound.ArrowHit{})
	}
}

// ignores returns whether the arrow should ignore collision with the entity passed.
func (a *Arrow) ignores(entity world.Entity) bool {
	_, ok := entity.(Living)
	return !ok || entity == a || (a.age < 5 && entity == a.owner)
}

// New creates and returns an Arrow with the position, velocity, yaw, and pitch provided. It doesn't spawn the Arrow
// by itself.
func (a *Arrow) New(pos, vel mgl64.Vec3, yaw, pitch, damage float64, owner world.Entity, critical, disallowPickup, obtainArrowOnPickup bool, punchLevel int, tip potion.Potion) world.Entity {
	arrow := NewTippedArrowWithDamage(pos, yaw, pitch, damage, owner, tip)
	arrow.vel = vel
	arrow.punchLevel = punchLevel
	arrow.disallowPickup = disallowPickup
	arrow.obtainArrowOnPickup = obtainArrowOnPickup
	arrow.setCritical(critical)
	return arrow
}

// Explode ...
func (a *Arrow) Explode(explosionPos mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	a.mu.Lock()
	a.vel = a.vel.Add(a.pos.Sub(explosionPos).Normalize().Mul(impact))
	a.mu.Unlock()
}

// Owner returns the world.Entity that fired the Arrow, or nil if it did not have any.
func (a *Arrow) Owner() world.Entity {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.owner
}

// DecodeNBT decodes the properties in a map to an Arrow and returns a new Arrow entity.
func (a *Arrow) DecodeNBT(data map[string]any) any {
	arr := NewTippedArrowWithDamage(nbtconv.MapVec3(data, "Pos"), float64(nbtconv.Map[float32](data, "Yaw")), float64(nbtconv.Map[float32](data, "Pitch")), float64(nbtconv.Map[float32](data, "Damage")), nil, potion.From(nbtconv.Map[int32](data, "auxValue")-1))
	arr.vel = nbtconv.MapVec3(data, "Motion")
	arr.disallowPickup = nbtconv.Map[byte](data, "player") == 0
	arr.obtainArrowOnPickup = nbtconv.Map[byte](data, "isCreative") == 1
	arr.fireTicks = int64(nbtconv.Map[int16](data, "Fire"))
	arr.punchLevel = int(nbtconv.Map[byte](data, "enchantPunch"))
	if _, ok := data["StuckToBlockPos"]; ok {
		arr.collisionPos = nbtconv.MapPos(data, "StuckToBlockPos")
		arr.collided = true
	}
	return arr
}

// EncodeNBT encodes the Arrow entity's properties as a map and returns it.
func (a *Arrow) EncodeNBT() map[string]any {
	yaw, pitch := a.Rotation()
	data := map[string]any{
		"Pos":          nbtconv.Vec3ToFloat32Slice(a.Position()),
		"Yaw":          float32(yaw),
		"Pitch":        float32(pitch),
		"Motion":       nbtconv.Vec3ToFloat32Slice(a.Velocity()),
		"Damage":       float32(a.BaseDamage()),
		"Fire":         int16(a.OnFireDuration() * 20),
		"enchantPunch": byte(a.punchLevel),
		"auxValue":     int32(a.tip.Uint8() + 1),
		"player":       boolByte(!a.disallowPickup),
		"isCreative":   boolByte(!a.obtainArrowOnPickup),
	}
	// TODO: Save critical flag if Minecraft ever saves it?
	if collisionPos, ok := a.CollisionPos(); ok {
		data["StuckToBlockPos"] = nbtconv.PosToInt32Slice(collisionPos)
	}
	return data
}

// checkNearby checks for nearby arrow collectors and closes the Arrow if one was found and when the Arrow can be
// picked up.
func (a *Arrow) checkNearby(w *world.World) {
	grown := a.BBox().GrowVec3(mgl64.Vec3{1, 0.5, 1}).Translate(a.pos)
	ignore := func(e world.Entity) bool {
		return e == a
	}
	for _, e := range w.EntitiesWithin(a.BBox().Translate(a.pos).Grow(2), ignore) {
		if e.BBox().Translate(e.Position()).IntersectsWith(grown) {
			if collector, ok := e.(Collector); ok {
				if a.obtainArrowOnPickup {
					// A collector was within range to pick up the entity.
					for _, viewer := range w.Viewers(a.pos) {
						viewer.ViewEntityAction(a, PickedUpAction{Collector: collector})
					}
					_ = collector.Collect(item.NewStack(item.Arrow{Tip: a.tip}, 1))
				}
				a.close = true
				return
			}
		}
	}
}

// damage returns the full damage the arrow should deal. In Bedrock, this is the initial velocity length multiplied by
// base damage and then multiplied by 0.6. If the arrow is critical, it is also multiplied by 1.5.
func (a *Arrow) damage(vel mgl64.Vec3) float64 {
	base := math.Ceil(vel.Len() * a.BaseDamage())
	if a.Critical() {
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

package entity

import (
	"iter"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/loot"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// FishingHookBehaviourConfig holds optional parameters for a
// FishingHookBehaviour.
type FishingHookBehaviourConfig struct {
	// Owner is the entity that cast the fishing hook. A fishing hook without an
	// owner, or with an owner that stops fishing, removes itself.
	Owner *world.EntityHandle
	// Lure is the level of the Lure enchantment of the fishing rod used to
	// cast the hook. Every level reduces the time until a fish is lured by
	// 5 seconds.
	Lure int
	// Luck is the level of the Luck of the Sea enchantment of the fishing rod
	// used to cast the hook. It increases the chance of catching treasure and
	// decreases the chance of catching junk.
	Luck int
}

func (conf FishingHookBehaviourConfig) Apply(data *world.EntityData) {
	data.Data = conf.New()
}

// New creates a FishingHookBehaviour using the parameters in conf.
func (conf FishingHookBehaviourConfig) New() *FishingHookBehaviour {
	return &FishingHookBehaviour{
		BaseBehaviour: NewBaseBehaviour(),
		conf:          conf,
		mc:            &MovementComputer{},
	}
}

// fishingHookState is the state a fishing hook is in.
type fishingHookState int

const (
	// fishingHookFlying is the state of a hook that was cast and is flying
	// through the air or lying on the ground.
	fishingHookFlying fishingHookState = iota
	// fishingHookHookedInEntity is the state of a hook that is attached to an
	// entity.
	fishingHookHookedInEntity
	// fishingHookBobbing is the state of a hook that is floating in water.
	fishingHookBobbing
)

const (
	// fishingHookGravity is the Y velocity subtracted from the hook every tick
	// it is not in water.
	fishingHookGravity = 0.03
	// fishingHookDrag is the multiplier applied to the velocity of the hook
	// every tick.
	fishingHookDrag = 0.92
	// fishingHookMaxDistanceSquared is the squared distance between the hook
	// and its owner beyond which the line snaps.
	fishingHookMaxDistanceSquared = 32 * 32
	// fishingHookMaxGroundTicks is the amount of ticks a hook may stay on the
	// ground before it is removed.
	fishingHookMaxGroundTicks = 1200
)

// FishingHookBehaviour implements the behaviour of a fishing hook. It closely
// follows the vanilla implementation: The hook flies through the air until it
// lands in water, where it bobs until a fish is lured, approaches and bites.
// Reeling the hook in while a fish is biting catches it. A hook that hits an
// entity attaches to it, pulling it towards the owner when reeled in.
type FishingHookBehaviour struct {
	BaseBehaviour

	conf FishingHookBehaviourConfig
	mc   *MovementComputer

	state  fishingHookState
	hooked *world.EntityHandle
	closed bool

	groundTicks, outOfWaterTicks int

	nibble, timeUntilLured, timeUntilHooked int
	fishAngle                               float64
	biting                                  bool
}

// Owner returns the entity that cast the fishing hook.
func (f *FishingHookBehaviour) Owner() *world.EntityHandle {
	return f.conf.Owner
}

// HookedEntity returns the entity the fishing hook is attached to, or nil if
// it is not attached to any entity.
func (f *FishingHookBehaviour) HookedEntity() *world.EntityHandle {
	if f.state != fishingHookHookedInEntity {
		return nil
	}
	return f.hooked
}

// Biting returns true if a fish is currently biting the hook. Reeling in the
// hook while a fish is biting catches it.
func (f *FishingHookBehaviour) Biting() bool {
	return f.nibble > 0
}

// Tick moves the fishing hook and runs the logic for luring and catching fish.
func (f *FishingHookBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	if f.closed || f.shouldStopFishing(e, tx) {
		f.close(e)
		return nil
	}
	if f.mc.OnGround() {
		if f.groundTicks++; f.groundTicks >= fishingHookMaxGroundTicks {
			f.close(e)
			return nil
		}
	} else {
		f.groundTicks = 0
	}

	pos, vel := e.Position(), e.Velocity()
	blockPos := cube.PosFromVec3(pos)
	fluidHeight := waterHeight(tx, blockPos)
	inWater := fluidHeight > 0

	switch f.state {
	case fishingHookFlying:
		if inWater {
			e.data.Vel = mgl64.Vec3{vel[0] * 0.3, vel[1] * 0.2, vel[2] * 0.3}
			f.state = fishingHookBobbing
			return nil
		}
		if f.checkEntityHit(e, tx, pos, vel) {
			return f.tickHooked(e, tx)
		}
	case fishingHookHookedInEntity:
		return f.tickHooked(e, tx)
	case fishingHookBobbing:
		d := pos[1] + vel[1] - float64(blockPos[1]) - fluidHeight
		if math.Abs(d) < 0.01 {
			d += math.Copysign(0.1, d)
		}
		vel = mgl64.Vec3{vel[0] * 0.9, vel[1] - d*rand.Float64()*0.2, vel[2] * 0.9}

		if inWater {
			f.outOfWaterTicks = max(0, f.outOfWaterTicks-1)
			if f.biting {
				vel[1] -= 0.1 * rand.Float64() * rand.Float64()
			}
			e.data.Vel = vel
			f.catchFish(e, tx, blockPos)
			vel = e.data.Vel
		} else {
			f.outOfWaterTicks = min(10, f.outOfWaterTicks+1)
		}
	}

	if !inWater {
		vel[1] -= fishingHookGravity
	}
	viewers := tx.Viewers(pos)
	velBefore := e.Velocity()
	dPos, collidedVel := f.mc.CheckCollision(tx, e, pos, vel)
	if f.state == fishingHookFlying && (collidedVel != vel || f.mc.OnGround()) {
		// The hook hit a block while flying, so it stops moving.
		collidedVel = mgl64.Vec3{}
	}
	vel = collidedVel.Mul(fishingHookDrag)

	e.data.Pos, e.data.Vel = pos.Add(dPos), vel
	return &Movement{v: viewers, e: e,
		pos: e.data.Pos, vel: vel, dpos: dPos, dvel: vel.Sub(velBefore),
		rot: e.data.Rot, onGround: f.mc.OnGround(),
	}
}

// tickHooked moves the hook to the entity it is attached to. If the entity no
// longer exists, the hook starts falling again.
func (f *FishingHookBehaviour) tickHooked(e *Ent, tx *world.Tx) *Movement {
	hooked, ok := f.hooked.Entity(tx)
	if !ok {
		f.setHooked(e, tx, nil)
		return nil
	}
	pos := e.Position()
	target := hooked.Position().Add(mgl64.Vec3{0, hooked.H().Type().BBox(hooked).Height() * 0.8})
	e.data.Pos, e.data.Vel = target, mgl64.Vec3{}

	return &Movement{v: tx.Viewers(pos), e: e,
		pos: target, dpos: target.Sub(pos), rot: e.data.Rot,
	}
}

// checkEntityHit checks if the hook hits an entity while moving with the
// velocity passed. If so, the hook attaches to the entity and true is
// returned.
func (f *FishingHookBehaviour) checkEntityHit(e *Ent, tx *world.Tx, pos, vel mgl64.Vec3) bool {
	if mgl64.FloatEqual(vel.LenSqr(), 0) {
		return false
	}
	hit, ok := trace.Perform(pos, pos.Add(vel), tx, e.H().Type().BBox(e).Grow(1.0), f.ignores(e))
	if !ok {
		return false
	}
	r, ok := hit.(trace.EntityResult)
	if !ok {
		return false
	}
	f.setHooked(e, tx, r.Entity().H())
	return true
}

// setHooked attaches the hook to the entity passed, or detaches it if nil is
// passed.
func (f *FishingHookBehaviour) setHooked(e *Ent, tx *world.Tx, hooked *world.EntityHandle) {
	f.hooked = hooked
	if hooked == nil {
		f.state = fishingHookFlying
	} else {
		f.state = fishingHookHookedInEntity
	}
	for _, v := range tx.Viewers(e.Position()) {
		v.ViewEntityState(e)
	}
}

// ignores returns a filter for the entities that a fishing hook cannot attach
// to: Its owner, the hook itself, spectators and entities that are neither
// living entities nor items.
func (f *FishingHookBehaviour) ignores(e *Ent) trace.EntityFilter {
	return func(seq iter.Seq[world.Entity]) iter.Seq[world.Entity] {
		return func(yield func(world.Entity) bool) {
			for other := range seq {
				g, ok := other.(interface{ GameMode() world.GameMode })
				spectator := ok && !g.GameMode().HasCollision()
				_, living := other.(Living)
				isItem := other.H().Type() == ItemType
				if spectator || other.H() == e.H() || other.H() == f.conf.Owner || (!living && !isItem) {
					continue
				}
				if !yield(other) {
					return
				}
			}
		}
	}
}

// catchFish runs the logic for luring fish towards the hook. It is called
// every tick that the hook is bobbing in water.
func (f *FishingHookBehaviour) catchFish(e *Ent, tx *world.Tx, blockPos cube.Pos) {
	pos := e.Position()
	speed, above := 1, blockPos.Side(cube.FaceUp)
	if tx.RainingAt(above) && rand.IntN(4) == 0 {
		speed++
	}
	if tx.HighestLightBlocker(above[0], above[2]) >= above[1] && rand.IntN(2) == 0 {
		// Fishing takes longer if the hook cannot see the sky.
		speed--
	}

	switch {
	case f.nibble > 0:
		if f.nibble--; f.nibble <= 0 {
			// The fish got away.
			f.timeUntilLured, f.timeUntilHooked, f.biting = 0, 0, false
		}
	case f.timeUntilHooked > 0:
		if f.timeUntilHooked -= speed; f.timeUntilHooked > 0 {
			// A fish is approaching the hook, leaving a trail of bubbles.
			f.fishAngle += 9.188 * (rand.Float64() - rand.Float64())
			sin, cos := math.Sincos(mgl64.DegToRad(f.fishAngle))
			dist := float64(f.timeUntilHooked) * 0.1
			trail := mgl64.Vec3{pos[0] + sin*dist, math.Floor(pos[1]) + 1, pos[2] + cos*dist}
			if waterHeight(tx, cube.PosFromVec3(trail.Sub(mgl64.Vec3{0, 1}))) > 0 && rand.Float64() < 0.85 {
				tx.AddParticle(trail.Sub(mgl64.Vec3{0, 0.1}), particle.Bubble{})
			}
			return
		}
		// The fish bites the hook, pulling it underwater.
		e.data.Vel[1] = -0.4 * (0.6 + rand.Float64()*0.4)
		f.nibble, f.biting = 20+rand.IntN(21), true
		tx.PlaySound(pos, sound.Splash{})
		for range 5 {
			tx.AddParticle(mgl64.Vec3{pos[0] + rand.Float64()*0.5 - 0.25, math.Floor(pos[1]) + 0.9, pos[2] + rand.Float64()*0.5 - 0.25}, particle.Bubble{})
		}
		for _, v := range tx.Viewers(pos) {
			v.ViewEntityAction(e, FishingHookBiteAction{})
		}
	case f.timeUntilLured > 0:
		f.timeUntilLured -= speed
		chance := 0.15
		switch {
		case f.timeUntilLured < 20:
			chance += float64(20-f.timeUntilLured) * 0.05
		case f.timeUntilLured < 40:
			chance += float64(40-f.timeUntilLured) * 0.02
		case f.timeUntilLured < 60:
			chance += float64(60-f.timeUntilLured) * 0.01
		}
		if rand.Float64() < chance {
			sin, cos := math.Sincos(rand.Float64() * 2 * math.Pi)
			dist := (25 + rand.Float64()*35) * 0.1
			splash := mgl64.Vec3{pos[0] + sin*dist, math.Floor(pos[1]) + 1, pos[2] + cos*dist}
			if waterHeight(tx, cube.PosFromVec3(splash.Sub(mgl64.Vec3{0, 1}))) > 0 {
				tx.AddParticle(splash.Sub(mgl64.Vec3{0, 0.1}), particle.Bubble{})
			}
		}
		if f.timeUntilLured <= 0 {
			f.fishAngle = rand.Float64() * 360
			f.timeUntilHooked = 20 + rand.IntN(61)
		}
	default:
		f.timeUntilLured = 100 + rand.IntN(501) - f.conf.Lure*100
	}
}

// Reel reels in the fishing hook, closing it. If a fish was biting, loot is
// caught and launched towards the owner. If the hook was attached to an
// entity, that entity is pulled towards the owner. The durability damage that
// the fishing rod should take is returned.
func (f *FishingHookBehaviour) Reel(e *Ent, tx *world.Tx) (damage int) {
	defer f.close(e)
	owner, ok := f.conf.Owner.Entity(tx)
	if !ok || f.closed {
		return 0
	}
	pos, ownerPos := e.Position(), owner.Position()

	if hooked, ok := f.hooked.Entity(tx); ok && f.state == fishingHookHookedInEntity {
		if v, ok := hooked.(interface {
			Velocity() mgl64.Vec3
			SetVelocity(mgl64.Vec3)
		}); ok {
			v.SetVelocity(v.Velocity().Add(ownerPos.Sub(hooked.Position()).Mul(0.1)))
		}
		damage = 5
		if hooked.H().Type() == ItemType {
			damage = 3
		}
	} else if f.nibble > 0 {
		d := ownerPos.Sub(pos)
		vel := mgl64.Vec3{d[0] * 0.1, d[1]*0.1 + math.Sqrt(d.Len())*0.08, d[2] * 0.1}
		conf := tx.World().EntityRegistry().Config()
		for _, s := range fishingLoot(tx, cube.PosFromVec3(pos), f.conf.Luck) {
			tx.AddEntity(conf.Item(world.EntitySpawnOpts{Position: pos, Velocity: vel}, s))
		}
		tx.AddEntity(NewExperienceOrb(world.EntitySpawnOpts{Position: ownerPos.Add(mgl64.Vec3{0, 0.5, 0.5})}, 1+rand.IntN(6)))
		damage = 1
	}
	if f.mc.OnGround() {
		damage = 2
	}
	return damage
}

// fishingLoot generates the loot caught by a fishing hook at the position passed with the luck passed, using
// the vanilla fishing loot tables. Fishing in a jungle uses the jungle loot table.
func fishingLoot(tx *world.Tx, pos cube.Pos, luck int) []item.Stack {
	path := loot.Fishing
	if slices.Contains(tx.Biome(pos).Tags(), "jungle") {
		path = loot.JungleFishing
	}
	t, _ := loot.Lookup(path)
	return t.Generate(loot.Context{Luck: float64(luck)})
}

// shouldStopFishing checks if the owner of the hook stopped fishing: If it
// no longer exists, is dead, cast another hook, no longer holds a fishing rod or
// is too far away from the hook.
func (f *FishingHookBehaviour) shouldStopFishing(e *Ent, tx *world.Tx) bool {
	owner, ok := f.conf.Owner.Entity(tx)
	if !ok {
		return true
	}
	if l, ok := owner.(Living); ok && l.Dead() {
		return true
	}
	if a, ok := owner.(interface{ FishingHook() *world.EntityHandle }); ok && a.FishingHook() != e.H() {
		return true
	}
	c, ok := owner.(interface {
		HeldItems() (mainHand, offHand item.Stack)
	})
	if !ok {
		return true
	}
	main, off := c.HeldItems()
	_, mainRod := main.Item().(item.FishingRod)
	_, offRod := off.Item().(item.FishingRod)
	if !mainRod && !offRod {
		return true
	}
	return owner.Position().Sub(e.Position()).LenSqr() > fishingHookMaxDistanceSquared
}

// close closes the fishing hook and clears it from its owner.
func (f *FishingHookBehaviour) close(e *Ent) {
	if f.closed {
		return
	}
	f.closed = true
	if owner, ok := f.conf.Owner.Entity(e.tx); ok {
		if a, ok := owner.(interface {
			FishingHook() *world.EntityHandle
			SetFishingHook(*world.EntityHandle)
		}); ok && a.FishingHook() == e.H() {
			a.SetFishingHook(nil)
		}
	}
	_ = e.Close()
}

// waterHeight returns the height of the water surface within the block at the
// position passed, in the range 0-1. 0 is returned if there is no water at the
// position.
func waterHeight(tx *world.Tx, pos cube.Pos) float64 {
	l, ok := tx.Liquid(pos)
	if !ok {
		return 0
	}
	w, ok := l.(block.Water)
	if !ok {
		return 0
	}
	if w.Falling {
		return 1
	}
	if above, ok := tx.Liquid(pos.Side(cube.FaceUp)); ok {
		if _, ok := above.(block.Water); ok {
			return 1
		}
	}
	return float64(w.Depth) / 9
}

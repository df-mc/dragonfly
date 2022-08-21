package entity

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/entity/healing"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"time"
)

// Item represents an item entity which may be added to the world. Players and several humanoid entities such
// as zombies are able to pick up these entities so that the items are added to their inventory.
type Item struct {
	transform

	age, pickupDelay int
	fireTicks        int64

	i item.Stack

	h *HealthManager
	c *MovementComputer
}

// NewItem creates a new item entity using the item stack passed. The item entity will be positioned at the
// position passed.
// If the stack's count exceeds its max count, the count of the stack will be changed to the maximum.
func NewItem(i item.Stack, pos mgl64.Vec3) *Item {
	if i.Count() > i.MaxCount() {
		i = i.Grow(i.MaxCount() - i.Count())
	}
	i = nbtconv.ReadItem(nbtconv.WriteItem(i, true), nil)

	it := &Item{
		i:           i,
		pickupDelay: 10,

		h: NewHealthManager(5, 5),
		c: &MovementComputer{
			Gravity:           0.04,
			Drag:              0.02,
			DragBeforeGravity: true,
		},
	}
	it.transform = newTransform(it, pos)
	return it
}

// Name ...
func (it *Item) Name() string {
	return fmt.Sprintf("%T", it.i.Item())
}

// EncodeEntity ...
func (it *Item) EncodeEntity() string {
	return "minecraft:item"
}

// BBox ...
func (it *Item) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Item returns the item stack that the item entity holds.
func (it *Item) Item() item.Stack {
	return it.i
}

// SetPickupDelay sets a delay passed until the item can be picked up. If d is negative or d.Seconds()*20
// higher than math.MaxInt16, the item will never be able to be picked up.
func (it *Item) SetPickupDelay(d time.Duration) {
	ticks := int(d.Seconds() * 20)
	if ticks < 0 || ticks >= math.MaxInt16 {
		ticks = math.MaxInt16
	}
	it.pickupDelay = ticks
}

// Health ...
func (it *Item) Health() float64 {
	return it.h.Health()
}

// MaxHealth ...
func (it *Item) MaxHealth() float64 {
	return it.h.MaxHealth()
}

// SetMaxHealth ...
func (it *Item) SetMaxHealth(v float64) {
	it.h.SetMaxHealth(v)
}

// Dead ...
func (it *Item) Dead() bool {
	return it.Health() <= 0
}

// AttackImmune ...
func (it *Item) AttackImmune() bool {
	return false
}

// fireSource represents a damage.Source that emits fire.
type fireSource interface {
	Fire() bool
}

// Hurt ...
func (it *Item) Hurt(dmg float64, source damage.Source) (n float64, vulnerable bool) {
	if dmg < 0 {
		return 0, true
	}
	if f, ok := source.(fireSource); ok && f.Fire() && it.FireProof() {
		return 0, false
	}

	it.h.AddHealth(-dmg)
	if it.Dead() {
		_ = it.Close()
	}
	return dmg, true
}

// Heal ...
func (it *Item) Heal(health float64, _ healing.Source) {
	if health < 0 {
		return
	}
	it.h.AddHealth(health)
}

// OnFireDuration ...
func (it *Item) OnFireDuration() time.Duration {
	it.mu.Lock()
	defer it.mu.Unlock()
	return time.Duration(it.fireTicks) * time.Second / 20
}

// SetOnFire ...
func (it *Item) SetOnFire(duration time.Duration) {
	if it.FireProof() {
		return
	}

	it.mu.Lock()
	it.fireTicks = int64(duration.Seconds() * 20)
	it.mu.Unlock()
	for _, v := range it.World().Viewers(it.Position()) {
		v.ViewEntityState(it)
	}
}

// Extinguish ...
func (it *Item) Extinguish() {
	it.SetOnFire(0)
}

// armourTiered represents an item.Armour that has an ArmourTier.
type armourTiered interface {
	item.Armour
	ArmourTier() item.ArmourTier
}

// toolTiered represents an item.Tool that has a ToolTier.
type toolTiered interface {
	item.Tool
	ToolTier() item.ToolTier
}

// FireProof ...
func (it *Item) FireProof() bool {
	switch i := it.Item().Item().(type) {
	case armourTiered:
		if _, ok := i.ArmourTier().(item.ArmourTierNetherite); ok {
			return true
		}
	case toolTiered:
		return i.ToolTier() == item.ToolTierNetherite
	}
	return false
}

// Tick ticks the entity, performing movement.
func (it *Item) Tick(w *world.World, current int64) {
	it.mu.Lock()
	m := it.c.TickMovement(it, it.pos, it.vel, 0, 0)
	it.pos, it.vel = m.pos, m.vel
	it.mu.Unlock()

	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		_ = it.Close()
		return
	}
	if it.age++; it.age > 6000 {
		_ = it.Close()
		return
	}

	if it.pickupDelay == 0 {
		it.checkNearby(w, m.pos)
	} else if it.pickupDelay != math.MaxInt16 {
		it.pickupDelay--
	}

	it.tickFire()
	it.checkBlockCollisions(w)
}

// checkNearby checks the entities of the chunks around for item collectors and other item stacks. If a
// collector is found in range, the item will be picked up. If another item stack with the same item type is
// found in range, the item stacks will merge.
func (it *Item) checkNearby(w *world.World, pos mgl64.Vec3) {
	grown := it.BBox().GrowVec3(mgl64.Vec3{1, 0.5, 1}).Translate(pos)
	for _, e := range w.EntitiesWithin(it.BBox().Translate(pos).Grow(2), nil) {
		if e == it {
			// Skip the item entity itself.
			continue
		}
		if e.BBox().Translate(e.Position()).IntersectsWith(grown) {
			if collector, ok := e.(Collector); ok {
				// A collector was within range to pick up the entity.
				it.collect(w, collector, pos)
				return
			} else if other, ok := e.(*Item); ok {
				// Another item entity was in range to merge with.
				if it.merge(w, other, pos) {
					return
				}
			}
		}
	}
}

// entityInsider represents a block that reacts to an entity going inside its 1x1x1 axis aligned bounding box.
type entityInsider interface {
	// EntityInside is called when an entity goes inside the block's 1x1x1 axis aligned bounding box.
	EntityInside(pos cube.Pos, w *world.World, e world.Entity)
}

// checkCollisions checks the item's block collisions.
func (it *Item) checkBlockCollisions(w *world.World) {
	box := it.BBox().Translate(it.Position()).Grow(-0.0001)
	min, max := cube.PosFromVec3(box.Min()), cube.PosFromVec3(box.Max())

	for y := min[1]; y <= max[1]; y++ {
		for x := min[0]; x <= max[0]; x++ {
			for z := min[2]; z <= max[2]; z++ {
				blockPos := cube.Pos{x, y, z}
				b := w.Block(blockPos)
				if collide, ok := b.(entityInsider); ok {
					collide.EntityInside(blockPos, w, it)
					if _, liquid := b.(world.Liquid); liquid {
						continue
					}
				}

				if l, ok := w.Liquid(blockPos); ok {
					if collide, ok := l.(entityInsider); ok {
						collide.EntityInside(blockPos, w, it)
					}
				}
			}
		}
	}
}

// tickFire ticks fire on the entity if it is on fire.
func (it *Item) tickFire() {
	if it.OnFireDuration() <= 0 {
		return
	}

	it.mu.Lock()
	it.fireTicks--
	it.mu.Unlock()

	if it.OnFireDuration() <= 0 {
		it.Extinguish()
	}
	if it.OnFireDuration()%time.Second == 0 {
		it.Hurt(1, damage.SourceFireTick{})
	}
}

// merge merges the item entity with another item entity.
func (it *Item) merge(w *world.World, other *Item, pos mgl64.Vec3) bool {
	if other.i.Count() == other.i.MaxCount() || it.i.Count() == it.i.MaxCount() {
		// Either stack is already filled up to the maximum, meaning we can't change anything any way.
		return false
	}
	if !it.i.Comparable(other.i) {
		return false
	}

	a, b := other.i.AddStack(it.i)

	newA := NewItem(a, other.Position())
	newA.SetVelocity(other.Velocity())
	w.AddEntity(newA)

	if !b.Empty() {
		newB := NewItem(b, pos)
		newB.SetVelocity(it.vel)
		w.AddEntity(newB)
	}
	_ = it.Close()
	_ = other.Close()
	return true
}

// collect makes a collector collect the item (or at least part of it).
func (it *Item) collect(w *world.World, collector Collector, pos mgl64.Vec3) {
	n := collector.Collect(it.i)
	if n == 0 {
		return
	}
	for _, viewer := range w.Viewers(pos) {
		viewer.ViewEntityAction(it, PickedUpAction{Collector: collector})
	}

	if n == it.i.Count() {
		// The collector picked up the entire stack.
		_ = it.Close()
		return
	}
	// Create a new item entity and shrink it by the amount of items that the collector collected.
	w.AddEntity(NewItem(it.i.Grow(-n), pos))

	_ = it.Close()
}

// New creates and returns an Item with the item.Stack, position, and velocity provided. It doesn't spawn the Item
// by itself.
func (it *Item) New(stack item.Stack, pos, vel mgl64.Vec3) world.Entity {
	itemEntity := NewItem(stack, pos)
	itemEntity.vel = vel
	return itemEntity
}

// Explode ...
func (it *Item) Explode(mgl64.Vec3, float64, block.ExplosionConfig) {
	_ = it.Close()
}

// DecodeNBT decodes the properties in a map to an Item and returns a new Item entity.
func (it *Item) DecodeNBT(data map[string]any) any {
	i := nbtconv.MapItem(data, "Item")
	if i.Empty() {
		return nil
	}
	n := NewItem(i, nbtconv.MapVec3(data, "Pos"))
	n.SetVelocity(nbtconv.MapVec3(data, "Motion"))
	n.age = int(nbtconv.Map[int16](data, "Age"))
	n.pickupDelay = int(nbtconv.Map[int64](data, "PickupDelay"))
	return n
}

// EncodeNBT encodes the Item entity's properties as a map and returns it.
func (it *Item) EncodeNBT() map[string]any {
	return map[string]any{
		"Health":      int16(5),
		"Age":         int16(it.age),
		"PickupDelay": int64(it.pickupDelay),
		"Pos":         nbtconv.Vec3ToFloat32Slice(it.Position()),
		"Motion":      nbtconv.Vec3ToFloat32Slice(it.Velocity()),
		"Item":        nbtconv.WriteItem(it.Item(), true),
	}
}

// Collector represents an entity in the world that is able to collect an item, typically an entity such as
// a player or a zombie.
type Collector interface {
	world.Entity
	// Collect collects the stack passed. It is called if the Collector is standing near an item entity that
	// may be picked up.
	// The count of items collected from the stack n is returned.
	Collect(stack item.Stack) (n int)
	// GameMode returns the gamemode of the collector.
	GameMode() world.GameMode
}

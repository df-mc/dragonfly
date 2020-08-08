package entity

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/action"
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/entity/state"
	"github.com/df-mc/dragonfly/dragonfly/internal/nbtconv"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"sync/atomic"
	"time"
)

// Item represents an item entity which may be added to the world. Players and several humanoid entities such
// as zombies are able to pick up these entities so that the items are added to their inventory.
type Item struct {
	age, pickupDelay int
	i                item.Stack
	velocity, pos    atomic.Value

	*movementComputer
}

// NewItem creates a new item entity using the item stack passed. The item entity will be positioned at the
// position passed.
// If the stack's count exceeds its max count, the count of the stack will be changed to the maximum.
func NewItem(i item.Stack, pos mgl64.Vec3) *Item {
	if i.Count() > i.MaxCount() {
		i = i.Grow(i.Count() - i.MaxCount())
	}
	i = nbtconv.ItemFromNBT(nbtconv.ItemToNBT(i, false), nil)

	it := &Item{i: i, movementComputer: &movementComputer{
		gravity:           0.04,
		dragBeforeGravity: true,
	}}
	it.SetPickupDelay(time.Second / 2)
	it.pos.Store(pos)
	it.velocity.Store(mgl64.Vec3{})

	return it
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

// Position returns the current position of the item entity.
func (it *Item) Position() mgl64.Vec3 {
	return it.pos.Load().(mgl64.Vec3)
}

// World returns the world that the item entity is currently in, or nil if it is not added to a world.
func (it *Item) World() *world.World {
	w, _ := world.OfEntity(it)
	return w
}

// Tick ticks the entity, performing movement.
func (it *Item) Tick(current int64) {
	if it.Position()[1] < 0 && current%10 == 0 {
		_ = it.Close()
		return
	}
	if it.age++; it.age > 6000 {
		_ = it.Close()
		return
	}
	it.pos.Store(it.tickMovement(it))

	if it.pickupDelay == 0 {
		it.checkNearby()
	} else if it.pickupDelay != math.MaxInt16 {
		it.pickupDelay--
	}
}

// checkNearby checks the entities of the chunks around for item collectors and other item stacks. If a
// collector is found in range, the item will be picked up. If another item stack with the same item type is
// found in range, the item stacks will merge.
func (it *Item) checkNearby() {
	grown := it.AABB().GrowVec3(mgl64.Vec3{1, 0.5, 1}).Translate(it.Position())
	for _, e := range it.World().EntitiesWithin(it.AABB().Translate(it.Position()).Grow(2)) {
		if e == it {
			// Skip the item entity itself.
			continue
		}
		if e.AABB().Translate(e.Position()).IntersectsWith(grown) {
			if collector, ok := e.(item.Collector); ok {
				// A collector was within range to pick up the entity.
				it.collect(collector)
				return
			} else if other, ok := e.(*Item); ok {
				// Another item entity was in range to merge with.
				if it.merge(other) {
					return
				}
			}
		}
	}
}

// merge merges the item entity with another item entity.
func (it *Item) merge(other *Item) bool {
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
	it.World().AddEntity(newA)

	if !b.Empty() {
		newB := NewItem(b, it.Position())
		newB.SetVelocity(it.Velocity())
		it.World().AddEntity(newB)
	}
	_ = it.Close()
	_ = other.Close()
	return true
}

// collect makes a collector collect the item (or at least part of it).
func (it *Item) collect(collector item.Collector) {
	n := collector.Collect(it.i)
	if n == 0 {
		return
	}
	for _, viewer := range it.World().Viewers(it.Position()) {
		viewer.ViewEntityAction(it, action.PickedUp{Collector: collector})
	}

	if n == it.i.Count() {
		// The collector picked up the entire stack.
		_ = it.Close()
		return
	}
	// Create a new item entity and shrink it by the amount of items that the collector collected.
	it.World().AddEntity(NewItem(it.i.Grow(-n), it.Position()))

	_ = it.Close()
}

// Velocity returns the current velocity of the item. The values in the Vec3 returned represent the speed on
// that axis in blocks/tick.
func (it *Item) Velocity() mgl64.Vec3 {
	return it.velocity.Load().(mgl64.Vec3)
}

// SetVelocity sets the velocity of the item entity. The values in the Vec3 passed represent the speed on
// that axis in blocks/tick.
func (it *Item) SetVelocity(v mgl64.Vec3) {
	it.velocity.Store(v)
}

// Yaw always returns 0.
func (it *Item) Yaw() float64 { return 0 }

// Pitch always returns 0.
func (it *Item) Pitch() float64 { return 0 }

// AABB ...
func (it *Item) AABB() physics.AABB {
	return physics.NewAABB(mgl64.Vec3{-0.125, 0, -0.125}, mgl64.Vec3{0.125, 0.25, 0.125})
}

// State ...
func (it *Item) State() []state.State {
	return nil
}

// Close closes the item, removing it from the world that it is currently in.
func (it *Item) Close() error {
	if it.World() != nil {
		it.World().RemoveEntity(it)
	}
	return nil
}

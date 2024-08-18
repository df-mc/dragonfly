package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"time"
)

// ItemBehaviourConfig holds optional parameters for an ItemBehaviour.
type ItemBehaviourConfig struct {
	// Gravity is the amount of Y velocity subtracted every tick.
	Gravity float64
	// Drag is used to reduce all axes of the velocity every tick. Velocity is
	// multiplied with (1-Drag) every tick.
	Drag float64
	// ExistenceDuration specifies how long the item stack should last. The
	// default is time.Minute * 5.
	ExistenceDuration time.Duration
	// PickupDelay specifies how much time must expire before the item can be
	// picked up by collectors. The default is time.Second / 2.
	PickupDelay time.Duration
}

// New creates an ItemBehaviour using i and the optional parameters in conf.
func (conf ItemBehaviourConfig) New(i item.Stack) *ItemBehaviour {
	if i.Count() > i.MaxCount() {
		i = i.Grow(i.MaxCount() - i.Count())
	}
	i = nbtconv.Item(nbtconv.WriteItem(i, true), nil)

	if conf.PickupDelay == 0 {
		conf.PickupDelay = time.Second / 2
	}
	if conf.ExistenceDuration == 0 {
		conf.ExistenceDuration = time.Minute * 5
	}

	b := &ItemBehaviour{conf: conf, i: i, pickupDelay: conf.PickupDelay}
	b.passive = PassiveBehaviourConfig{
		Gravity:           conf.Gravity,
		Drag:              conf.Drag,
		ExistenceDuration: conf.ExistenceDuration,
		Tick:              b.tick,
	}.New()
	return b
}

// ItemBehaviour implements the behaviour of item entities.
type ItemBehaviour struct {
	conf    ItemBehaviourConfig
	passive *PassiveBehaviour
	i       item.Stack

	pickupDelay time.Duration
}

// Item returns the item.Stack held by the entity.
func (i *ItemBehaviour) Item() item.Stack {
	return i.i
}

// Tick moves the entity, checks if it should be picked up by a nearby collector
// or if it should merge with nearby item entities.
func (i *ItemBehaviour) Tick(e *Ent) *Movement {
	w := e.World()
	pos := cube.PosFromVec3(e.Position())
	blockPos := pos.Side(cube.FaceDown)

	bl, ok := w.Block(blockPos).(block.Hopper)
	if ok && !bl.Powered && bl.CollectCooldown <= 0 {
		addedCount, err := bl.Inventory(w, blockPos).AddItem(i.i)
		if err != nil {
			if addedCount == 0 {
				return i.passive.Tick(e)
			}

			// This is only reached if part of the item stack was collected into the hopper.
			w.AddEntity(NewItem(i.Item().Grow(-addedCount), pos.Vec3Centre()))
		}

		_ = e.Close()
		bl.CollectCooldown = 8
		w.SetBlock(blockPos, bl, nil)
	}
	return i.passive.Tick(e)
}

// tick checks if the item can be picked up or merged with nearby item stacks.
func (i *ItemBehaviour) tick(e *Ent) {
	if i.pickupDelay == 0 {
		i.checkNearby(e)
	} else if i.pickupDelay < math.MaxInt16*(time.Second/20) {
		i.pickupDelay -= time.Second / 20
	}
}

// checkNearby checks the nearby entities for item collectors and other item
// stacks. If a collector is found in range, the item will be picked up. If
// another item stack with the same item type is found in range, the item
// stacks will merge.
func (i *ItemBehaviour) checkNearby(e *Ent) {
	w, pos := e.World(), e.Position()
	bbox := e.Type().BBox(e)
	grown := bbox.GrowVec3(mgl64.Vec3{1, 0.5, 1}).Translate(pos)
	nearby := w.EntitiesWithin(bbox.Translate(pos).Grow(2), func(entity world.Entity) bool {
		return entity == e
	})
	for _, other := range nearby {
		if !other.Type().BBox(other).Translate(other.Position()).IntersectsWith(grown) {
			continue
		}
		if collector, ok := other.(Collector); ok {
			// A collector was within range to pick up the entity.
			i.collect(e, collector)
			return
		} else if _, ok := other.Type().(ItemType); ok {
			// Another item entity was in range to merge with.
			if i.merge(e, other.(*Ent)) {
				return
			}
		}
	}
}

// merge merges the item entity with another item entity.
func (i *ItemBehaviour) merge(e *Ent, other *Ent) bool {
	w, pos := e.World(), e.Position()
	otherBehaviour := other.Behaviour().(*ItemBehaviour)
	if otherBehaviour.i.Count() == otherBehaviour.i.MaxCount() || i.i.Count() == i.i.MaxCount() || !i.i.Comparable(otherBehaviour.i) {
		// Either stack is already filled up to the maximum, meaning we can't
		// change anything any way, other the stack types weren't comparable.
		return false
	}
	a, b := otherBehaviour.i.AddStack(i.i)

	newA := NewItem(a, other.Position())
	newA.SetVelocity(other.Velocity())
	w.AddEntity(newA)

	if !b.Empty() {
		newB := NewItem(b, pos)
		newB.SetVelocity(e.Velocity())
		w.AddEntity(newB)
	}
	_ = e.Close()
	_ = other.Close()
	return true
}

// collect makes a collector collect the item (or at least part of it).
func (i *ItemBehaviour) collect(e *Ent, collector Collector) {
	w, pos := e.World(), e.Position()
	n := collector.Collect(i.i)
	if n == 0 {
		return
	}
	for _, viewer := range w.Viewers(pos) {
		viewer.ViewEntityAction(e, PickedUpAction{Collector: collector})
	}

	if n == i.i.Count() {
		// The collector picked up the entire stack.
		_ = e.Close()
		return
	}
	// Create a new item entity and shrink it by the amount of items that the
	// collector collected.
	w.AddEntity(NewItem(i.i.Grow(-n), pos))
	_ = e.Close()
}

// Collector represents an entity in the world that is able to collect an item, typically an entity such as
// a player or a zombie.
type Collector interface {
	world.Entity
	// Collect collects the stack passed. It is called if the Collector is standing near an item entity that
	// may be picked up.
	// The count of items collected from the stack n is returned.
	Collect(stack item.Stack) (n int)
}

package entity

import (
	"math"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// ItemBehaviourConfig holds optional parameters for an ItemBehaviour.
type ItemBehaviourConfig struct {
	Item item.Stack
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

func (conf ItemBehaviourConfig) Apply(data *world.EntityData) {
	data.Data = conf.New()
}

// New creates an ItemBehaviour using i and the optional parameters in conf.
func (conf ItemBehaviourConfig) New() *ItemBehaviour {
	i := conf.Item
	if i.Count() > i.MaxCount() {
		i = i.Grow(i.MaxCount() - i.Count())
	}
	i = item.ReadNBT(item.WriteNBT(i, true), nil)

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

// PortalTravelComputer returns the interdimensional travel state for the behaviour.
func (i *ItemBehaviour) PortalTravelComputer() *PortalTravelComputer {
	return i.passive.PortalTravelComputer()
}

// Item returns the item.Stack held by the entity.
func (i *ItemBehaviour) Item() item.Stack {
	return i.i
}

// Tick moves the entity, checks if it should be picked up by a nearby collector
// or if it should merge with nearby item entities.
func (i *ItemBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	pos := cube.PosFromVec3(e.Position())
	blockPos := pos.Side(cube.FaceDown)

	bl, ok := tx.Block(blockPos).(block.Hopper)
	if ok && !bl.Powered && bl.CollectCooldown <= 0 {
		addedCount, err := bl.Inventory(tx, blockPos).AddItem(i.i)
		if err != nil {
			if addedCount == 0 {
				return i.passive.Tick(e, tx)
			}

			// This is only reached if part of the item stack was collected into the hopper.
			opts := world.EntitySpawnOpts{Position: pos.Vec3Centre()}
			tx.AddEntity(NewItem(opts, i.Item().Grow(-addedCount)))
		}

		_ = e.Close()
		bl.CollectCooldown = 8
		tx.SetBlock(blockPos, bl, nil)
		return nil
	}
	return i.passive.Tick(e, tx)
}

// Explode reacts to explosions. The item entity is destroyed, unless the item
// type is blast proof.
func (i *ItemBehaviour) Explode(e *Ent, _ world.ExplosionSource, impact float64) {
	if impact > 0 {
		if expl, ok := i.Item().Item().(interface{ BlastProof() bool }); ok && expl.BlastProof() {
			return
		}
		_ = e.Close()
	}
}

// tick checks if the item can be picked up or merged with nearby item stacks.
func (i *ItemBehaviour) tick(e *Ent, tx *world.Tx) {
	if i.pickupDelay <= 0 {
		i.checkNearby(e, tx)
		return
	}
	if i.pickupDelay < math.MaxInt16*(time.Second/20) {
		i.pickupDelay -= time.Second / 20
	}
}

// collectRadius is how far a collector may be to pick an item up, mergeRadius
// how far apart two item entities may be to merge.
var (
	collectRadius = mgl64.Vec3{1, 0.5, 1}
	mergeRadius   = mgl64.Vec3{0.5, 0, 0.5}
)

// checkNearby checks the nearby entities for item collectors and other item
// stacks. If a collector is found in range, the item will be picked up. If
// another item stack with the same item type is found in range, the item
// stacks will merge.
func (i *ItemBehaviour) checkNearby(e *Ent, tx *world.Tx) {
	bbox := e.H().Type().BBox(e).Translate(e.Position())
	collect, merge := bbox.GrowVec3(collectRadius), bbox.GrowVec3(mergeRadius)

	for other := range tx.EntitiesWithin(bbox.Grow(2)) {
		if e.H() == other.H() {
			continue
		}
		otherBBox := other.H().Type().BBox(other).Translate(other.Position())
		if collector, ok := other.(Collector); ok {
			if otherBBox.IntersectsWith(collect) {
				// A collector was within range to pick up the entity.
				i.collect(e, collector, tx)
				return
			}
		} else if other.H().Type() == ItemType && otherBBox.IntersectsWith(merge) {
			// Another item entity was in range to merge with.
			if i.merge(e, other.(*Ent), tx) {
				return
			}
		}
	}
}

// merge merges two item entities into one. Both stacks must fit into one and
// both items must have landed. The bigger stack is the one that stays, keeping
// the lower age and the higher pickup delay of the two.
func (i *ItemBehaviour) merge(e *Ent, other *Ent, tx *world.Tx) bool {
	otherBehaviour := other.Behaviour().(*ItemBehaviour)
	if !i.passive.mc.OnGround() || !otherBehaviour.passive.mc.OnGround() {
		return false
	}
	if !i.i.Comparable(otherBehaviour.i) || i.i.Count()+otherBehaviour.i.Count() > i.i.MaxCount() {
		return false
	}

	dst, dstEnt, src, srcEnt := otherBehaviour, other, i, e
	if otherBehaviour.i.Count() < i.i.Count() {
		dst, dstEnt, src, srcEnt = i, e, otherBehaviour, other
	}

	if srcEnt.data.Age < dstEnt.data.Age {
		dstEnt.data.Age = srcEnt.data.Age
	}
	if src.pickupDelay > dst.pickupDelay {
		dst.pickupDelay = src.pickupDelay
	}
	dst.setStack(dstEnt, dst.i.Grow(src.i.Count()), tx)

	_ = srcEnt.Close()
	return true
}

// setStack changes the item stack held by the entity. The stack only reaches
// the client when the entity is added, so it is sent to viewers again.
func (i *ItemBehaviour) setStack(e *Ent, s item.Stack, tx *world.Tx) {
	i.i = s
	for _, v := range tx.Viewers(e.Position()) {
		v.HideEntity(e)
		v.ViewEntity(e)
	}
}

// collect makes a collector collect the item (or at least part of it).
func (i *ItemBehaviour) collect(e *Ent, collector Collector, tx *world.Tx) {
	pos := e.Position()
	n, _ := collector.Collect(i.i)
	if n == 0 {
		return
	}
	for _, viewer := range tx.Viewers(pos) {
		viewer.ViewEntityAction(e, PickedUpAction{Collector: collector})
	}

	if n == i.i.Count() {
		// The collector picked up the entire stack.
		_ = e.Close()
		return
	}
	// Create a new item entity and shrink it by the amount of items that the
	// collector collected.
	tx.AddEntity(NewItem(world.EntitySpawnOpts{Position: pos}, i.i.Grow(-n)))
	_ = e.Close()
}

// Collector represents an entity in the world that is able to collect an item, typically an entity such as
// a player or a zombie.
type Collector interface {
	world.Entity
	// Collect collects the stack passed. It is called if the Collector is standing near an item entity that
	// may be picked up.
	// The count of items collected from the stack n is returned, along with a
	// bool that indicates if the Collector was in a state where it could
	// collect any items in the first place.
	Collect(stack item.Stack) (n int, ok bool)
}

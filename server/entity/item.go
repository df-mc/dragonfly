package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// NewItem creates a new item entity using the item stack passed. The item
// entity will be positioned at the position passed. If the stack's count
// exceeds its max count, the count of the stack will be changed to the
// maximum.
func NewItem(i item.Stack, pos mgl64.Vec3) *Ent {
	return Config{Behaviour: itemConf.New(i)}.New(ItemType{}, pos)
}

// NewItemPickupDelay creates a new item entity containing item stack i. A
// delay may be specified which defines for how long the item stack cannot be
// picked up from the ground.
func NewItemPickupDelay(i item.Stack, pos mgl64.Vec3, delay time.Duration) *Ent {
	config := itemConf
	config.PickupDelay = delay
	return Config{Behaviour: config.New(i)}.New(ItemType{}, pos)
}

var itemConf = ItemBehaviourConfig{
	Gravity: 0.04,
	Drag:    0.02,
}

// ItemType is a world.EntityType implementation for Item.
type ItemType struct{}

func (ItemType) EncodeEntity() string   { return "minecraft:item" }
func (ItemType) NetworkOffset() float64 { return 0.125 }
func (ItemType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (ItemType) DecodeNBT(m map[string]any) world.Entity {
	i := nbtconv.MapItem(m, "Item")
	if i.Empty() {
		return nil
	}
	n := NewItem(i, nbtconv.Vec3(m, "Pos"))
	n.SetVelocity(nbtconv.Vec3(m, "Motion"))
	n.age = time.Duration(nbtconv.Int16(m, "Age")) * (time.Second / 20)
	n.Behaviour().(*ItemBehaviour).pickupDelay = time.Duration(nbtconv.Int64(m, "PickupDelay")) * (time.Second / 20)
	return n
}

func (ItemType) EncodeNBT(e world.Entity) map[string]any {
	it := e.(*Ent)
	b := it.Behaviour().(*ItemBehaviour)
	return map[string]any{
		"Health":      int16(5),
		"Age":         int16(it.Age() / (time.Second * 20)),
		"PickupDelay": int64(b.pickupDelay / (time.Second * 20)),
		"Pos":         nbtconv.Vec3ToFloat32Slice(it.Position()),
		"Motion":      nbtconv.Vec3ToFloat32Slice(it.Velocity()),
		"Item":        nbtconv.WriteItem(b.Item(), true),
	}
}

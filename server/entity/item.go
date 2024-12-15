package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// NewItem creates a new item entity using the item stack passed. The item
// entity will be positioned at the position passed. If the stack's count
// exceeds its max count, the count of the stack will be changed to the
// maximum.
func NewItem(opts world.EntitySpawnOpts, i item.Stack) *world.EntityHandle {
	conf := itemConf
	conf.Item = i
	return opts.New(ItemType, conf)
}

// NewItemPickupDelay creates a new item entity containing item stack i. A
// delay may be specified which defines for how long the item stack cannot be
// picked up from the ground.
func NewItemPickupDelay(opts world.EntitySpawnOpts, i item.Stack, delay time.Duration) *world.EntityHandle {
	conf := itemConf
	conf.Item = i
	conf.PickupDelay = delay
	return opts.New(ItemType, conf)
}

var itemConf = ItemBehaviourConfig{
	Gravity: 0.04,
	Drag:    0.02,
}

// ItemType is a world.EntityType implementation for Item.
var ItemType itemType

type itemType struct{}

func (t itemType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (itemType) EncodeEntity() string   { return "minecraft:item" }
func (itemType) NetworkOffset() float64 { return 0.125 }
func (itemType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (itemType) DecodeNBT(m map[string]any, data *world.EntityData) {
	conf := itemConf
	conf.Item = nbtconv.MapItem(m, "Item")
	conf.PickupDelay = time.Duration(nbtconv.Int64(m, "PickupDelay")) * (time.Second / 20)

	data.Data = conf.New()
}

func (itemType) EncodeNBT(data *world.EntityData) map[string]any {
	b := data.Data.(*ItemBehaviour)
	return map[string]any{
		"Health":      int16(5),
		"PickupDelay": int64(b.pickupDelay / (time.Second * 20)),
		"Item":        nbtconv.WriteItem(b.Item(), true),
	}
}

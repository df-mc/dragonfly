package item

import (
	"fmt"
	"git.jetbrains.space/dragonfly/dragonfly/dragonfly/block"
	"git.jetbrains.space/dragonfly/dragonfly/dragonfly/block/material"
)

func init() {
	Register(0, 0, block.Air{})
	Register(1, 0, block.Stone{})
	Register(1, 1, block.Granite{})
	Register(1, 2, block.Granite{Polished: true})
	Register(1, 3, block.Diorite{})
	Register(1, 4, block.Diorite{Polished: true})
	Register(1, 5, block.Andesite{})
	Register(1, 6, block.Andesite{Polished: true})
	Register(2, 0, block.Grass{})
	Register(3, 0, block.Dirt{})
	Register(3, 1, block.Dirt{Coarse: true})
	Register(7, 0, block.Bedrock{})
	Register(17, 0, block.Log{Wood: material.OakWood()})
	Register(17, 1, block.Log{Wood: material.SpruceWood()})
	Register(17, 2, block.Log{Wood: material.BirchWood()})
	Register(17, 3, block.Log{Wood: material.JungleWood()})
	Register(162, 0, block.Log{Wood: material.AcaciaWood()})
	Register(162, 1, block.Log{Wood: material.DarkOakWood()})
	Register(-5, 0, block.Log{Wood: material.SpruceWood(), Stripped: true})
	Register(-6, 0, block.Log{Wood: material.BirchWood(), Stripped: true})
	Register(-7, 0, block.Log{Wood: material.JungleWood(), Stripped: true})
	Register(-8, 0, block.Log{Wood: material.AcaciaWood(), Stripped: true})
	Register(-9, 0, block.Log{Wood: material.DarkOakWood(), Stripped: true})
	Register(-10, 0, block.Log{Wood: material.OakWood(), Stripped: true})
}

var items = map[int32]Item{}
var ids = map[Item]int32{}

// Register registers an item with the ID and meta passed. Once registered, items may be obtained from an ID
// and metadata value using item.ByID().
// If an item with the ID and meta passed already exists, Register panics.
func Register(id int32, meta int16, item Item) {
	k := (id << 4) | int32(meta)
	if _, ok := items[k]; ok {
		panic(fmt.Sprintf("item registered with ID %v and meta %v already exists", id, meta))
	}
	items[k] = item
	ids[item] = k
}

// ByID attempts to return an item by the ID and meta it was registered with. If found, the item found is
// returned and the bool true.
func ByID(id int32, meta int16) (Item, bool) {
	it, ok := items[(id<<4)|int32(meta)]
	return it, ok
}

// ToID encodes an item to its numerical ID and metadata value.
func ToID(it Item) (id int32, meta int16) {
	v, ok := ids[it]
	if !ok {
		return 0, 0
	}
	return v >> 4, int16(v & 0x0f)
}

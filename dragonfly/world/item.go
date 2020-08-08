package world

import (
	"fmt"
)

// Item represents an item that may be added to an inventory. It has a method to encode the item to an ID and
// a metadata value.
type Item interface {
	// EncodeItem encodes the item to its Minecraft representation, which consists of a numerical ID and a
	// metadata value.
	EncodeItem() (id int32, meta int16)
}

// NBTer represents either an item or a block which may decode NBT data and encode to NBT data. Typically
// this is done to store additional data.
type NBTer interface {
	// DecodeNBT returns the item or block, depending on which of those the NBTer was, with the NBT data
	// decoded into it.
	DecodeNBT(data map[string]interface{}) interface{}
	EncodeNBT() map[string]interface{}
}

// TickerBlock is an implementation of NBTer with an additional Tick method that is called on every world
// tick for loaded blocks that implement this interface.
type TickerBlock interface {
	NBTer
	Tick(currentTick int64, pos BlockPos, w *World)
}

// RegisterItem registers an item with the ID and meta passed. Once registered, items may be obtained from an
// ID and metadata value using itemByID().
// If an item with the ID and meta passed already exists, RegisterItem panics.
func RegisterItem(name string, item Item) {
	id, meta := item.EncodeItem()
	k := (id << 16) | int32(meta)
	if _, ok := items[k]; ok {
		panic(fmt.Sprintf("item registered with ID %v and meta %v already exists", id, meta))
	}
	items[k] = item
	itemsNames[name] = id
	names[id] = name
}

var items = map[int32]Item{}
var itemsNames = map[string]int32{}
var names = map[int32]string{}

// itemByID attempts to return an item by the ID and meta it was registered with. If found, the item found is
// returned and the bool true.
//lint:ignore U1000 Function is used using compiler directives.
func itemByID(id int32, meta int16) (Item, bool) {
	it, ok := items[(id<<16)|int32(meta)]
	if !ok {
		// Also try obtaining the item with a metadata value of 0, for cases with durability.
		it, ok = items[(id<<16)|int32(0)]
	}
	return it, ok
}

// itemByName attempts to return an item by a name and a metadata value, rather than an ID.
//lint:ignore U1000 Function is used using compiler directives.
//noinspection GoUnusedFunction
func itemByName(name string, meta int16) (Item, bool) {
	id, ok := itemsNames[name]
	if !ok {
		return nil, false
	}
	return itemByID(id, meta)
}

// itemToName encodes an item to its string ID and metadata value.
//lint:ignore U1000 Function is used using compiler directives.
//noinspection GoUnusedFunction
func itemToName(it Item) (name string, meta int16) {
	id, meta := it.EncodeItem()
	return names[id], meta
}

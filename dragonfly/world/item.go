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
	DecodeNBT(data map[string]interface{}) Block
	EncodeNBT() map[string]interface{}
}

// RegisterItem registers an item with the ID and meta passed. Once registered, items may be obtained from an
// ID and metadata value using itemByID().
// If an item with the ID and meta passed already exists, RegisterItem panics.
func RegisterItem(name string, item Item) {
	id, meta := item.EncodeItem()
	k := (id << 4) | int32(meta)
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
func itemByID(id int32, meta int16) (Item, bool) {
	it, ok := items[(id<<4)|int32(meta)]
	return it, ok
}

// itemByName attempts to return an item by a name and a metadata value, rather than an iD.
func itemByName(name string, meta int16) (Item, bool) {
	id, ok := itemsNames[name]
	if !ok {
		return nil, false
	}
	return itemByID(id, meta)
}

// itemToName encodes an item to its string ID and metadata value.
func itemToName(it Item) (name string, meta int16) {
	id, meta := it.EncodeItem()
	return names[id], meta
}

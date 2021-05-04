package world

import (
	_ "embed"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// Item represents an item that may be added to an inventory. It has a method to encode the item to an ID and
// a metadata value.
type Item interface {
	// EncodeItem encodes the item to its Minecraft representation, which consists of a numerical ID and a
	// metadata value.
	EncodeItem() (name string, meta int16)
}

// RegisterItem registers an item with the ID and meta passed. Once registered, items may be obtained from an
// ID and metadata value using itemByID().
// If an item with the ID and meta passed already exists, RegisterItem panics.
func RegisterItem(item Item) {
	name, meta := item.EncodeItem()
	h := itemHash{name: name, meta: meta}

	if _, ok := items[h]; ok {
		panic(fmt.Sprintf("item registered with name %v and meta %v already exists", name, meta))
	}
	if _, ok := itemNamesToRuntimeIDs[name]; !ok {
		panic(fmt.Sprintf("item name %v does not have a runtime ID", name))
	}
	items[h] = item
}

// itemHash is a combination of an item's name and metadata. It is used as a key in hash maps.
type itemHash struct {
	name string
	meta int16
}

var (
	//go:embed item_runtime_ids.nbt
	itemRuntimeIDData []byte
	// items holds a list of all registered items, indexed using the itemHash created when calling
	// Item.EncodeItem.
	items = map[itemHash]Item{}
	// itemRuntimeIDsToNames holds a map to translate item runtime IDs to string IDs.
	itemRuntimeIDsToNames = map[int32]string{}
	// itemNamesToRuntimeIDs holds a map to translate item string IDs to runtime IDs.
	itemNamesToRuntimeIDs = map[string]int32{}
)

// init reads all item entries from the resource JSON, and sets the according values in the runtime ID maps.
func init() {
	var m map[string]int32
	err := nbt.Unmarshal(itemRuntimeIDData, &m)
	if err != nil {
		panic(err)
	}
	for name, rid := range m {
		itemNamesToRuntimeIDs[name] = rid
		itemRuntimeIDsToNames[rid] = name
	}
}

// ItemByName attempts to return an item by a name and a metadata value.
func ItemByName(name string, meta int16) (Item, bool) {
	it, ok := items[itemHash{name: name, meta: meta}]
	if !ok {
		// Also try obtaining the item with a metadata value of 0, for cases with durability.
		it, ok = items[itemHash{name: name}]
	}
	return it, ok
}

// ItemRuntimeID attempts to return the runtime ID of the Item passed. False is returned if the Item is not
// registered.
func ItemRuntimeID(i Item) (rid int32, meta int16, ok bool) {
	name, meta := i.EncodeItem()
	rid, ok = itemNamesToRuntimeIDs[name]
	return rid, meta, ok
}

// ItemByRuntimeID attempts to return an Item by the runtime ID passed. If no item with that runtime ID exists,
// false is returned. ItemByRuntimeID also tries to find the item with a metadata value of 0.
func ItemByRuntimeID(rid int32, meta int16) (Item, bool) {
	name, ok := itemRuntimeIDsToNames[rid]
	if !ok {
		return nil, false
	}
	return ItemByName(name, meta)
}

// Items returns a slice of all registered items.
func Items() []Item {
	m := make([]Item, 0, len(items))
	for _, i := range items {
		m = append(m, i)
	}
	return m
}

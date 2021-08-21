package recipes

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Item is a recipe item. It is an item stack inherently, but it also has an extra value for if it applies to all types.
type Item struct {
	item.Stack
	// AllTypes is true if if the item applies to all of it's type. This is so it can be rendered in the recipe book
	// for alternative types.
	AllTypes bool
}

// ItemType is an item that has a name and a metadata value.
type ItemType struct {
	// Name is the name of the item type.
	Name string `nbt:"name"`
	// MetadataValue is the meta of the item type.
	MetadataValue int32 `nbt:"meta"`
}

// ToItem converts the item type to an item.
func (i ItemType) ToItem() (it world.Item, ok bool) {
	return world.ItemByName(i.Name, int16(i.MetadataValue))
}

// InputItem is an item that is inputted to a crafting menu.
type InputItem struct {
	// Name is the name of the item being inputted.
	Name string `nbt:"name"`
	// MetadataValue is the meta of the item. This can change the item almost completely, or act as durability.
	MetadataValue int32 `nbt:"meta"`
	// Count is the amount of the item.
	Count int32 `nbt:"count"`
}

// ToStack converts an input item to a stack.
func (i InputItem) ToStack() (Item, bool) {
	if len(i.Name) == 0 {
		return Item{}, true
	}
	it, ok := world.ItemByName(i.Name, int16(i.MetadataValue))
	if !ok {
		return Item{}, false
	}

	return Item{Stack: item.NewStack(it, int(i.Count)), AllTypes: i.MetadataValue == 32767}, true
}

// InputItems is an array of input items.
type InputItems []InputItem

// ToStacks converts InputItems into item stacks.
func (i InputItems) ToStacks() (s []Item, ok bool) {
	for _, it := range i {
		st, ok := it.ToStack()
		if !ok {
			return nil, false
		}
		s = append(s, st)
	}
	return s, true
}

// OutputItem is an item that is outputted after crafting.
type OutputItem struct {
	// Name is the name of the item being output.
	Name string `nbt:"name"`
	// MetadataValue is the meta of the item. This can change the item almost completely, or act as durability.
	MetadataValue int32 `nbt:"meta"`
	// Count is the amount of the item.
	Count int16 `nbt:"count"`
	// State is included if the output is a block. If it's not included, the meta can be discarded and the output item can be incorrect.
	State struct {
		Name       string                 `nbt:"name"`
		Properties map[string]interface{} `nbt:"states"`
		Version    int32                  `nbt:"version"`
	} `nbt:"block"`
	// NBTData contains extra NBTData which may modify the item in other, more discreet ways.
	NBTData map[string]interface{} `nbt:"data"`
}

// ToStack converts an input item to a stack.
func (o OutputItem) ToStack() (item.Stack, bool) {
	if len(o.Name) == 0 {
		return item.Stack{}, true
	}

	var it world.Item
	var ok bool
	if o.State.Version != 0 {
		// Item with a block, try parsing the block, then try asserting that to an item. Blocks no longer
		// have their metadata sent, but we still need to get that metadata in order to be able to register
		// different block states as different items.
		if b, ok := world.BlockByName(o.State.Name, o.State.Properties); ok {
			if it, ok = b.(world.Item); !ok {
				return item.Stack{}, false
			}
		}
	} else {
		it, ok = world.ItemByName(o.Name, int16(o.MetadataValue))
		if !ok {
			return item.Stack{}, false
		}
	}
	s := item.NewStack(it, int(o.Count))
	for k, v := range o.NBTData {
		s = s.WithValue(k, v)
	}

	return s, true
}

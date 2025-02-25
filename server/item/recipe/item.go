package recipe

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"math"
)

// Item represents an item that can be used as either the input or output of an item. These do not
// necessarily resolve to an actual item, but can be just as simple as a tag etc.
type Item interface {
	// Count returns the amount of items that is present on the stack. The count is guaranteed never to be
	// negative.
	Count() int
	// Empty checks if the stack is empty (has a count of 0).
	Empty() bool
}

// inputItem is a type representing an input item, with a helper function to convert it to an Item.
type inputItem struct {
	// Name is the name of the item being inputted.
	Name string `nbt:"name"`
	// Meta is the meta of the item. This can change the item almost completely, or act as durability.
	Meta int32 `nbt:"meta"`
	// Count is the amount of the item.
	Count int32 `nbt:"count"`
	// State is included if the output is a block. If it's not included, the meta can be discarded and the output item can be incorrect.
	State struct {
		Name       string                 `nbt:"name"`
		Properties map[string]interface{} `nbt:"states"`
		Version    int32                  `nbt:"version"`
	} `nbt:"block"`
	// Tag is included if the input item is defined by a tag instead of a specific item.
	Tag string `nbt:"tag"`
}

// Item converts an input item to a recipe item.
func (i inputItem) Item() (Item, bool) {
	if i.Tag != "" {
		return NewItemTag(i.Tag, int(i.Count)), true
	}

	it, ok := world.ItemByName(i.Name, int16(i.Meta))
	if !ok {
		return nil, false
	}
	st := item.NewStack(it, int(i.Count))
	if i.Meta == math.MaxInt16 {
		st = st.WithValue("variants", true)
	}

	return st, true
}

// inputItems is a type representing a list of input items, with a helper function to convert it to an Item.
type inputItems []inputItem

// Items converts input items to recipe items.
func (d inputItems) Items() ([]Item, bool) {
	s := make([]Item, 0, len(d))
	for _, i := range d {
		itemInput, ok := i.Item()
		if !ok {
			return nil, false
		}
		s = append(s, itemInput)
	}
	return s, true
}

// outputItem is an output item.
type outputItem struct {
	// Name is the name of the item being output.
	Name string `nbt:"name"`
	// Meta is the meta of the item. This can change the item almost completely, or act as durability.
	Meta int32 `nbt:"meta"`
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

// Stack converts an output item to an item stack.
func (o outputItem) Stack() (item.Stack, bool) {
	it, ok := world.ItemByName(o.Name, int16(o.Meta))
	if !ok {
		return item.Stack{}, false
	}
	if n, ok := it.(world.NBTer); ok {
		it = n.DecodeNBT(o.NBTData).(world.Item)
	}

	return item.NewStack(it, int(o.Count)), true
}

// outputItems is an array of output items.
type outputItems []outputItem

// Stacks converts output items to item stacks.
func (d outputItems) Stacks() ([]item.Stack, bool) {
	s := make([]item.Stack, 0, len(d))
	for _, o := range d {
		itemOutput, ok := o.Stack()
		if !ok {
			return nil, false
		}
		s = append(s, itemOutput)
	}
	return s, true
}

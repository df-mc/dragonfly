package recipe

import (
	"math"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Item represents an item that can be used as either the input or output of a recipe. It does not
// necessarily resolve to an actual item, but can be just as simple as a tag etc.
type Item interface {
	// Count returns the amount of items that is present on the stack. The count is guaranteed never to be
	// negative.
	Count() int
	// Empty checks if the stack is empty (has a count of 0).
	Empty() bool
}

// blockState is the exact encoded state of a block, included with input and output items that are blocks.
type blockState struct {
	Name       string         `nbt:"name"`
	Properties map[string]any `nbt:"states"`
	Version    int32          `nbt:"version"`
}

// inputItem is an input item as present in the recipe data, convertible to an [Item].
type inputItem struct {
	// Name is the name of the item being inputted.
	Name string `nbt:"name"`
	// Meta may change the item almost completely, or act as durability. A value of math.MaxInt16 means any
	// meta matches.
	Meta int32 `nbt:"meta"`
	// Count is the amount of the item.
	Count int32 `nbt:"count"`
	// State is included if the input is a block.
	State blockState `nbt:"block"`
	// Tag is set if the input is defined by an item tag instead of a specific item.
	Tag string `nbt:"tag"`
}

// Item converts an input item to a recipe [Item].
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

// inputItems is a list of input items, where each is convertible to an [Item].
type inputItems []inputItem

// Items converts each input item to an [Item].
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

// outputItem is an output item as present in the recipe data, convertible to an [item.Stack].
type outputItem struct {
	// Name is the name of the item being output.
	Name string `nbt:"name"`
	// Meta may change the item almost completely, or act as durability.
	Meta int32 `nbt:"meta"`
	// Count is the amount of the item.
	Count int16 `nbt:"count"`
	// State holds the exact block state of the output if it is a block.
	State blockState `nbt:"block"`
	// NBTData contains extra NBT which may modify the item in other, more discreet ways.
	NBTData map[string]any `nbt:"data"`
}

// Stack converts an output item to an [item.Stack].
func (o outputItem) Stack() (item.Stack, bool) {
	it, ok := o.item()
	if !ok {
		return item.Stack{}, false
	}
	if n, ok := it.(world.NBTer); ok && len(o.NBTData) > 0 {
		it = n.DecodeNBT(o.NBTData).(world.Item)
	}

	return item.NewStack(it, int(o.Count)), true
}

// item resolves the [world.Item] of an output item.
func (o outputItem) item() (world.Item, bool) {
	if o.State.Name != "" {
		if b, ok := world.BlockByName(o.State.Name, o.State.Properties); ok {
			if it, ok := b.(world.Item); ok {
				return it, true
			}
		}
	}
	return world.ItemByName(o.Name, int16(o.Meta))
}

// outputItems is a list of output items, where each is convertible to an [item.Stack].
type outputItems []outputItem

// Stacks converts each output item to an [item.Stack].
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

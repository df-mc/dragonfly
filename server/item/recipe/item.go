package recipe

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"math"
)

// inputItem contains data for an item that is inputted to a crafting menu.
type inputItem struct {
	// Name is the name of the item being inputted.
	Name string `nbt:"name"`
	// Meta is the meta of the item. This can change the item almost completely, or act as durability.
	Meta int32 `nbt:"meta"`
	// Count is the amount of the item.
	Count int32 `nbt:"count"`
}

// inputItems is a type representing a list of input items, with helper functions to convert them to
type inputItems []inputItem

// Stacks converts input items to item stacks.
func (d inputItems) Stacks() ([]item.Stack, bool) {
	s := make([]item.Stack, 0, len(d))
	for _, i := range d {
		if len(i.Name) == 0 {
			s = append(s, item.Stack{})
			continue
		}
		it, ok := world.ItemByName(i.Name, int16(i.Meta))
		if !ok {
			return nil, false
		}
		st := item.NewStack(it, int(i.Count))
		if i.Meta == math.MaxInt16 {
			st = st.WithVariants()
		}
		s = append(s, st)
	}
	return s, true
}

// outputItem is an item that is outputted after crafting.
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

// outputItems is an array of output items.
type outputItems []outputItem

// Stacks converts output items to item stacks.
func (d outputItems) Stacks() ([]item.Stack, bool) {
	s := make([]item.Stack, 0, len(d))
	for _, o := range d {
		it, ok := world.ItemByName(o.Name, int16(o.Meta))
		if !ok {
			return nil, false
		}
		if b, ok := world.BlockByName(o.State.Name, o.State.Properties); ok {
			if it, ok = b.(world.Item); !ok {
				return nil, false
			}
		}
		s = append(s, item.NewStack(it, int(o.Count)))
	}
	return s, true
}

package item_internal

import (
	"github.com/df-mc/dragonfly/server/world"
)

// Air holds an air block.
var Air world.Block

// IsCarvedPumpkin is a function set to check if an item is a wearable carved pumpkin
var IsCarvedPumpkin func(i world.Item) bool

// Fire holds a fire block.
var Fire world.Block

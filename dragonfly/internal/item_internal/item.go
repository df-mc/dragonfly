package item_internal

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// Air holds an air block.
var Air world.Block

// Water and Lava hold blocks for their respective liquids.
var Water, Lava world.Liquid

// IsCarvedPumpkin is a function set to check if an item is a wearable carved pumpkin
var IsCarvedPumpkin func(i world.Item) bool

// IsWater is a function used to check if a liquid is water.
var IsWater func(b world.Block) bool

// Fire holds a fire block.
var Fire world.Block

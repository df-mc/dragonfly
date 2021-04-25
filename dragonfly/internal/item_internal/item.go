package item_internal

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
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

// IsWaterSource is a function used to check if a block is a water source.
var IsWaterSource func(b world.Block) bool

// Replaceable is a function used to check if a block is replaceable.
var Replaceable func(w *world.World, pos cube.Pos, with world.Block) bool

// Fire holds a fire block.
var Fire world.Block

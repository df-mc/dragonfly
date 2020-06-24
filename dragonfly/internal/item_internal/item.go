package item_internal

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// Air holds an air block.
var Air world.Block

// GrassPath holds a grass path block.
var GrassPath world.Block

// Grass holds a grass block.
var Grass world.Block

// Water and Lava hold blocks for their respective liquids.
var Water, Lava world.Liquid

// IsUnstrippedLog is a function set to check if a block is a log.
var IsUnstrippedLog func(b world.Block) bool

// StripLog is a function used to convert a log block to a stripped log block.
var StripLog func(b world.Block) world.Block

// IsWater is a function used to check if a liquid is water.
var IsWater func(b world.Liquid) bool

// Replaceable is a function used to check if a block is replaceable.
var Replaceable func(w *world.World, pos world.BlockPos, with world.Block) bool

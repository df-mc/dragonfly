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

// Dirt holds a dirt block.
var Dirt world.Block

// Farmland holds a farmland block.
var Farmland world.Block

// Water and Lava hold blocks for their respective liquids.
var Water, Lava world.Liquid

// IsUnstrippedLog is a function set to check if a block is a log.
var IsUnstrippedLog func(b world.Block) bool

// StripLog is a function used to convert a log block to a stripped log block.
var StripLog func(b world.Block) world.Block

// IsCarvedPumpkin is a function set to check if an item is a wearable carved pumpkin
var IsCarvedPumpkin func(i world.Item) bool

// IsUncarvedPumpkin is a function set to check if a block is an uncarved pumpkin.
var IsUncarvedPumpkin func(b world.Block) bool

// CarvePumpkin is a function used to convert a pumpkin block to a carved pumpkin block.
var CarvePumpkin func(b world.Block, face world.Face) world.Block

// IsWater is a function used to check if a liquid is water.
var IsWater func(b world.Block) bool

// IsWaterSource is a function used to check if a block is a water source.
var IsWaterSource func(b world.Block) bool

// Bonemeal is a function used to attempt to use it on a block.
var Bonemeal func(pos world.BlockPos, w *world.World) bool

// Replaceable is a function used to check if a block is replaceable.
var Replaceable func(w *world.World, pos world.BlockPos, with world.Block) bool

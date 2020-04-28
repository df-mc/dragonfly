package item_internal

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
)

// GrassPath holds a grass path block.
var GrassPath world.Block

// Grass holds a grass block.
var Grass world.Block

// IsUnstrippedLog is a function set to check if a block is a log.
var IsUnstrippedLog func(b world.Block) bool

// StripLog is a function used to convert a log block to a stripped log block.
var StripLog func(b world.Block) world.Block

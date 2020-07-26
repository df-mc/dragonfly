package sound

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// BlockPlace is a sound sent when a block is placed.
type BlockPlace struct {
	// Block is the block which is placed, for which a sound should be played. The sound played depends on
	// the block type.
	Block world.Block

	sound
}

// BlockBreaking is a sound sent continuously while a player is breaking a block.
type BlockBreaking struct {
	// Block is the block which is being broken, for which a sound should be played. The sound played depends
	// on the block type.
	Block world.Block

	sound
}

// Fizz is a sound sent when a lava block and a water block interact with each other in a way that one of
// them turns into a solid block.
type Fizz struct{ sound }

// ChestOpen is played when a chest is opened.
type ChestOpen struct{ sound }

// ChestClose is played when a chest is closed.
type ChestClose struct{ sound }

// Deny is a sound played when a block is placed or broken above a 'Deny' block from Education edition.
type Deny struct{ sound }

// Door is a sound played when a (trap)door is opened or closed.
type Door struct{ sound }

// sound implements the world.Sound interface.
type sound struct{}

// Play ...
func (sound) Play(*world.World, mgl64.Vec3) {}

package sound

import (
	"github.com/df-mc/dragonfly/server/block/instrument"
	"github.com/df-mc/dragonfly/server/world"
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

// BarrelOpen is played when a barrel is opened.
type BarrelOpen struct{ sound }

// BarrelClose is played when a barrel is closed.
type BarrelClose struct{ sound }

// Deny is a sound played when a block is placed or broken above a 'Deny' block from Education edition.
type Deny struct{ sound }

// Door is a sound played when a (trap)door is opened or closed.
type Door struct{ sound }

// DoorCrash is a sound played when a door is forced open.
type DoorCrash struct{ sound }

// Click is a clicking sound.
type Click struct{ sound }

// Ignite is a sound played when using a flint & steel.
type Ignite struct{ sound }

// FireExtinguish is a sound played when a fire is extinguished.
type FireExtinguish struct{ sound }

// Note is a sound played by note blocks.
type Note struct {
	sound
	// Instrument is the instrument of the note block.
	Instrument instrument.Instrument
	// Pitch is the pitch of the note.
	Pitch int
}

// sound implements the world.Sound interface.
type sound struct{}

// Play ...
func (sound) Play(*world.World, mgl64.Vec3) {}

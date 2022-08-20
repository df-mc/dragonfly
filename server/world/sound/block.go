package sound

import (
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

// GlassBreak is a sound played when a glass block or item is broken.
type GlassBreak struct{ sound }

// Fizz is a sound sent when a lava block and a water block interact with each other in a way that one of
// them turns into a solid block.
type Fizz struct{ sound }

// AnvilLand is played when an anvil lands on the ground.
type AnvilLand struct{ sound }

// AnvilUse is played when an anvil is used.
type AnvilUse struct{ sound }

// AnvilBreak is played when an anvil is broken.
type AnvilBreak struct{ sound }

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

// TNT is a sound played when TNT is ignited.
type TNT struct{ sound }

// FireExtinguish is a sound played when a fire is extinguished.
type FireExtinguish struct{ sound }

// Note is a sound played by note blocks.
type Note struct {
	sound
	// Instrument is the instrument of the note block.
	Instrument Instrument
	// Pitch is the pitch of the note.
	Pitch int
}

// MusicDiscPlay is a sound played when a music disc has started playing in a jukebox.
type MusicDiscPlay struct {
	sound

	// DiscType is the disc type of the music disc.
	DiscType DiscType
}

// MusicDiscEnd is a sound played when a music disc has stopped playing in a jukebox.
type MusicDiscEnd struct{ sound }

// ItemFrameAdd is a sound played when an item is added to an item frame.
type ItemFrameAdd struct{ sound }

// ItemFrameRemove is a sound played when an item is removed from an item frame.
type ItemFrameRemove struct{ sound }

// ItemFrameRotate is a sound played when an item frame's item is rotated.
type ItemFrameRotate struct{ sound }

// FurnaceCrackle is a sound played every one to five seconds from a furnace.
type FurnaceCrackle struct{ sound }

// BlastFurnaceCrackle is a sound played every one to five seconds from a blast furnace.
type BlastFurnaceCrackle struct{ sound }

// SmokerCrackle is a sound played every one to five seconds from a smoker.
type SmokerCrackle struct{ sound }

// ComposterEmpty is a sound played when a composter has been emptied.
type ComposterEmpty struct{ sound }

// ComposterFill is a sound played when a composter has been filled, but not gone up a layer.
type ComposterFill struct{ sound }

// ComposterFillLayer is a sound played when a composter has been filled and gone up a layer.
type ComposterFillLayer struct{ sound }

// ComposterReady is a sound played when a composter has produced bone meal and is ready to be collected.
type ComposterReady struct{ sound }

// sound implements the world.Sound interface.
type sound struct{}

// Play ...
func (sound) Play(*world.World, mgl64.Vec3) {}

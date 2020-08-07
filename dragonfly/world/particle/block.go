package particle

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// BlockBreak is a particle sent when a block is broken. It represents a bunch of particles that are textured
// like the block that the particle holds.
type BlockBreak struct {
	// Block is the block of which particles should be shown. The particles will change depending on what
	// block is held.
	Block world.Block
}

// PunchBlock is a particle shown when a player is punching a block. It shows particles of a specific block
// type at a particular face of a block.
type PunchBlock struct {
	// Block is the block of which particles should be shown. The particles will change depending on what
	// block is punched.
	Block world.Block
	// Face is the face of the block that was punched. It is here that the particles will be shown.
	Face world.Face
}

// BlockForceField is a particle that shows up as a block that turns invisible from an opaque black colour.
type BlockForceField struct{}

// Bonemeal is a particle that shows up on bonemeal usage.
type Bonemeal struct{}

// Spawn ...
func (PunchBlock) Spawn(*world.World, mgl64.Vec3) {}

// Spawn ...
func (BlockBreak) Spawn(*world.World, mgl64.Vec3) {}

// Spawn ...
func (BlockForceField) Spawn(*world.World, mgl64.Vec3) {}

// Spawn ...
func (Bonemeal) Spawn(*world.World, mgl64.Vec3) {}

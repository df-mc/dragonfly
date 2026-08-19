package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Connections holds the sides that a block such as a fence connects to.
type Connections struct {
	north, east, south, west bool
}

// Uint8 returns the Connections as a uint8.
func (c Connections) Uint8() uint8 {
	return boolByte(c.north) | boolByte(c.east)<<1 | boolByte(c.south)<<2 | boolByte(c.west)<<3
}

// properties returns the block state properties of the Connections.
func (c Connections) properties() map[string]any {
	return map[string]any{
		"minecraft:connection_north": c.north,
		"minecraft:connection_east":  c.east,
		"minecraft:connection_south": c.south,
		"minecraft:connection_west":  c.west,
	}
}

// connector is a block or block model that reports the sides it connects to.
type connector interface {
	// Connects returns true if the block connects to the block at the face passed.
	Connects(pos cube.Pos, face cube.Face, src world.BlockSource) bool
}

// calculateConnections returns the Connections of the connector at a position.
func calculateConnections(c connector, tx *world.Tx, pos cube.Pos) Connections {
	return Connections{
		north: c.Connects(pos, cube.FaceNorth, tx),
		east:  c.Connects(pos, cube.FaceEast, tx),
		south: c.Connects(pos, cube.FaceSouth, tx),
		west:  c.Connects(pos, cube.FaceWest, tx),
	}
}

// allConnections returns a list of all combinations of Connections.
func allConnections() (all []Connections) {
	for i := range 16 {
		all = append(all, Connections{north: i&0x1 != 0, east: i&0x2 != 0, south: i&0x4 != 0, west: i&0x8 != 0})
	}
	return
}

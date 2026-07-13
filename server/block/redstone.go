package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// redstoneAttachmentSupported reports whether a redstone component may attach
// to the block behind the face passed.
func redstoneAttachmentSupported(tx *world.Tx, pos cube.Pos, face cube.Face) bool {
	support := pos.Side(face.Opposite())
	if support.OutOfBounds(tx.Range()) {
		return false
	}
	return tx.Block(support).Model().FaceSolid(support, face, tx)
}

// redstoneFloorComponentSupported reports whether a floor-mounted redstone
// component is supported by the block below it.
func redstoneFloorComponentSupported(tx *world.Tx, pos cube.Pos) bool {
	support := pos.Side(cube.FaceDown)
	if support.OutOfBounds(tx.Range()) {
		return false
	}
	return tx.Block(support).Model().FaceSolid(support, cube.FaceUp, tx)
}

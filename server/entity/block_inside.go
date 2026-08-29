package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// HandleEntityInsideBlocks applies the inside effects of every block and liquid intersecting e. Custom entity
// implementations may call this function from their tick method to use the same block-effect pipeline as Ent and
// players.
func HandleEntityInsideBlocks(e world.Entity, tx *world.Tx) {
	box := e.H().Type().BBox(e).Translate(e.Position()).Grow(-0.0001)
	low, high := cube.PosFromVec3(box.Min()), cube.PosFromVec3(box.Max())

	for blockPos := range cube.Range3D(low, high) {
		b := tx.Block(blockPos)
		if inside, ok := b.(block.EntityInsider); ok {
			inside.EntityInside(blockPos, tx, e)
			if stopHandlingInsideBlocks(e) {
				return
			}
			if _, liquid := b.(world.Liquid); liquid {
				continue
			}
		}
		if liquid, ok := tx.Liquid(blockPos); ok {
			if inside, ok := liquid.(block.EntityInsider); ok {
				inside.EntityInside(blockPos, tx, e)
				if stopHandlingInsideBlocks(e) {
					return
				}
			}
		}
	}
}

// stopHandlingInsideBlocks reports whether an entity reached a terminal state while handling an inside-block effect.
func stopHandlingInsideBlocks(e world.Entity) bool {
	stopper, ok := e.(interface{ stopHandlingInsideBlocks() bool })
	return ok && stopper.stopHandlingInsideBlocks()
}

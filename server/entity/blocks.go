package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// CheckEntityInsiders checks if the entity is colliding with any EntityInsider blocks.
func CheckEntityInsiders(tx *world.Tx, box cube.BBox, ent world.Entity) {
	low, high := cube.PosFromVec3(box.Min()), cube.PosFromVec3(box.Max())
	for blockPos := range cube.Range3D(low, high) {
		b := tx.Block(blockPos)
		if collide, ok := b.(block.EntityInsider); ok {
			collide.EntityInside(blockPos, tx, ent)
			if _, liquid := b.(world.Liquid); liquid {
				continue
			}
		}

		if l, ok := tx.Liquid(blockPos); ok {
			if collide, ok := l.(block.EntityInsider); ok {
				collide.EntityInside(blockPos, tx, ent)
			}
		}
	}
}

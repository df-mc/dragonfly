package entity

import (
	"math"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// CheckEntityInsiders dispatches the block.EntityInsider hook of every block and liquid that the bounding
// box passed intersects. The box is the entity's bounding box translated to its position.
func CheckEntityInsiders(tx *world.Tx, e world.Entity, box cube.BBox) {
	grown := box.Grow(-0.0001)
	low, high := cube.PosFromVec3(grown.Min()), cube.PosFromVec3(grown.Max())

	for y := low[1]; y <= high[1]; y++ {
		for x := low[0]; x <= high[0]; x++ {
			for z := low[2]; z <= high[2]; z++ {
				blockPos := cube.Pos{x, y, z}
				b := tx.Block(blockPos)
				if collide, ok := b.(block.EntityInsider); ok {
					collide.EntityInside(blockPos, tx, e)
					if _, liquid := b.(world.Liquid); liquid {
						continue
					}
				}
				if l, ok := tx.Liquid(blockPos); ok {
					if collide, ok := l.(block.EntityInsider); ok {
						collide.EntityInside(blockPos, tx, e)
					}
				}
			}
		}
	}
}

// CheckEntitySteppers dispatches the block.EntityStepper hook of the block that the bounding box passed
// stands on. Callers only call it while the entity is on the ground.
func CheckEntitySteppers(tx *world.Tx, e world.Entity, box cube.BBox) {
	low, high := blocksUnderBox(box)
	for x := low[0]; x <= high[0]; x++ {
		for z := low[2]; z <= high[2]; z++ {
			pos := cube.Pos{x, low[1], z}
			if stepper, ok := tx.Block(pos).(block.EntityStepper); ok {
				stepper.EntityStepOn(pos, tx, e)
				return
			}
		}
	}
}

// CheckEntityLanders dispatches the block.EntityLander hook of the block that the bounding box passed
// lands on. The hook may change the fall distance passed.
func CheckEntityLanders(tx *world.Tx, e world.Entity, box cube.BBox, distance float64) float64 {
	low, high := blocksUnderBox(box)
	for x := low[0]; x <= high[0]; x++ {
		for z := low[2]; z <= high[2]; z++ {
			pos := cube.Pos{x, low[1], z}
			if l, ok := tx.Block(pos).(block.EntityLander); ok {
				l.EntityLand(pos, tx, e, &distance)
				return distance
			}
		}
	}
	return distance
}

// blocksUnderBox returns the corners of the range of block positions directly below the bounding box
// passed. Every block in that range is one the entity stands on: the entity may be narrower than a block
// and rest on the edge of one with its centre over the block beside it.
func blocksUnderBox(box cube.BBox) (low, high cube.Pos) {
	y := int(math.Floor(box.Min()[1] - 0.0001))
	horizontal := box.Grow(-0.0001)
	low, high = cube.PosFromVec3(horizontal.Min()), cube.PosFromVec3(horizontal.Max())
	low[1], high[1] = y, y
	return low, high
}

// breathingDistanceBelowEyes is the lowest distance an entity can be in water and still be able to breathe,
// based on the entity's eye height.
const breathingDistanceBelowEyes = 0.11111111

// Submerged checks if the eyes of the entity passed are underwater.
func Submerged(e world.Entity, tx *world.Tx) bool {
	pos := cube.PosFromVec3(EyePosition(e))
	if l, ok := tx.Liquid(pos); ok {
		if _, ok := l.(block.Water); ok {
			d := float64(l.SpreadDecay()) + 1
			if l.LiquidFalling() {
				d = 1
			}
			return e.Position()[1] < (pos.Side(cube.FaceUp).Vec3()[1])-(d/9-breathingDistanceBelowEyes)
		}
	}
	return false
}

// InsideOfSolid checks if the entity passed is suffocating inside a solid block.
func InsideOfSolid(e world.Entity, tx *world.Tx) bool {
	pos := cube.PosFromVec3(EyePosition(e))
	b, box := tx.Block(pos), e.H().Type().BBox(e).Translate(e.Position())

	_, solid := b.Model().(model.Solid)
	if !solid {
		return false
	}
	if d, diffuses := b.(block.LightDiffuser); diffuses && d.LightDiffusionLevel() == 0 {
		return false
	}
	if immune, ok := b.(block.NonSuffocating); ok && immune.PreventsSuffocation() {
		return false
	}
	for _, blockBox := range b.Model().BBox(pos, tx) {
		if blockBox.Translate(pos.Vec3()).IntersectsWith(box) {
			return true
		}
	}
	return false
}

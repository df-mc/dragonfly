package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

const (
	bubbleColumnRisePerTick        = 0.06
	bubbleColumnSurfaceRisePerTick = 0.10
	bubbleColumnSinkPerTick        = -0.30
	bubbleColumnSurfaceSinkPerTick = -0.03
	bubbleColumnRiseSpeedLimit     = 0.70
	bubbleColumnSurfaceRiseLimit   = 1.80
	bubbleColumnSinkSpeedLimit     = -0.30
	bubbleColumnSurfaceSinkLimit   = -0.90
)

// BubbleColumn is a column of bubbles formed in source water above soul sand or magma. Soul sand pushes entities
// upward, whereas magma pulls them downward.
type BubbleColumn struct {
	empty
	replaceable
	sourceWaterDisplacer

	// DragDown specifies if the bubble column was formed by magma and pulls entities downward.
	DragDown bool
}

// EntityInside accelerates entities vertically while they are inside the bubble column.
func (b BubbleColumn) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if fall, ok := e.(fallDistanceEntity); ok {
		fall.ResetFallDistance()
	}
	v, ok := e.(velocityEntity)
	if !ok {
		return
	}

	velocity := v.Velocity()
	surface := isAir(tx.Block(pos.Side(cube.FaceUp)))
	if b.DragDown {
		change, limit := bubbleColumnSinkPerTick, bubbleColumnSinkSpeedLimit
		if surface {
			change, limit = bubbleColumnSurfaceSinkPerTick, bubbleColumnSurfaceSinkLimit
		}
		velocity[1] = max(velocity[1]+change, limit)
	} else {
		change, limit := bubbleColumnRisePerTick, bubbleColumnRiseSpeedLimit
		if surface {
			change, limit = bubbleColumnSurfaceRisePerTick, bubbleColumnSurfaceRiseLimit
		}
		velocity[1] = min(velocity[1]+change, limit)
	}
	v.SetVelocity(velocity)
}

// NeighbourUpdateTick updates this bubble column and all bubble column blocks above it.
func (BubbleColumn) NeighbourUpdateTick(pos, changedNeighbour cube.Pos, tx *world.Tx) {
	if changedNeighbour != pos && changedNeighbour != pos.Side(cube.FaceDown) && changedNeighbour != pos.Side(cube.FaceUp) {
		return
	}
	updateBubbleColumn(pos, tx)
}

// LightDiffusionLevel returns the amount of light lost when travelling through a bubble column.
func (BubbleColumn) LightDiffusionLevel() uint8 {
	return 2
}

// HasLiquidDrops returns false because bubble columns do not drop items when displaced by liquid.
func (BubbleColumn) HasLiquidDrops() bool {
	return false
}

// SideClosed returns false because water may flow through every side of a bubble column.
func (BubbleColumn) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// EncodeBlock ...
func (b BubbleColumn) EncodeBlock() (string, map[string]any) {
	return "minecraft:bubble_column", map[string]any{"drag_down": b.DragDown}
}

// allBubbleColumns returns all possible bubble column states.
func allBubbleColumns() []world.Block {
	return []world.Block{BubbleColumn{}, BubbleColumn{DragDown: true}}
}

// updateBubbleColumn creates, redirects, or removes consecutive bubble column blocks above pos.
func updateBubbleColumn(pos cube.Pos, tx *world.Tx) {
	dragDown, active := bubbleColumnDirection(tx.Block(pos.Side(cube.FaceDown)))
	for !pos.OutOfBounds(tx.Range()) {
		current := tx.Block(pos)
		if column, ok := current.(BubbleColumn); ok {
			liquid, waterPresent := tx.Liquid(pos)
			if !waterPresent || !isSourceWater(liquid) {
				tx.SetBlock(pos, nil, nil)
				active = false
				pos = pos.Side(cube.FaceUp)
				continue
			}
			if !active {
				tx.SetBlock(pos, nil, nil)
				pos = pos.Side(cube.FaceUp)
				continue
			}
			if column.DragDown != dragDown {
				tx.SetBlock(pos, BubbleColumn{DragDown: dragDown}, nil)
			}
			pos = pos.Side(cube.FaceUp)
			continue
		}

		if !active || !isSourceWater(current) {
			return
		}
		tx.SetBlock(pos, BubbleColumn{DragDown: dragDown}, nil)
		pos = pos.Side(cube.FaceUp)
	}
}

func bubbleColumnDirection(b world.Block) (dragDown bool, active bool) {
	switch b := b.(type) {
	case BubbleColumn:
		return b.DragDown, true
	case Magma:
		return true, true
	case SoulSand:
		return false, true
	default:
		return false, false
	}
}

func isSourceWater(b world.Block) bool {
	w, ok := b.(Water)
	return ok && w.Depth == 8 && !w.Falling
}

func isAir(b world.Block) bool {
	_, ok := b.(Air)
	return ok
}

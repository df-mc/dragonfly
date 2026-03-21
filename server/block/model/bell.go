package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Bell is a model used by bells.
type Bell struct {
	// Attachment is the bell attachment type.
	Attachment BellAttachment
	// Facing is the bell's horizontal direction.
	Facing cube.Direction
}

// BellAttachment is the attachment variant used by the bell model.
type BellAttachment uint8

const (
	BellAttachmentStanding BellAttachment = iota
	BellAttachmentHanging
	BellAttachmentSide
	BellAttachmentMultiple
)

// BBox ...
func (b Bell) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	switch b.Attachment {
	case BellAttachmentStanding:
		if b.Facing.Face().Axis() == cube.X {
			return []cube.BBox{cube.Box(0.25, 0, 0, 0.75, 1, 1)}
		}
		return []cube.BBox{cube.Box(0, 0, 0.25, 1, 1, 0.75)}
	case BellAttachmentHanging:
		return []cube.BBox{
			cube.Box(0.3125, 0.375, 0.3125, 0.6875, 0.8125, 0.6875),
			cube.Box(0.25, 0.25, 0.25, 0.75, 0.375, 0.75),
			cube.Box(0.4375, 0.8125, 0.4375, 0.5625, 1, 0.5625),
		}
	case BellAttachmentMultiple:
		boxes := []cube.BBox{
			cube.Box(0.3125, 0.375, 0.3125, 0.6875, 0.8125, 0.6875),
			cube.Box(0.25, 0.25, 0.25, 0.75, 0.375, 0.75),
		}
		if b.Facing.Face().Axis() == cube.X {
			return append(boxes, cube.Box(0.4375, 0.8125, 0, 0.5625, 0.9375, 1))
		}
		return append(boxes, cube.Box(0, 0.8125, 0.4375, 1, 0.9375, 0.5625))
	case BellAttachmentSide:
		boxes := []cube.BBox{
			cube.Box(0.3125, 0.375, 0.3125, 0.6875, 0.8125, 0.6875),
			cube.Box(0.25, 0.25, 0.25, 0.75, 0.375, 0.75),
		}
		switch b.Facing {
		case cube.North:
			return append(boxes, cube.Box(0.4375, 0.8125, 0, 0.5625, 0.9375, 0.8125))
		case cube.South:
			return append(boxes, cube.Box(0.4375, 0.8125, 0.1875, 0.5625, 0.9375, 1))
		case cube.West:
			return append(boxes, cube.Box(0, 0.8125, 0.4375, 0.8125, 0.9375, 0.5625))
		default:
			return append(boxes, cube.Box(0.1875, 0.8125, 0.4375, 1, 0.9375, 0.5625))
		}
	default:
		return []cube.BBox{
			cube.Box(0.3125, 0.375, 0.3125, 0.6875, 0.8125, 0.6875),
			cube.Box(0.25, 0.25, 0.25, 0.75, 0.375, 0.75),
		}
	}
}

// FaceSolid always returns false.
func (Bell) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}

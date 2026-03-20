package block

import "github.com/df-mc/dragonfly/server/block/cube"

// HangingAttachment describes how a hanging sign is attached. It may be mounted to the side of a block, hang
// underneath a block by chains, or be attached underneath a block with the 16-direction ceiling variant.
type HangingAttachment struct {
	ceiling  bool
	attached bool
	facing   cube.Direction
	o        cube.Orientation
}

// WallHangingAttachment returns a HangingAttachment for a hanging sign mounted to the side of a block.
func WallHangingAttachment(facing cube.Direction) HangingAttachment {
	return HangingAttachment{facing: facing}
}

// CeilingHangingAttachment returns a HangingAttachment for a hanging sign hanging underneath a block.
func CeilingHangingAttachment(facing cube.Direction) HangingAttachment {
	return HangingAttachment{ceiling: true, facing: facing}
}

// AttachedCeilingHangingAttachment returns a HangingAttachment for a hanging sign attached underneath a block using
// the 16-direction ceiling variant.
func AttachedCeilingHangingAttachment(o cube.Orientation) HangingAttachment {
	return HangingAttachment{ceiling: true, attached: true, o: o}
}

// Uint8 returns the HangingAttachment as a uint8.
func (a HangingAttachment) Uint8() uint8 {
	if !a.ceiling {
		return uint8(a.facing)
	}
	if !a.attached {
		return 4 | uint8(a.facing)
	}
	return 8 | uint8(a.o)
}

// Rotation returns the rotation of the HangingAttachment.
func (a HangingAttachment) Rotation() cube.Rotation {
	if a.attached {
		return cube.Rotation{a.o.Yaw()}
	}
	return WallAttachment(a.facing).Rotation()
}

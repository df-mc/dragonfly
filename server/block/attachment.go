package block

import "github.com/df-mc/dragonfly/server/block/cube"

// Attachment describes the attachment of a block to another block. It is either of the type WallAttachment, which can
// only have 90 degree facing values, or StandingAttachment, which has more freedom using a cube.Orientation.
type Attachment struct {
	hanging bool
	facing  cube.Direction
	o       cube.Orientation
}

// WallAttachment returns an Attachment to a wall with a facing direction.
func WallAttachment(facing cube.Direction) Attachment {
	return Attachment{hanging: true, facing: facing}
}

// StandingAttachment returns an Attachment to the ground with an orientation.
func StandingAttachment(o cube.Orientation) Attachment {
	return Attachment{o: o}
}

// Uint8 returns the Attachment as a uint8.
func (a Attachment) Uint8() uint8 {
	if !a.hanging {
		return 1 | (uint8(a.o) << 1)
	}
	return uint8(a.facing) << 1
}

// FaceUint8 returns the facing of the Attachment as a uint8.
func (a Attachment) FaceUint8() uint8 {
	return uint8(a.facing)
}

// RotateLeft rotates the Attachment the left way around by 90 degrees.
func (a Attachment) RotateLeft() Attachment {
	return Attachment{hanging: a.hanging, facing: a.facing.RotateLeft(), o: a.o.RotateLeft()}
}

// RotateRight rotates the Attachment the right way around by 90 degrees.
func (a Attachment) RotateRight() Attachment {
	return Attachment{hanging: a.hanging, facing: a.facing.RotateLeft(), o: a.o.RotateLeft()}
}

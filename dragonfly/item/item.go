package item

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/block"
	"github.com/go-gl/mathgl/mgl32"
)

// Item represents an item that may be added to an inventory.
type Item interface {
	// EncodeItem encodes an item to its Minecraft representation - A numerical ID with a numerical meta
	// value.
	EncodeItem() (id int32, meta int16)
}

// UsableOnBlock represents an item that may be used on a block. If an item implements this interface, the
// UseOnBlock method is called whenever the item is used on a block.
type UsableOnBlock interface {
	// UseOnBlock is called when an item is used on a block. The IO passed is the world that the item was used
	// in. The user passed is the entity that used the item. Usually this entity is a player.
	// The position of the block that was clicked, along with the clicked face and the position clicked
	// relative to the corner of the block are passed.
	UseOnBlock(io IO, user User, pos block.Position, clickedFace block.Face, clickPos mgl32.Vec3)
}

// IO represents an IO source that items may be used on to edit the world or to obtain data from the world,
// such as setting a block or adding an entity.
type IO interface {
	block.IO
}

// User represents an entity that is able to use an item in the world, typically entities such as players,
// which interact with the world using an item.
type User interface {
	// Position returns the current position of the entity in the world.
	Position() mgl32.Vec3
	// Yaw returns the yaw of the entity. This is horizontal rotation (rotation around the vertical axis), and
	// is 0 when the entity faces forward.
	Yaw() float32
	// Pitch returns the pitch of the entity. This is vertical rotation (rotation around the horizontal axis),
	// and is 0 when the entity faces forward.
	Pitch() float32
}

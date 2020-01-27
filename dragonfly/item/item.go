package item

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block"
	"github.com/go-gl/mathgl/mgl32"
)

// Item represents an item that may be added to an inventory.
type Item interface{}

// MaxCounter represents an item that has a specific max count. By default, each item will be expected to have
// a maximum count of 64. MaxCounter may be implemented to change this behaviour.
type MaxCounter interface {
	// MaxCount returns the maximum number of items that a stack may be composed of. The number returned must
	// be positive.
	MaxCount() int
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

// Usable represents an item that may be used 'in the air'. If an item implements this interface, the Use
// method is called whenever the item is used while pointing at the air. (For example, when throwing an egg.)
type Usable interface {
	// Use is called when the item is used in the air. The user that used the item is passed to the method.
	Use(user User)
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

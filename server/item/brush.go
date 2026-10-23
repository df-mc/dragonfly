package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Brush is a tool used to excavate suspicious sand and suspicious gravel.
type Brush struct{}

// brusher represents a User that can brush blocks, such as a player.
type brusher interface {
	Brush(pos cube.Pos, face cube.Face)
}

// UseOnBlock brushes the block clicked. The client keeps using the brush on the block while it is held.
func (Brush) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, _ *world.Tx, user User, _ *UseContext) bool {
	if b, ok := user.(brusher); ok {
		b.Brush(pos, face)
	}
	return false
}

// MaxCount ...
func (Brush) MaxCount() int {
	return 1
}

// DurabilityInfo ...
func (Brush) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: 65,
		BrokenItem:    simpleItem(Stack{}),
	}
}

// EncodeItem ...
func (Brush) EncodeItem() (name string, meta int16) {
	return "minecraft:brush", 0
}

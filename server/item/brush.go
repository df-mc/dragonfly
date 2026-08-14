package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Brush is a tool used to excavate suspicious sand and suspicious gravel.
type Brush struct{}

// brusher represents a User that is able to brush blocks over a longer duration, such as a player.
type brusher interface {
	// StartBrushing makes the brusher start brushing the block at the position passed on the face passed.
	StartBrushing(pos cube.Pos, face cube.Face)
}

// UseOnBlock ...
func (Brush) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, _ *world.Tx, user User, _ *UseContext) bool {
	if b, ok := user.(brusher); ok {
		b.StartBrushing(pos, face)
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

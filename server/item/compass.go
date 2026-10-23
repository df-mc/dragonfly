package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Compass is an item used to find the spawn position of a world.
type Compass struct {
	// TrackingHandle is the handle of the lodestone the compass tracks, or 0 to point to the world spawn.
	TrackingHandle int32
}

// UseOnBlock links the compass to the block clicked if it is a lodestone.
func (c Compass) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	l, ok := tx.Block(pos).(compassLinker)
	if !ok {
		return false
	}
	held, _ := user.HeldItems()
	l.LinkCompass(pos, tx, held, ctx)
	return true
}

// compassLinker is a block that a compass may be linked to, such as a lodestone.
type compassLinker interface {
	LinkCompass(pos cube.Pos, tx *world.Tx, held Stack, ctx *UseContext)
}

// EncodeItem ...
func (c Compass) EncodeItem() (name string, meta int16) {
	if c.TrackingHandle != 0 {
		return "minecraft:lodestone_compass", 0
	}
	return "minecraft:compass", 0
}

// Glinted returns true if the compass is linked to a lodestone.
func (c Compass) Glinted() bool { return c.TrackingHandle != 0 }

// EncodeNBT ...
func (c Compass) EncodeNBT() map[string]any {
	if c.TrackingHandle == 0 {
		return nil
	}
	return map[string]any{"trackingHandle": c.TrackingHandle}
}

// DecodeNBT ...
func (c Compass) DecodeNBT(data map[string]any) any {
	c.TrackingHandle = 0
	switch handle := data["trackingHandle"].(type) {
	case int32:
		c.TrackingHandle = handle
	case int:
		c.TrackingHandle = int32(handle)
	}
	return c
}

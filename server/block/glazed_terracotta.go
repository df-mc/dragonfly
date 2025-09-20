package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// GlazedTerracotta is a vibrant solid block that comes in the 16 regular dye colours.
type GlazedTerracotta struct {
	solid
	bassDrum

	// Colour specifies the colour of the block.
	Colour item.Colour
	// Facing specifies the face of the block.
	Facing cube.Direction
}

func (t GlazedTerracotta) BreakInfo() BreakInfo {
	return newBreakInfo(1.4, pickaxeHarvestable, pickaxeEffective, oneOf(t))
}

func (t GlazedTerracotta) EncodeItem() (name string, meta int16) {
	return "minecraft:" + t.Colour.SilverString() + "_glazed_terracotta", 0
}

func (t GlazedTerracotta) EncodeBlock() (name string, properties map[string]any) {
	if t.Facing == unknownDirection {
		return "minecraft:" + t.Colour.SilverString() + "_glazed_terracotta", map[string]any{"facing_direction": int32(0)}
	}
	return "minecraft:" + t.Colour.SilverString() + "_glazed_terracotta", map[string]any{"facing_direction": int32(2 + t.Facing)}
}

// UseOnBlock ensures the proper facing is used when placing a glazed terracotta block, by using the opposite of the player.
func (t GlazedTerracotta) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, t)
	if !used {
		return
	}
	t.Facing = user.Rotation().Direction().Opposite()

	place(tx, pos, t, user, ctx)
	return placed(ctx)
}

// allGlazedTerracotta returns glazed terracotta blocks with all possible colours.
func allGlazedTerracotta() (b []world.Block) {
	for _, dir := range append(cube.Directions(), unknownDirection) {
		for _, c := range item.Colours() {
			b = append(b, GlazedTerracotta{Colour: c, Facing: dir})
		}
	}
	return b
}

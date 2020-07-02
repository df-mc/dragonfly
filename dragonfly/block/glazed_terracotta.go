package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/colour"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// GlazedTerracotta is a vibrant solid block that comes in the 16 regular dye colors.
type GlazedTerracotta struct {
	// Colour specifies the colour of the block.
	Colour colour.Colour
	// Facing specifies the face of the block.
	Facing world.Face
}

// BreakInfo ...
func (t GlazedTerracotta) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    1.4,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(t, 1)),
	}
}

// EncodeItem ...
func (t GlazedTerracotta) EncodeItem() (id int32, meta int16) {
	// Item ID for glazed terracotta is equal to 220 + color number.
	return int32(220 + t.Colour.Uint8()), meta
}

// EncodeBlock ...
func (t GlazedTerracotta) EncodeBlock() (name string, properties map[string]interface{}) {
	var colourName string
	if t.Colour == colour.LightGrey() {
		// Light grey is actually called "silver" in the block name. Mojang pls.
		colourName = "silver"
	} else {
		colourName = t.Colour.String()
	}
	return "minecraft:" + colourName + "_glazed_terracotta", map[string]interface{}{"facing_direction": int32(t.Facing)}
}

// UseOnBlock ensures the proper facing is used when placing a glazed terracotta block, by using the opposite of the player.
func (t GlazedTerracotta) UseOnBlock(pos world.BlockPos, _ world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, user.Facing().Opposite().Face(), t)
	if !used {
		return
	}

	place(w, pos, t, user, ctx)
	return placed(ctx)
}

// allGlazedTerracotta returns glazed terracotta blocks with all possible colours.
func allGlazedTerracotta() []world.Block {
	b := make([]world.Block, 0, 16)
	for _, c := range colour.All() {
		b = append(b, GlazedTerracotta{Colour: c})
	}
	return b
}

package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Carpet is a colourful block that can be obtained by killing/shearing sheep, or crafted using four string.
type Carpet struct {
	carpet
	transparent

	// Colour is the colour of the carpet.
	Colour item.Colour
}

// FlammabilityInfo ...
func (c Carpet) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 60, true)
}

// CanDisplace ...
func (Carpet) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (Carpet) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// BreakInfo ...
func (c Carpet) BreakInfo() BreakInfo {
	return newBreakInfo(0.1, alwaysHarvestable, nothingEffective, oneOf(c))
}

// EncodeItem ...
func (c Carpet) EncodeItem() (name string, meta int16) {
	return "minecraft:carpet", int16(c.Colour.Uint8())
}

// EncodeBlock ...
func (c Carpet) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:carpet", map[string]any{"color": c.Colour.String()}
}

// HasLiquidDrops ...
func (Carpet) HasLiquidDrops() bool {
	return true
}

// NeighbourUpdateTick ...
func (Carpet) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if _, ok := w.Block(pos.Add(cube.Pos{0, -1})).(Air); ok {
		w.SetBlock(pos, nil, nil)
	}
}

// UseOnBlock handles not placing carpets on top of air blocks.
func (c Carpet) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, c)
	if !used {
		return
	}

	if _, ok := w.Block((cube.Pos{pos.X(), pos.Y() - 1, pos.Z()})).(Air); ok {
		return
	}

	place(w, pos, c, user, ctx)
	return placed(ctx)
}

// allCarpet ...
func allCarpet() (carpets []world.Block) {
	for _, c := range item.Colours() {
		carpets = append(carpets, Carpet{Colour: c})
	}
	return
}

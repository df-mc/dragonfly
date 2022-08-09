package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// MuddyMangroveRoots are a decorative variant of mangrove roots.
type MuddyMangroveRoots struct {
	solid

	// Axis is the axis which the basalt faces.
	Axis cube.Axis
}

// BreakInfo ...
func (m MuddyMangroveRoots) BreakInfo() BreakInfo {
	return newBreakInfo(0.7, alwaysHarvestable, shovelEffective, oneOf(m))
}

// SoilFor ...
func (MuddyMangroveRoots) SoilFor(block world.Block) bool {
	switch block.(type) {
	case TallGrass, DoubleTallGrass, Flower, DoubleFlower, NetherSprouts:
		return true
	}
	return false
}

// UseOnBlock ...
func (m MuddyMangroveRoots) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, m)
	if !used {
		return
	}
	m.Axis = face.Axis()

	place(w, pos, m, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (MuddyMangroveRoots) EncodeItem() (name string, meta int16) {
	return "minecraft:muddy_mangrove_roots", 0
}

// EncodeBlock ...
func (m MuddyMangroveRoots) EncodeBlock() (string, map[string]any) {
	return "minecraft:muddy_mangrove_roots", map[string]any{"pillar_axis": m.Axis.String()}
}

// allMuddyMangroveRoots ...
func allMuddyMangroveRoots() (roots []world.Block) {
	for _, axis := range cube.Axes() {
		roots = append(roots, MuddyMangroveRoots{Axis: axis})
	}
	return
}

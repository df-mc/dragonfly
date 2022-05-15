package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Anvil is a block that allows players to repair items, rename items, and combine enchantments.
type Anvil struct {
	gravityAffected
	solid
	transparent

	// Type is the type of anvil.
	Type AnvilType
	// Facing is the direction that the anvil is facing.
	Facing cube.Direction
}

// BreakInfo ...
func (a Anvil) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(a))
}

// Activate ...
func (a Anvil) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
		return true
	}
	return false
}

// UseOnBlock ...
func (a Anvil) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, a)
	if !used {
		return
	}
	a.Facing = user.Facing().RotateRight()
	place(w, pos, a, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (a Anvil) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	a.fall(a, pos, w)
}

// EncodeItem ...
func (a Anvil) EncodeItem() (name string, meta int16) {
	return "minecraft:anvil", int16(a.Type.Uint8())
}

// EncodeBlock ...
func (a Anvil) EncodeBlock() (string, map[string]any) {
	return "minecraft:anvil", map[string]any{
		"damage":    a.Type.String(),
		"direction": int32(a.Facing.Horizontal()),
	}
}

// allAnvils ...
func allAnvils() (anvils []world.Block) {
	for _, t := range AnvilTypes() {
		for _, d := range cube.Directions() {
			anvils = append(anvils, Anvil{Type: t, Facing: d})
		}
	}
	return
}

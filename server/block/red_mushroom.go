package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// RedMushroom is a transparent plant block which can be used as decoration.
type RedMushroom struct {
	transparent
	empty
}

// BreakInfo ...
func (r RedMushroom) BreakInfo() BreakInfo {
	return newBreakInfo(0, nothingEffective, alwaysHarvestable, oneOf(r))
}

// CompostChance ...
func (r RedMushroom) CompostChance() float64 {
	return 0.65
}

// NeighbourUpdateTick ...
func (r RedMushroom) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		breakBlock(r, pos, tx)
	}
}

// HasLiquidDrops ...
func (r RedMushroom) HasLiquidDrops() bool {
	return true
}

// UseOnBlock ...
func (r RedMushroom) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, r)
	if !used {
		return false
	}
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		return false
	}

	place(tx, pos, r, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (r RedMushroom) EncodeItem() (name string, meta int16) {
	return "minecraft:red_mushroom", 0
}

// EncodeBlock ...
func (r RedMushroom) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:red_mushroom", nil
}

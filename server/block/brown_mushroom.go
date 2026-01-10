package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// BrownMushroom is a transparent plant block which can be used as decoration.
type BrownMushroom struct {
	transparent
	empty
}

// BreakInfo ...
func (b BrownMushroom) BreakInfo() BreakInfo {
	return newBreakInfo(0, nothingEffective, alwaysHarvestable, oneOf(b))
}

// CompostChance ...
func (b BrownMushroom) CompostChance() float64 {
	return 0.65
}

// NeighbourUpdateTick ...
func (b BrownMushroom) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		breakBlock(b, pos, tx)
	}
}

// HasLiquidDrops ...
func (b BrownMushroom) HasLiquidDrops() bool {
	return true
}

// UseOnBlock ...
func (b BrownMushroom) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, b)
	if !used {
		return false
	}
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		return false
	}

	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (b BrownMushroom) EncodeItem() (name string, meta int16) {
	return "minecraft:brown_mushroom", 0
}

// EncodeBlock ...
func (b BrownMushroom) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:brown_mushroom", nil
}

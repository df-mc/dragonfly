package block

import (
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Bush is a transparent plant block which can be used to obtain seeds and as decoration.
type ShortDryGrass struct {
	replaceable
	transparent
	empty
}

// FlammabilityInfo ...
func (s ShortDryGrass) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// BreakInfo ...
func (s ShortDryGrass) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, func(t item.Tool) bool {
		return t.ToolType() == item.TypeShears
	}, nothingEffective, oneOf(s))
}

// BoneMeal attempts to affect the block using a bone meal item.
func (s ShortDryGrass) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	tx.SetBlock(pos, TallDryGrass{}, nil)
	return true
}

// FuelInfo ...
func (s ShortDryGrass) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 5)
}

// CompostChance ...
func (s ShortDryGrass) CompostChance() float64 {
	return 0.3
}

// NeighbourUpdateTick ...
func (s ShortDryGrass) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		breakBlock(s, pos, tx)
	}
}

// HasLiquidDrops ...
func (s ShortDryGrass) HasLiquidDrops() bool {
	return false
}

// UseOnBlock ...
func (s ShortDryGrass) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, s)
	if !used {
		return false
	}
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		return false
	}

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (s ShortDryGrass) EncodeItem() (name string, meta int16) {
	return "minecraft:short_dry_grass", 0
}

// EncodeBlock ...
func (s ShortDryGrass) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:short_dry_grass", nil
}

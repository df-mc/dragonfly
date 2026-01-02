package block

import (
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// LeafLitter is a decorative block that generates in forest biomes.
type LeafLitter struct {
	empty
	transparent

	// AdditionalCount is the amount of additional leaf litter. This can range
	// from 0-7, where only 0-3 can occur in-game.
	AdditionalCount int
	// Facing is the direction the leaf litter are facing. This is opposite to
	// the direction the player is facing when placing.
	Facing cube.Direction
}

// FuelInfo ...
func (l LeafLitter) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 5)
}

// UseOnBlock ...
func (l LeafLitter) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	if existing, ok := tx.Block(pos).(LeafLitter); ok {
		if existing.AdditionalCount >= 3 {
			return false
		}

		existing.AdditionalCount++
		place(tx, pos, existing, user, ctx)
		return placed(ctx)
	}

	pos, _, used := firstReplaceable(tx, pos, face, l)
	if !used {
		return false
	}
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		return false
	}

	l.Facing = user.Rotation().Direction().Opposite()
	place(tx, pos, l, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (l LeafLitter) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		breakBlock(l, pos, tx)
	}
}

// HasLiquidDrops ...
func (LeafLitter) HasLiquidDrops() bool {
	return true
}

// BreakInfo ...
func (l LeafLitter) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, simpleDrops(item.NewStack(l, l.AdditionalCount+1)))
}

// FlammabilityInfo ...
func (l LeafLitter) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 100, true)
}

// CompostChance ...
func (LeafLitter) CompostChance() float64 {
	return 0.3
}

// EncodeItem ...
func (LeafLitter) EncodeItem() (name string, meta int16) {
	return "minecraft:leaf_litter", 0
}

// EncodeBlock ...
func (l LeafLitter) EncodeBlock() (string, map[string]any) {
	return "minecraft:leaf_litter", map[string]any{"growth": int32(l.AdditionalCount), "minecraft:cardinal_direction": l.Facing.String()}
}

// LeafLitter ...
func allLeafLitter() (b []world.Block) {
	for i := 0; i <= 7; i++ {
		for _, d := range cube.Directions() {
			b = append(b, LeafLitter{AdditionalCount: i, Facing: d})
		}
	}
	return
}

package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// WildFlowers is a decorative block that generates in flower biomes.
type WildFlowers struct {
	empty
	transparent

	// AdditionalCount is the amount of additional wild flowers. This can range
	// from 0-7, where only 0-3 can occur in-game.
	AdditionalCount int
	// Facing is the direction the wild flowers are facing. This is opposite to
	// the direction the player is facing when placing.
	Facing cube.Direction
}

// UseOnBlock ...
func (w WildFlowers) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	if existing, ok := tx.Block(pos).(WildFlowers); ok {
		if existing.AdditionalCount >= 3 {
			return false
		}

		existing.AdditionalCount++
		place(tx, pos, existing, user, ctx)
		return placed(ctx)
	}

	pos, _, used := firstReplaceable(tx, pos, face, w)
	if !used {
		return false
	}
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		return false
	}

	w.Facing = user.Rotation().Direction().Opposite()
	place(tx, pos, w, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (w WildFlowers) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceDown.Opposite(), tx) {
		breakBlock(w, pos, tx)
	}
}

// HasLiquidDrops ...
func (WildFlowers) HasLiquidDrops() bool {
	return true
}

// BreakInfo ...
func (w WildFlowers) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, simpleDrops(item.NewStack(w, w.AdditionalCount+1)))
}

// FlammabilityInfo ...
func (w WildFlowers) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 100, true)
}

// CompostChance ...
func (WildFlowers) CompostChance() float64 {
	return 0.3
}

// EncodeItem ...
func (WildFlowers) EncodeItem() (name string, meta int16) {
	return "minecraft:wildflowers", 0
}

// EncodeBlock ...
func (w WildFlowers) EncodeBlock() (string, map[string]any) {
	return "minecraft:wildflowers", map[string]any{"growth": int32(w.AdditionalCount), "minecraft:cardinal_direction": w.Facing.String()}
}

// WildFlowers ...
func allWildFlowers() (b []world.Block) {
	for i := 0; i <= 7; i++ {
		for _, d := range cube.Directions() {
			b = append(b, WildFlowers{AdditionalCount: i, Facing: d})
		}
	}
	return
}

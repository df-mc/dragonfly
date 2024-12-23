package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// PinkPetals is a decorative block that generates in cherry grove biomes.
type PinkPetals struct {
	empty
	transparent

	// AdditionalCount is the amount of additional pink petals. This can range
	// from 0-7, where only 0-3 can occur in-game.
	AdditionalCount int
	// Facing is the direction the pink petals are facing. This is opposite to
	// the direction the player is facing when placing.
	Facing cube.Direction
}

// BoneMeal ...
func (p PinkPetals) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	if p.AdditionalCount < 3 {
		p.AdditionalCount++
		tx.SetBlock(pos, p, nil)
		return true
	}
	dropItem(tx, item.NewStack(p, 1), pos.Vec3Centre())
	return true
}

// UseOnBlock ...
func (p PinkPetals) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	if existing, ok := tx.Block(pos).(PinkPetals); ok {
		if existing.AdditionalCount >= 3 {
			return false
		}

		existing.AdditionalCount++
		place(tx, pos, existing, user, ctx)
		return placed(ctx)
	}

	pos, _, used := firstReplaceable(tx, pos, face, p)
	if !used {
		return false
	}
	if !supportsVegetation(p, tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}

	p.Facing = user.Rotation().Direction().Opposite()
	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (p PinkPetals) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsVegetation(p, tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(p, pos, tx)
	}
}

// HasLiquidDrops ...
func (PinkPetals) HasLiquidDrops() bool {
	return true
}

// BreakInfo ...
func (p PinkPetals) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, simpleDrops(item.NewStack(p, p.AdditionalCount+1)))
}

// FlammabilityInfo ...
func (p PinkPetals) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 100, true)
}

// CompostChance ...
func (PinkPetals) CompostChance() float64 {
	return 0.3
}

// EncodeItem ...
func (PinkPetals) EncodeItem() (name string, meta int16) {
	return "minecraft:pink_petals", 0
}

// EncodeBlock ...
func (p PinkPetals) EncodeBlock() (string, map[string]any) {
	return "minecraft:pink_petals", map[string]any{"growth": int32(p.AdditionalCount), "minecraft:cardinal_direction": p.Facing.String()}
}

// allPinkPetals ...
func allPinkPetals() (b []world.Block) {
	for i := 0; i <= 7; i++ {
		for _, d := range cube.Directions() {
			b = append(b, PinkPetals{AdditionalCount: i, Facing: d})
		}
	}
	return
}

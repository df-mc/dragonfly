package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// SoulSand is a block found naturally only in the Nether. SoulSand slows movement of mobs & players.
type SoulSand struct {
	solid
}

// NeighbourUpdateTick schedules the bubble column above the soul sand to be updated when it or the block above changes.
func (s SoulSand) NeighbourUpdateTick(pos, changedNeighbour cube.Pos, tx *world.Tx) {
	if changedNeighbour == pos || changedNeighbour == pos.Side(cube.FaceUp) {
		tx.ScheduleBlockUpdate(pos, s, 0)
	}
}

// ScheduledTick updates the bubble column above the soul sand.
func (SoulSand) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	updateBubbleColumn(pos.Side(cube.FaceUp), tx)
}

// RandomTick forms or revalidates the bubble column above soul sand loaded without block callbacks.
func (SoulSand) RandomTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	updateBubbleColumn(pos.Side(cube.FaceUp), tx)
}

// SoilFor ...
func (s SoulSand) SoilFor(block world.Block) bool {
	flower, ok := block.(Flower)
	return ok && flower.Type == WitherRose()
}

// Instrument ...
func (s SoulSand) Instrument() sound.Instrument {
	return sound.CowBell()
}

// BreakInfo ...
func (s SoulSand) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, shovelEffective, oneOf(s))
}

// EncodeItem ...
func (SoulSand) EncodeItem() (name string, meta int16) {
	return "minecraft:soul_sand", 0
}

// EncodeBlock ...
func (SoulSand) EncodeBlock() (string, map[string]any) {
	return "minecraft:soul_sand", nil
}

package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// SuspiciousSand is a fragile, gravity affected block holding an item that may be brushed out of it.
type SuspiciousSand struct {
	gravityAffected
	solid
	snare
	brushProgress

	// Hanging is a Bedrock Edition state Dragonfly never sets itself. It is kept so blocks loaded from a
	// world encode back to the same block state.
	Hanging bool
	// Dust is the amount of dust brushed off the block, from 0 to 3. It is updated while brushing.
	Dust int
	// Item is the item brushed out of the block. Blocks placed by a player hold none.
	Item item.Stack
	// LootTable is the loot table the item held by the block is rolled from when it is first brushed.
	// Dragonfly has no loot tables, so it is only kept to avoid losing it when a vanilla world is saved.
	LootTable string
	// LootTableSeed is the seed LootTable is rolled with. A seed of 0 means a random seed is used.
	LootTableSeed int32
}

// Brush ...
func (s SuspiciousSand) Brush(pos cube.Pos, tx *world.Tx, face cube.Face) bool {
	s.brushProgress = s.advance(s.Dust, tx.CurrentTick(), face)
	if s.completed() {
		tx.SetBlock(pos, Sand{}, nil)
		tx.PlaySound(pos.Vec3Centre(), sound.BrushingCompleted{Block: s})
		if !s.Item.Empty() {
			dropItem(tx, s.Item, BrushOffset(pos, s.face))
		}
		return true
	}
	s.Dust = s.dust()
	tx.SetBlock(pos, s, nil)
	tx.ScheduleBlockUpdate(pos, s, brushDecayInterval)
	return false
}

// ScheduledTick ...
func (s SuspiciousSand) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if s.count == 0 {
		return
	}
	if tx.CurrentTick() < s.resetsAt {
		tx.ScheduleBlockUpdate(pos, s, brushDecayInterval)
		return
	}
	s.brushProgress = s.decayed()
	s.Dust = s.dust()
	tx.SetBlock(pos, s, nil)
	if s.count > 0 {
		tx.ScheduleBlockUpdate(pos, s, brushDecayInterval)
	}
}

// SoilFor ...
func (SuspiciousSand) SoilFor(block world.Block) bool {
	switch block.(type) {
	case Cactus, DeadBush, SugarCane, BambooSapling, Bamboo:
		return true
	}
	return false
}

// NeighbourUpdateTick ...
func (s SuspiciousSand) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	s.fall(s, pos, tx)
}

// Landed ...
func (s SuspiciousSand) Landed(tx *world.Tx, pos cube.Pos) {
	tx.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: s})
}

// BreaksOnLanding ...
func (SuspiciousSand) BreaksOnLanding() bool {
	return true
}

// BreakInfo ...
func (SuspiciousSand) BreakInfo() BreakInfo {
	return newBreakInfo(0.25, alwaysHarvestable, shovelEffective, simpleDrops())
}

// EncodeItem ...
func (SuspiciousSand) EncodeItem() (name string, meta int16) {
	return "minecraft:suspicious_sand", 0
}

// EncodeBlock ...
func (s SuspiciousSand) EncodeBlock() (string, map[string]any) {
	return "minecraft:suspicious_sand", map[string]any{"brushed_progress": int32(s.Dust), "hanging": boolByte(s.Hanging)}
}

// EncodeNBT ...
func (s SuspiciousSand) EncodeNBT() map[string]any {
	name, _ := s.EncodeItem()
	m := s.encodeNBT(name, s.Item)
	if s.LootTable != "" {
		m["LootTable"], m["LootTableSeed"] = s.LootTable, s.LootTableSeed
	}
	return m
}

// DecodeNBT ...
func (s SuspiciousSand) DecodeNBT(data map[string]any) any {
	s.Item = item.MapNBT(data, "item")
	s.LootTable, s.LootTableSeed = nbtconv.String(data, "LootTable"), nbtconv.Int32(data, "LootTableSeed")
	s.brushProgress = s.decodeNBT(data)
	return s
}

// allSuspiciousSand ...
func allSuspiciousSand() (blocks []world.Block) {
	for _, hanging := range []bool{false, true} {
		for dust := range 4 {
			blocks = append(blocks, SuspiciousSand{Hanging: hanging, Dust: dust})
		}
	}
	return
}

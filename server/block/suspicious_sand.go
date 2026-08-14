package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// SuspiciousSand is a fragile, gravity affected block that holds a single item which may be extracted from it
// by brushing it with a brush.
type SuspiciousSand struct {
	gravityAffected
	solid
	snare
	brushProgress

	// Hanging specifies if the block has no block below it. It is a Bedrock Edition specific state which
	// Dragonfly itself never sets: It is only kept so that blocks loaded from a world that has it set encode
	// back to the same block state.
	Hanging bool
	// Dust is the amount of dust brushed off the block, ranging from 0 to 3. It is updated automatically as
	// the block is brushed and should not normally be set manually.
	Dust int
	// Item is the item extracted from the block once brushing it completes. Suspicious blocks placed by a
	// player hold no item and produce nothing when brushed.
	Item item.Stack
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
	tx.PlaySound(pos.Vec3Centre(), sound.BlockBreaking{Block: s})
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
	return s.encodeNBT(name, s.Item)
}

// DecodeNBT ...
func (s SuspiciousSand) DecodeNBT(data map[string]any) any {
	s.Item = item.MapNBT(data, "item")
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

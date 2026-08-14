package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// SuspiciousGravel is a fragile, gravity affected block holding an item that may be brushed out of it.
type SuspiciousGravel struct {
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
}

// Brush ...
func (s SuspiciousGravel) Brush(pos cube.Pos, tx *world.Tx, face cube.Face) bool {
	s.brushProgress = s.advance(s.Dust, tx.CurrentTick(), face)
	if s.completed() {
		tx.SetBlock(pos, Gravel{}, nil)
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
func (s SuspiciousGravel) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
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
func (SuspiciousGravel) SoilFor(block world.Block) bool {
	switch block.(type) {
	case BambooSapling, Bamboo:
		return true
	}
	return false
}

// NeighbourUpdateTick ...
func (s SuspiciousGravel) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	s.fall(s, pos, tx)
}

// Landed ...
func (s SuspiciousGravel) Landed(tx *world.Tx, pos cube.Pos) {
	tx.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: s})
	tx.PlaySound(pos.Vec3Centre(), sound.BlockBreaking{Block: s})
}

// BreaksOnLanding ...
func (SuspiciousGravel) BreaksOnLanding() bool {
	return true
}

// BreakInfo ...
func (SuspiciousGravel) BreakInfo() BreakInfo {
	return newBreakInfo(0.25, alwaysHarvestable, shovelEffective, simpleDrops())
}

// EncodeItem ...
func (SuspiciousGravel) EncodeItem() (name string, meta int16) {
	return "minecraft:suspicious_gravel", 0
}

// EncodeBlock ...
func (s SuspiciousGravel) EncodeBlock() (string, map[string]any) {
	return "minecraft:suspicious_gravel", map[string]any{"brushed_progress": int32(s.Dust), "hanging": boolByte(s.Hanging)}
}

// EncodeNBT ...
func (s SuspiciousGravel) EncodeNBT() map[string]any {
	name, _ := s.EncodeItem()
	return s.encodeNBT(name, s.Item)
}

// DecodeNBT ...
func (s SuspiciousGravel) DecodeNBT(data map[string]any) any {
	s.Item = item.MapNBT(data, "item")
	s.brushProgress = s.decodeNBT(data)
	return s
}

// allSuspiciousGravel ...
func allSuspiciousGravel() (blocks []world.Block) {
	for _, hanging := range []bool{false, true} {
		for dust := range 4 {
			blocks = append(blocks, SuspiciousGravel{Hanging: hanging, Dust: dust})
		}
	}
	return
}

package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	// brushesRequired is the number of brushing actions needed to fully brush a SuspiciousBlock.
	brushesRequired = 10
	// brushResetDelay is the amount of ticks a SuspiciousBlock keeps its brushing progress after it stops
	// being brushed.
	brushResetDelay = 40
	// brushDecayInterval is the time between two brushing progress losses of a SuspiciousBlock that is no
	// longer being brushed.
	brushDecayInterval = 2 * time.Second / 20
	// brushDirectionNone is the value written to the brush_direction NBT tag of a SuspiciousBlock that has not
	// been brushed yet.
	brushDirectionNone = 6
)

// Brushable represents a block that may be brushed with a brush to extract an item from it.
type Brushable interface {
	world.Block
	// Brush performs a single brushing action on the block at the position passed, brushed from the face
	// passed. It returns true if the block was fully brushed as a result.
	Brush(pos cube.Pos, tx *world.Tx, face cube.Face) bool
}

// SuspiciousBlock is a fragile, gravity affected block that holds a single item which may be extracted from it
// by brushing it with a brush. It comes in a sand and a gravel variant.
type SuspiciousBlock struct {
	gravityAffected
	solid
	snare

	// Gravel toggles the suspicious gravel variant of the block. If false, the block is suspicious sand.
	Gravel bool
	// Hanging specifies if the block has no block below it. It is a Bedrock Edition specific state which
	// Dragonfly itself never sets: it is only kept so that blocks loaded from a world that has it set encode
	// back to the same block state.
	Hanging bool
	// Dust is the amount of dust brushed off the block, ranging from 0 to 3. It is updated automatically as
	// the block is brushed and should not normally be set manually.
	Dust int
	// Item is the item extracted from the block once brushing it completes. Suspicious blocks placed by a
	// player hold no item and produce nothing when brushed.
	Item item.Stack

	brushCount int
	brushFace  cube.Face
	resetsAt   int64
}

// Brush ...
func (s SuspiciousBlock) Brush(pos cube.Pos, tx *world.Tx, face cube.Face) bool {
	if s.brushCount == 0 {
		s.brushFace, s.brushCount = face, brushStageStart(s.Dust)
	}
	s.brushCount++
	s.resetsAt = tx.CurrentTick() + brushResetDelay

	if s.brushCount >= brushesRequired {
		tx.SetBlock(pos, s.brushed(), nil)
		tx.PlaySound(pos.Vec3Centre(), sound.BrushingCompleted{Block: s})
		if !s.Item.Empty() {
			dropItem(tx, s.Item, BrushOffset(pos, s.brushFace))
		}
		return true
	}

	s.Dust = brushDust(s.brushCount)
	tx.SetBlock(pos, s, nil)
	tx.ScheduleBlockUpdate(pos, s, brushDecayInterval)
	return false
}

// BrushOffset returns the position just outside the face of the block at the position passed. It is the
// position that brushing particles are displayed at and that items brushed out of a block drop at.
func BrushOffset(pos cube.Pos, face cube.Face) mgl64.Vec3 {
	return pos.Vec3Centre().Add(pos.Side(face).Vec3().Sub(pos.Vec3()).Mul(0.6))
}

// ScheduledTick ...
func (s SuspiciousBlock) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if s.brushCount == 0 {
		return
	}
	if tx.CurrentTick() < s.resetsAt {
		tx.ScheduleBlockUpdate(pos, s, brushDecayInterval)
		return
	}
	s.brushCount--
	s.Dust = brushDust(s.brushCount)
	tx.SetBlock(pos, s, nil)
	if s.brushCount > 0 {
		tx.ScheduleBlockUpdate(pos, s, brushDecayInterval)
	}
}

// brushed returns the block that the SuspiciousBlock turns into once brushing it completes.
func (s SuspiciousBlock) brushed() world.Block {
	if s.Gravel {
		return Gravel{}
	}
	return Sand{}
}

// SoilFor ...
func (s SuspiciousBlock) SoilFor(block world.Block) bool {
	switch block.(type) {
	case BambooSapling, Bamboo:
		return true
	case Cactus, DeadBush, SugarCane:
		return !s.Gravel
	}
	return false
}

// NeighbourUpdateTick ...
func (s SuspiciousBlock) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	s.fall(s, pos, tx)
}

// Landed ...
func (s SuspiciousBlock) Landed(tx *world.Tx, pos cube.Pos) {
	tx.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: s})
	tx.PlaySound(pos.Vec3Centre(), sound.BlockBreaking{Block: s})
}

// BreaksOnLanding ...
func (s SuspiciousBlock) BreaksOnLanding() bool {
	return true
}

// BreakInfo ...
func (s SuspiciousBlock) BreakInfo() BreakInfo {
	return newBreakInfo(0.25, alwaysHarvestable, shovelEffective, simpleDrops())
}

// EncodeItem ...
func (s SuspiciousBlock) EncodeItem() (name string, meta int16) {
	if s.Gravel {
		return "minecraft:suspicious_gravel", 0
	}
	return "minecraft:suspicious_sand", 0
}

// EncodeBlock ...
func (s SuspiciousBlock) EncodeBlock() (string, map[string]any) {
	properties := map[string]any{"brushed_progress": int32(s.Dust), "hanging": boolByte(s.Hanging)}
	if s.Gravel {
		return "minecraft:suspicious_gravel", properties
	}
	return "minecraft:suspicious_sand", properties
}

// EncodeNBT ...
func (s SuspiciousBlock) EncodeNBT() map[string]any {
	name, _ := s.EncodeItem()
	direction := byte(brushDirectionNone)
	if s.brushCount > 0 {
		direction = byte(s.brushFace)
	}
	m := map[string]any{
		"id":              "BrushableBlock",
		"type":            name,
		"brush_count":     int32(s.brushCount),
		"brush_direction": direction,
	}
	if !s.Item.Empty() {
		m["item"] = item.WriteNBT(s.Item, true)
	}
	return m
}

// DecodeNBT ...
func (s SuspiciousBlock) DecodeNBT(data map[string]any) any {
	s.Item = item.MapNBT(data, "item")
	s.brushCount = min(max(int(nbtconv.Int32(data, "brush_count")), 0), brushesRequired-1)
	if direction := nbtconv.Uint8(data, "brush_direction"); direction < brushDirectionNone {
		s.brushFace = cube.Face(direction)
	}
	return s
}

// brushDust returns the amount of dust displayed on a SuspiciousBlock brushed the amount of times passed.
func brushDust(brushCount int) int {
	switch {
	case brushCount <= 0:
		return 0
	case brushCount < 3:
		return 1
	case brushCount < 6:
		return 2
	}
	return 3
}

// brushStageStart returns the lowest brush count that displays the amount of dust passed.
func brushStageStart(dust int) int {
	switch dust {
	case 1:
		return 1
	case 2:
		return 3
	case 3:
		return 6
	}
	return 0
}

// allSuspiciousBlocks ...
func allSuspiciousBlocks() (blocks []world.Block) {
	for _, gravel := range []bool{false, true} {
		for _, hanging := range []bool{false, true} {
			for dust := range 4 {
				blocks = append(blocks, SuspiciousBlock{Gravel: gravel, Hanging: hanging, Dust: dust})
			}
		}
	}
	return
}

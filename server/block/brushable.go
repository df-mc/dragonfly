package block

import (
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
)

const (
	// brushesRequired is the number of brushing actions needed to fully brush a suspicious block.
	brushesRequired = 10
	// brushResetInterval is how long a block keeps its progress after being brushed. Like vanilla, the progress
	// is then checked again brushRecheckInterval later, and from there lost two brushes every brushDecayInterval.
	brushResetInterval   = 2 * time.Second
	brushRecheckInterval = time.Second / 5
	brushDecayInterval   = time.Second / 5
	// brushDirectionNone is the brush_direction of a block that has not been brushed yet.
	brushDirectionNone = 6
)

// Brushable represents a block that may be brushed with a brush to extract an item from it.
type Brushable interface {
	world.Block
	// Brush performs a single brushing action on the block at the position passed, brushed from the face
	// passed. It returns true if the block was fully brushed as a result.
	Brush(pos cube.Pos, tx *world.Tx, face cube.Face) bool
}

// brushProgress is the brushing state a suspicious block keeps track of.
type brushProgress struct {
	count    int
	face     cube.Face
	resetsAt int64
	brushed  bool

	nextBrush int64
}

// advance adds a brushing action at the tick passed, unless the block was brushed 10 ticks ago or less.
func (p brushProgress) advance(tick int64, face cube.Face) (brushProgress, bool) {
	if tick <= p.nextBrush {
		return p, false
	}
	p.nextBrush = tick + 10
	if p.count == 0 {
		p.face = face
	}
	p.count, p.resetsAt, p.brushed = p.count+1, tick+int64(brushResetInterval/(time.Second/20)), true
	return p, true
}

// tick returns the progress after a scheduled update at the tick passed, and the delay before the next one.
func (p brushProgress) tick(tick int64) (brushProgress, time.Duration) {
	if p.brushed {
		if tick <= p.resetsAt {
			return p, brushRecheckInterval
		}
		p.brushed = false
	}
	if p.count == 0 {
		return p, 0
	}
	p.count = max(p.count-2, 0)
	return p, brushDecayInterval
}

// completed returns whether the block was fully brushed.
func (p brushProgress) completed() bool {
	return p.count >= brushesRequired
}

// dust returns the dust to display on the block, from 0 to 3.
func (p brushProgress) dust() int {
	switch {
	case p.count <= 0:
		return 0
	case p.count < 3:
		return 1
	case p.count < 6:
		return 2
	}
	return 3
}

// brushed drops the item held by a fully brushed block in front of the face brushed and replaces the block.
func brushed(pos cube.Pos, tx *world.Tx, b world.Block, it item.Stack, face cube.Face, with world.Block) {
	if !it.Empty() {
		create := tx.World().EntityRegistry().Config().Item
		tx.AddEntity(create(world.EntitySpawnOpts{Position: pos.Side(face).Vec3Centre()}, it))
	}
	tx.PlaySound(pos.Vec3Centre(), sound.BrushingCompleted{Block: b})
	tx.AddParticle(pos.Vec3(), particle.BlockBreakNoSound{Block: b})
	tx.SetBlock(pos, with, nil)
}

// encodeNBT returns block entity data for the identifier and item passed.
func (p brushProgress) encodeNBT(name string, it item.Stack) map[string]any {
	direction := byte(brushDirectionNone)
	if p.count > 0 {
		direction = byte(p.face)
	}
	m := map[string]any{
		"id":              "BrushableBlock",
		"type":            name,
		"brush_count":     int32(p.count),
		"brush_direction": direction,
	}
	if !it.Empty() {
		m["item"] = item.WriteNBT(it, true)
	}
	return m
}

// decodeNBT reads the progress from the block entity data passed.
func (p brushProgress) decodeNBT(data map[string]any) brushProgress {
	p.count = min(max(int(nbtconv.Int32(data, "brush_count")), 0), brushesRequired-1)
	if direction := nbtconv.Uint8(data, "brush_direction"); direction < brushDirectionNone {
		p.face = cube.Face(direction)
	}
	return p
}

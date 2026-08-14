package block

import (
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	// brushesRequired is the number of brushing actions needed to fully brush a suspicious block.
	brushesRequired = 10
	// brushResetDelay is the amount of ticks a suspicious block keeps its brushing progress for after it
	// stops being brushed.
	brushResetDelay = 40
	// brushDecayInterval is the time between two brushing progress losses.
	brushDecayInterval = 2 * time.Second / 20
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

// BrushOffset returns the position just outside the face of the block at the position passed. Brushing
// particles are displayed and items brushed out of a block are dropped at this position.
func BrushOffset(pos cube.Pos, face cube.Face) mgl64.Vec3 {
	return pos.Vec3Centre().Add(pos.Side(face).Vec3().Sub(pos.Vec3()).Mul(0.6))
}

// brushProgress is the brushing state that suspicious blocks keep track of while they are being brushed.
type brushProgress struct {
	count    int
	face     cube.Face
	resetsAt int64
}

// advance adds a single brushing action to the progress. The dust passed is the dust currently displayed on
// the block, which is used as a starting point if the block was not being brushed yet.
func (p brushProgress) advance(dust int, tick int64, face cube.Face) brushProgress {
	if p.count == 0 {
		p.face, p.count = face, brushStageStart(dust)
	}
	p.count, p.resetsAt = p.count+1, tick+brushResetDelay
	return p
}

// completed returns whether enough brushing actions were performed to fully brush the block.
func (p brushProgress) completed() bool {
	return p.count >= brushesRequired
}

// decayed returns the progress with a single brushing action removed from it.
func (p brushProgress) decayed() brushProgress {
	p.count--
	return p
}

// dust returns the amount of dust that should be displayed on the block, ranging from 0 to 3.
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

// encodeNBT returns the block entity data of a suspicious block with the identifier and item passed.
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

// decodeNBT reads the brushing progress from the block entity data passed.
func (p brushProgress) decodeNBT(data map[string]any) brushProgress {
	p.count = min(max(int(nbtconv.Int32(data, "brush_count")), 0), brushesRequired-1)
	if direction := nbtconv.Uint8(data, "brush_direction"); direction < brushDirectionNone {
		p.face = cube.Face(direction)
	}
	return p
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

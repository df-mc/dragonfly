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
	// brushResetDelay is the amount of ticks a block keeps its brushing progress after being brushed.
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

// BrushOffset returns the position just outside the face passed, where brushing particles show and brushed
// items drop.
func BrushOffset(pos cube.Pos, face cube.Face) mgl64.Vec3 {
	return pos.Vec3Centre().Add(pos.Side(face).Vec3().Sub(pos.Vec3()).Mul(0.6))
}

// brushProgress is the brushing state a suspicious block keeps track of.
type brushProgress struct {
	count    int
	face     cube.Face
	resetsAt int64
}

// advance adds a brushing action, starting off at the dust passed if the block was not being brushed yet.
func (p brushProgress) advance(dust int, tick int64, face cube.Face) brushProgress {
	if p.count == 0 {
		p.face, p.count = face, brushStageStart(dust)
	}
	p.count, p.resetsAt = p.count+1, tick+brushResetDelay
	return p
}

// decaying returns the progress along with whether it should lose a brushing action at the tick passed.
// Progress restored from disk has no reset deadline left, in which case a new one is started.
func (p brushProgress) decaying(tick int64) (brushProgress, bool) {
	if p.resetsAt == 0 {
		p.resetsAt = tick + brushResetDelay
		return p, false
	}
	return p, tick >= p.resetsAt
}

// completed returns whether the block was fully brushed.
func (p brushProgress) completed() bool {
	return p.count >= brushesRequired
}

// decayed returns the progress with one brushing action removed.
func (p brushProgress) decayed() brushProgress {
	p.count--
	return p
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

// brushStageStart returns the lowest brush count that displays the dust passed.
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

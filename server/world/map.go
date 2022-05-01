package world

import (
	"image/color"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// registeredMaps is a map that uses ID (int64) as key and NewMap as value.
var registeredMaps sync.Map

// RegisterMap returns the map ID.
// The map will be saved to disk if presistent is true.
func RegisterMap(m NewMap, presistent bool) int64 {
	panic("implement me")
}

// UpdateMap broadcasts the update to all viewers.
// And overrides NewMap by the update's offsets, so new viewers can also receive the update.
// If it is presistent, the updated NewMap will be saved to disk.
func UpdateMap(m UpdatedMap) {
	panic("implement me")
}

type UpdatedMap struct {
	XOffset, YOffset int32

	NewMap
}

type NewMap struct {
	Pixels        [][]color.RGBA
	TrackEntities []Entity
	TrackBlocks   []cube.Pos
}

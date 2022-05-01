package world

import (
	"image/color"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// registeredMaps is a map that uses ID as key and NewMap as value.
var registeredMaps sync.Map

// RegisterMap returns the map ID.
// The map will be saved to disk
func RegisterMap(m NewMap, presistent bool) int64 {
	panic("implement me")
}

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

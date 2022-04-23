package world

import (
	"image/color"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// registeredMaps is a map that uses ID as key and Map as value.
var registeredMaps sync.Map

type Map interface {
	SetPixels([][]color.RGBA)
	TrackEntity(Entity)
	TrackBlock(cube.Pos)

	GetPixels() [][]color.RGBA
	GetTrackedEntities() []Entity
	GetTrackedBlocks() []cube.Pos
}

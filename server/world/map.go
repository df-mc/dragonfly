package world

import (
	"image/color"

	"github.com/df-mc/dragonfly/server/block/cube"
)

type MapData struct {
	Pixels        [][]color.RGBA
	TrackEntities []Entity
	TrackBlocks   []cube.Pos
}

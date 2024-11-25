package chunk

import (
	"github.com/df-mc/dragonfly/server/block/cube"
)

type Column struct {
	Chunk         *Chunk
	Entities      []Entity
	BlockEntities []BlockEntity
}

type BlockEntity struct {
	Pos  cube.Pos
	Data map[string]any
}

type Entity struct {
	ID   int64
	Data map[string]any
}

package chunk

import (
	"github.com/df-mc/dragonfly/server/block/cube"
)

type Column struct {
	Chunk           *Chunk
	Entities        []Entity
	BlockEntities   []BlockEntity
	Tick            int64
	ScheduledBlocks []ScheduledBlockUpdate
	StructureStarts []StructureStart
	StructureRefs   []StructureReference
}

type BlockEntity struct {
	Pos  cube.Pos
	Data map[string]any
}

type Entity struct {
	ID   int64
	Data map[string]any
}

type ScheduledBlockUpdate struct {
	Pos   cube.Pos
	Block uint32
	Tick  int64
}

type StructureReference struct {
	StructureSet string `nbt:"structure_set"`
	Structure    string `nbt:"structure"`
	StartChunkX  int32  `nbt:"start_chunk_x"`
	StartChunkZ  int32  `nbt:"start_chunk_z"`
}

type StructureStart struct {
	StructureReference
	Template string `nbt:"template"`
	OriginX  int32  `nbt:"origin_x"`
	OriginY  int32  `nbt:"origin_y"`
	OriginZ  int32  `nbt:"origin_z"`
	SizeX    int32  `nbt:"size_x"`
	SizeY    int32  `nbt:"size_y"`
	SizeZ    int32  `nbt:"size_z"`
}

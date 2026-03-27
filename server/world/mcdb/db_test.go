package mcdb

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
)

type metadataGenerator struct{}

func (metadataGenerator) GenerateChunk(world.ChunkPos, *chunk.Chunk) {}

func (metadataGenerator) GenerateColumn(pos world.ChunkPos, col *chunk.Column) {
	if col == nil {
		return
	}
	col.StructureStarts = []chunk.StructureStart{{
		StructureReference: chunk.StructureReference{
			StructureSet: "villages",
			Structure:    "village_plains",
			StartChunkX:  pos[0],
			StartChunkZ:  pos[1],
		},
		Template: "village/plains/town_centers/plains_meeting_point_1",
		OriginX:  pos[0] << 4,
		OriginY:  70,
		OriginZ:  pos[1] << 4,
		SizeX:    32,
		SizeY:    20,
		SizeZ:    32,
	}}
	col.StructureRefs = []chunk.StructureReference{{
		StructureSet: "villages",
		Structure:    "village_plains",
		StartChunkX:  pos[0],
		StartChunkZ:  pos[1],
	}}
}

func TestColumnStructureDataRoundTrip(t *testing.T) {
	dir := t.TempDir()

	db, err := Open(dir)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close db: %v", err)
		}
	}()

	pos := world.ChunkPos{12, -7}
	col := &chunk.Column{
		Chunk: chunk.New(0, cube.Range{-64, 319}),
		StructureStarts: []chunk.StructureStart{{
			StructureReference: chunk.StructureReference{
				StructureSet: "villages",
				Structure:    "village_plains",
				StartChunkX:  12,
				StartChunkZ:  -7,
			},
			Template: "village/plains/town_centers/plains_meeting_point_1",
			OriginX:  192,
			OriginY:  70,
			OriginZ:  -112,
			SizeX:    32,
			SizeY:    20,
			SizeZ:    32,
		}},
		StructureRefs: []chunk.StructureReference{{
			StructureSet: "villages",
			Structure:    "village_plains",
			StartChunkX:  12,
			StartChunkZ:  -7,
		}},
	}

	if err := db.StoreColumn(pos, world.Overworld, col); err != nil {
		t.Fatalf("store column: %v", err)
	}
	loaded, err := db.LoadColumn(pos, world.Overworld)
	if err != nil {
		t.Fatalf("load column: %v", err)
	}
	if len(loaded.StructureStarts) != 1 {
		t.Fatalf("expected 1 structure start, got %d", len(loaded.StructureStarts))
	}
	if len(loaded.StructureRefs) != 1 {
		t.Fatalf("expected 1 structure reference, got %d", len(loaded.StructureRefs))
	}
	if loaded.StructureStarts[0].Template != col.StructureStarts[0].Template {
		t.Fatalf("expected template %q, got %q", col.StructureStarts[0].Template, loaded.StructureStarts[0].Template)
	}
	if loaded.StructureRefs[0].StartChunkX != col.StructureRefs[0].StartChunkX || loaded.StructureRefs[0].StartChunkZ != col.StructureRefs[0].StartChunkZ {
		t.Fatalf("unexpected structure reference start chunk: %+v", loaded.StructureRefs[0])
	}
}

func TestGeneratedColumnStructureDataPersistsThroughWorldSave(t *testing.T) {
	dir := t.TempDir()

	db, err := Open(dir)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	w := world.Config{
		Provider:     db,
		Generator:    metadataGenerator{},
		SaveInterval: -1,
	}.New()

	pos := world.ChunkPos{5, -3}
	blockPos := cube.Pos{int(pos[0] << 4), 64, int(pos[1] << 4)}
	<-w.Exec(func(tx *world.Tx) {
		_ = tx.Block(blockPos)
	})
	if err := w.Close(); err != nil {
		t.Fatalf("close world: %v", err)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	defer func() {
		if err := reopened.Close(); err != nil {
			t.Fatalf("close reopened db: %v", err)
		}
	}()

	loaded, err := reopened.LoadColumn(pos, world.Overworld)
	if err != nil {
		t.Fatalf("load generated column: %v", err)
	}
	if len(loaded.StructureStarts) != 1 {
		t.Fatalf("expected 1 persisted structure start, got %d", len(loaded.StructureStarts))
	}
	if len(loaded.StructureRefs) != 1 {
		t.Fatalf("expected 1 persisted structure ref, got %d", len(loaded.StructureRefs))
	}
}

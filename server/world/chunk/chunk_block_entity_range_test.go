package chunk_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
)

func TestChunkClearBlockEntityDataInRange(t *testing.T) {
	t.Parallel()

	ch := chunk.New(world.DefaultBlockRegistry, world.Overworld.Range())
	clearedPos := cube.Pos{32, 64, 48}
	keptPos := cube.Pos{32, 80, 48}
	outsideXZ := cube.Pos{48, 64, 48}

	ch.SetBlockEntityData(clearedPos, map[string]any{"id": "Chest"})
	ch.SetBlockEntityData(keptPos, map[string]any{"id": "Chest"})
	ch.SetBlockEntityData(outsideXZ, map[string]any{"id": "Chest"})

	ch.ClearBlockEntityDataInRange(cube.Pos{32, 64, 48}, cube.Pos{47, 79, 63})

	if _, ok := ch.BlockEntityData(clearedPos); ok {
		t.Fatal("expected block entity data inside range to be removed")
	}
	if _, ok := ch.BlockEntityData(keptPos); !ok {
		t.Fatal("expected block entity data outside Y range to remain")
	}
	if _, ok := ch.BlockEntityData(outsideXZ); !ok {
		t.Fatal("expected block entity data outside X/Z range to remain")
	}
}

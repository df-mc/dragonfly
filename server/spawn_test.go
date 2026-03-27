package server

import (
	"sync"
	"testing"
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	vanilla "github.com/df-mc/dragonfly/server/world/generator/vanilla"
)

func TestAdjustOverworldSpawnMovesOffOceanOrigin(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	w := world.Config{
		Dim:       world.Overworld,
		Provider:  world.NopProvider{},
		Generator: vanilla.New(0),
	}.New()
	defer func() {
		if err := w.Close(); err != nil {
			t.Fatalf("failed to close world: %v", err)
		}
	}()

	srv := &Server{}
	g := vanilla.New(0)
	chunkPos, ok := g.FindSpawnChunk(128)
	if !ok {
		t.Fatal("expected to find a viable spawn hint chunk")
	}
	if !srv.adjustOverworldSpawnHint(w, chunkPos) {
		t.Fatalf("expected spawn relocation to succeed near chunk %v", chunkPos)
	}

	spawn := w.Spawn()
	if spawn == (cube.Pos{}) {
		t.Fatal("unexpected zero spawn")
	}

	var (
		candidate spawnCandidate
		found     bool
	)
	<-w.Exec(func(tx *world.Tx) {
		candidate, found = currentSpawnCandidate(tx, spawn)
	})
	if !found {
		t.Fatalf("expected spawn %v to resolve to a loaded chunk candidate", spawn)
	}
	if candidate.score < 70 {
		t.Fatalf("expected relocated spawn to be in a tree-capable land biome, got score %d at %v", candidate.score, candidate.pos)
	}
	if spawn[0] == 0 && spawn[2] == 0 {
		t.Fatalf("expected spawn to relocate away from oceanic origin, got %v", spawn)
	}
}

var finaliseBlocksOnce sync.Once

//go:linkname worldFinaliseBlockRegistry github.com/df-mc/dragonfly/server/world.finaliseBlockRegistry
func worldFinaliseBlockRegistry()

package portal_test

import (
	"sync"
	"sync/atomic"
	"testing"
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	portal "github.com/df-mc/dragonfly/server/world/portal"
)

func TestNetherPortalFromPosRecognisesFramedPortal(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	w := world.Config{Dim: world.Overworld, Provider: world.NopProvider{}}.New()
	defer closeTestWorld(t, w)

	var (
		found portal.Nether
		ok    bool
	)
	<-w.Exec(func(tx *world.Tx) {
		buildNetherPortalFrame(tx, cube.Pos{0, 64, 0}, cube.Z)
		tx.SetBlock(cube.Pos{1, 65, 0}, block.Fire{}, nil)
		found, ok = portal.NetherPortalFromPos(tx, cube.Pos{1, 65, 0})
	})

	if !ok {
		t.Fatal("expected framed portal to be recognised")
	}
	if !found.Framed() {
		t.Fatal("expected portal scan to report a complete frame")
	}
	width, height := found.Bounds()
	if width != 2 || height != 3 {
		t.Fatalf("expected 2x3 portal interior, got %dx%d", width, height)
	}
}

func TestFindOrCreateNetherPortalCreatesUsablePortal(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	w := world.Config{Dim: world.Nether, Provider: world.NopProvider{}}.New()
	defer closeTestWorld(t, w)

	var (
		created portal.Nether
		ok      bool
		found   portal.Nether
		foundOK bool
		active  bool
		framed  bool
	)
	<-w.Exec(func(tx *world.Tx) {
		created, ok = portal.FindOrCreateNetherPortal(tx, cube.Pos{12, 70, -18}, 16)
		if ok {
			active = created.Activated()
			framed = created.Framed()
		}
		if ok {
			found, foundOK = portal.FindNetherPortal(tx, cube.Pos{12, 70, -18}, 16)
		}
	})

	if !ok {
		t.Fatal("expected portal creation to succeed")
	}
	if !active || !framed {
		t.Fatal("expected created portal to be framed and activated")
	}
	if !foundOK {
		t.Fatal("expected created portal to be discoverable")
	}
	if len(found.Positions()) == 0 {
		t.Fatal("expected discovered portal to have inner positions")
	}
}

func TestFindNetherPortalDoesNotGenerateMissingChunks(t *testing.T) {
	finaliseBlocksOnce.Do(worldFinaliseBlockRegistry)

	gen := &countingGenerator{}
	w := world.Config{
		Dim:       world.Overworld,
		Provider:  world.NopProvider{},
		Generator: gen,
	}.New()
	defer closeTestWorld(t, w)

	var found bool
	<-w.Exec(func(tx *world.Tx) {
		_, found = portal.FindNetherPortal(tx, cube.Pos{0, 64, 0}, 128)
	})

	if found {
		t.Fatal("expected no portal to be found in an empty world")
	}
	if calls := gen.calls.Load(); calls != 0 {
		t.Fatalf("expected missing-chunk portal search not to generate, got %d generation calls", calls)
	}
}

func buildNetherPortalFrame(tx *world.Tx, origin cube.Pos, axis cube.Axis) {
	for width := -1; width < 3; width++ {
		for height := -1; height < 4; height++ {
			pos := origin
			switch axis {
			case cube.X:
				pos = cube.Pos{origin.X(), origin.Y() + height, origin.Z() + width}
			default:
				pos = cube.Pos{origin.X() + width, origin.Y() + height, origin.Z()}
			}
			if width == -1 || width == 2 || height == -1 || height == 3 {
				tx.SetBlock(pos, block.Obsidian{}, nil)
			}
		}
	}
}

func closeTestWorld(t *testing.T, w *world.World) {
	t.Helper()
	if err := w.Close(); err != nil {
		t.Fatalf("failed closing world: %v", err)
	}
}

var finaliseBlocksOnce sync.Once

//go:linkname worldFinaliseBlockRegistry github.com/df-mc/dragonfly/server/world.finaliseBlockRegistry
func worldFinaliseBlockRegistry()

type countingGenerator struct {
	calls atomic.Int32
}

func (g *countingGenerator) GenerateChunk(world.ChunkPos, *chunk.Chunk) {
	g.calls.Add(1)
}

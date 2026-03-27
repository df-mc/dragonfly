package item_test

import (
	"sync"
	"testing"
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestFlintAndSteelActivatesNetherPortalFrame(t *testing.T) {
	finalisePortalItemBlocksOnce.Do(worldFinaliseBlockRegistry)

	w := world.Config{Dim: world.Overworld, Provider: world.NopProvider{}}.New()
	defer closePortalItemWorld(t, w)

	var (
		used          bool
		ctx           item.UseContext
		portalCreated bool
	)
	<-w.Exec(func(tx *world.Tx) {
		buildPortalFrame(tx, cube.Pos{0, 64, 0}, cube.Z)
		user := tx.AddEntity(world.NewEntity(player.Type, player.Config{
			Name:     "flint",
			Position: cube.Pos{1, 64, 0}.Vec3(),
		})).(*player.Player)
		used = (item.FlintAndSteel{}).UseOnBlock(cube.Pos{1, 63, 0}, cube.FaceUp, mgl64.Vec3{}, tx, user, &ctx)
		_, portalCreated = tx.Block(cube.Pos{1, 64, 0}).(block.Portal)
	})

	if !used {
		t.Fatal("expected flint and steel to activate the portal frame")
	}
	if ctx.Damage != 1 {
		t.Fatalf("expected flint and steel use to apply durability damage, got %d", ctx.Damage)
	}
	if !portalCreated {
		t.Fatal("expected portal interior to become an active portal")
	}
}

func TestFireChargeActivatesNetherPortalFrame(t *testing.T) {
	finalisePortalItemBlocksOnce.Do(worldFinaliseBlockRegistry)

	w := world.Config{Dim: world.Overworld, Provider: world.NopProvider{}}.New()
	defer closePortalItemWorld(t, w)

	var (
		used          bool
		ctx           item.UseContext
		portalCreated bool
	)
	<-w.Exec(func(tx *world.Tx) {
		buildPortalFrame(tx, cube.Pos{0, 64, 0}, cube.Z)
		user := tx.AddEntity(world.NewEntity(player.Type, player.Config{
			Name:     "charge",
			Position: cube.Pos{1, 64, 0}.Vec3(),
		})).(*player.Player)
		used = (item.FireCharge{}).UseOnBlock(cube.Pos{1, 63, 0}, cube.FaceUp, mgl64.Vec3{}, tx, user, &ctx)
		_, portalCreated = tx.Block(cube.Pos{1, 64, 0}).(block.Portal)
	})

	if !used {
		t.Fatal("expected fire charge to activate the portal frame")
	}
	if ctx.CountSub != 1 {
		t.Fatalf("expected fire charge use to subtract one item, got %d", ctx.CountSub)
	}
	if !portalCreated {
		t.Fatal("expected portal interior to become an active portal")
	}
}

func buildPortalFrame(tx *world.Tx, origin cube.Pos, axis cube.Axis) {
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
				continue
			}
			tx.SetBlock(pos, nil, nil)
		}
	}
}

func closePortalItemWorld(t *testing.T, w *world.World) {
	t.Helper()
	if err := w.Close(); err != nil {
		t.Fatalf("failed closing world: %v", err)
	}
}

var finalisePortalItemBlocksOnce sync.Once

//go:linkname worldFinaliseBlockRegistry github.com/df-mc/dragonfly/server/world.finaliseBlockRegistry
func worldFinaliseBlockRegistry()

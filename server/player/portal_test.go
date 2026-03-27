package player

import (
	"fmt"
	"sync"
	"testing"
	"time"
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	worldportal "github.com/df-mc/dragonfly/server/world/portal"
	"github.com/go-gl/mathgl/mgl64"
)

func TestNetherPortalTransfersBetweenOverworldAndNether(t *testing.T) {
	finalisePlayerBlocksOnce.Do(worldFinaliseBlockRegistry)

	overworld, nether, end := newPortalTestWorlds()
	defer closePlayerTestWorld(t, overworld)
	defer closePlayerTestWorld(t, nether)
	defer closePlayerTestWorld(t, end)

	handle := world.NewEntity(Type, Config{Name: "portal-tester", Position: cube.Pos{1, 64, 0}.Vec3()})
	var (
		detectedNetherPortal bool
		manualNetherPortal   bool
		netherPortalBlock    bool
		netherInsider        bool
	)

	<-overworld.Exec(func(tx *world.Tx) {
		buildTestNetherPortal(tx, cube.Pos{0, 64, 0}, cube.Z)
		_, netherPortalBlock = tx.Block(cube.Pos{1, 64, 0}).(block.Portal)
		_, netherInsider = tx.Block(cube.Pos{1, 64, 0}).(block.EntityInsider)
		p := tx.AddEntity(handle).(*Player)
		(block.Portal{Axis: cube.Z}).EntityInside(cube.Pos{1, 64, 0}, tx, p)
		manualNetherPortal = p.inNetherPortal
		p.inNetherPortal = false
		p.checkEntityInsiders(Type.BBox(p).Translate(p.Position()))
		detectedNetherPortal = p.inNetherPortal
		for i := 0; i < netherPortalUseTicks; i++ {
			p.Tick(tx, int64(i))
			if i == netherPortalUseTicks-1 {
				return
			}
		}
	})
	if !netherPortalBlock {
		t.Fatal("expected placed nether portal block to decode back as block.Portal")
	}
	if !netherInsider {
		t.Fatal("expected placed nether portal block to satisfy block.EntityInsider")
	}
	if !manualNetherPortal {
		t.Fatal("expected direct nether portal EntityInside call to mark the player")
	}
	if !detectedNetherPortal {
		t.Fatal("expected player collision checks to detect the nether portal block")
	}

	var (
		dimAfterFirst world.Dimension
		firstPos      mgl64.Vec3
	)
	handle.ExecWorld(func(tx *world.Tx, e world.Entity) {
		dimAfterFirst = tx.World().Dimension()
		firstPos = e.Position()
	})
	if dimAfterFirst != world.Nether {
		t.Fatalf("expected player in Nether after portal transfer, got %v", dimAfterFirst)
	}

	var foundPortal bool
	handle.ExecWorld(func(tx *world.Tx, e world.Entity) {
		p := e.(*Player)
		p.portalCooldownUntil = time.Time{}
		n, ok := worldportal.FindNetherPortal(tx, cube.PosFromVec3(firstPos), 32)
		if !ok || len(n.Positions()) == 0 {
			return
		}
		foundPortal = true
		inside := lowestPortalPos(n.Positions())
		p.teleport(inside.Vec3())
		for i := 0; i < netherPortalUseTicks; i++ {
			p.Tick(tx, int64(i))
			if i == netherPortalUseTicks-1 {
				return
			}
		}
	})
	if !foundPortal {
		t.Fatal("expected return portal to exist in the Nether")
	}

	var dimAfterReturn world.Dimension
	handle.ExecWorld(func(tx *world.Tx, _ world.Entity) {
		dimAfterReturn = tx.World().Dimension()
	})
	if dimAfterReturn != world.Overworld {
		t.Fatalf("expected player back in the Overworld after return trip, got %v", dimAfterReturn)
	}
}

func TestEndPortalActivationAndTravel(t *testing.T) {
	finalisePlayerBlocksOnce.Do(worldFinaliseBlockRegistry)

	overworld, nether, end := newPortalTestWorlds()
	defer closePlayerTestWorld(t, overworld)
	defer closePlayerTestWorld(t, nether)
	defer closePlayerTestWorld(t, end)

	center := cube.Pos{0, 64, 0}
	missingFrame := center.Add(cube.Pos{0, 0, 2})
	handle := world.NewEntity(Type, Config{Name: "ender", Position: center.Vec3()})

	var (
		portalFilled   bool
		useErr         error
		detectedEnd    bool
		manualEnd      bool
		endPortalBlock bool
		endInsider     bool
	)
	<-overworld.Exec(func(tx *world.Tx) {
		buildEndPortalRing(tx, center, missingFrame)
		p := tx.AddEntity(handle).(*Player)
		ctx := &item.UseContext{}
		if !(item.EyeOfEnder{}).UseOnBlock(missingFrame, cube.FaceUp, mgl64.Vec3{}, tx, p, ctx) {
			useErr = errString("expected eye of ender use on frame to succeed")
			return
		}
		if ctx.CountSub != 1 {
			useErr = errStringf("expected eye use to subtract one item, got %d", ctx.CountSub)
			return
		}
		portalFilled = isEndPortalFilled(tx, center)
		_, endPortalBlock = tx.Block(center).(block.EndPortal)
		_, endInsider = tx.Block(center).(block.EntityInsider)
		(block.EndPortal{}).EntityInside(center, tx, p)
		manualEnd = p.inEndPortal
		p.inEndPortal = false
		p.checkEntityInsiders(Type.BBox(p).Translate(p.Position()))
		detectedEnd = p.inEndPortal
		p.Tick(tx, 0)
	})
	if useErr != nil {
		t.Fatal(useErr)
	}
	if !portalFilled {
		t.Fatal("expected completed frame ring to fill the 3x3 end portal")
	}
	if !endPortalBlock {
		t.Fatal("expected activated end portal to decode back as block.EndPortal")
	}
	if !endInsider {
		t.Fatal("expected activated end portal block to satisfy block.EntityInsider")
	}
	if !manualEnd {
		t.Fatal("expected direct end portal EntityInside call to mark the player")
	}
	if !detectedEnd {
		t.Fatal("expected player collision checks to detect the end portal block")
	}

	var dimAfterFirst world.Dimension
	var (
		endPlatformFloor bool
		endPlatformAir   bool
		endPodiumPillar  bool
		endPodiumRing    bool
		endPodiumSupport bool
	)
	handle.ExecWorld(func(tx *world.Tx, e world.Entity) {
		dimAfterFirst = tx.World().Dimension()
		if dimAfterFirst != world.End {
			return
		}
		_, endPlatformFloor = tx.Block(cube.Pos{100, 48, 0}).(block.Obsidian)
		_, endPlatformAir = tx.Block(cube.Pos{100, 49, 0}).(block.Air)
		_, endPodiumPillar = tx.Block(cube.Pos{0, 63, 0}).(block.Bedrock)
		_, endPodiumRing = tx.Block(cube.Pos{3, 63, 0}).(block.Bedrock)
		_, endPodiumSupport = tx.Block(cube.Pos{3, 62, 0}).(block.EndStone)
	})
	if dimAfterFirst != world.End {
		t.Fatalf("expected player in End after entering end portal, got %v", dimAfterFirst)
	}
	if !endPlatformFloor || !endPlatformAir {
		t.Fatal("expected entering the End to ensure the obsidian entry platform and cleared air")
	}
	if !endPodiumPillar || !endPodiumRing || !endPodiumSupport {
		t.Fatal("expected entering the End to ensure the central inactive podium")
	}

	handle.ExecWorld(func(tx *world.Tx, e world.Entity) {
		p := e.(*Player)
		p.portalCooldownUntil = time.Time{}
		pos := cube.PosFromVec3(p.Position())
		tx.SetBlock(pos, block.EndPortal{}, nil)
		p.Tick(tx, 1)
	})

	var dimAfterReturn world.Dimension
	handle.ExecWorld(func(tx *world.Tx, _ world.Entity) {
		dimAfterReturn = tx.World().Dimension()
	})
	if dimAfterReturn != world.Overworld {
		t.Fatalf("expected player back in Overworld after leaving End, got %v", dimAfterReturn)
	}
}

func TestEndPortalActivationWithOutwardFacingFrames(t *testing.T) {
	finalisePlayerBlocksOnce.Do(worldFinaliseBlockRegistry)

	overworld, nether, end := newPortalTestWorlds()
	defer closePlayerTestWorld(t, overworld)
	defer closePlayerTestWorld(t, nether)
	defer closePlayerTestWorld(t, end)

	center := cube.Pos{0, 64, 0}
	missingFrame := center.Add(cube.Pos{0, 0, 2})

	var (
		portalFilled bool
		useErr       error
	)
	<-overworld.Exec(func(tx *world.Tx) {
		buildEndPortalRingFacing(tx, center, missingFrame, true)
		p := tx.AddEntity(world.NewEntity(Type, Config{Name: "ender", Position: center.Vec3()})).(*Player)
		ctx := &item.UseContext{}
		if !(item.EyeOfEnder{}).UseOnBlock(missingFrame, cube.FaceUp, mgl64.Vec3{}, tx, p, ctx) {
			useErr = errString("expected eye of ender use on outward-facing frame ring to succeed")
			return
		}
		portalFilled = isEndPortalFilled(tx, center)
	})
	if useErr != nil {
		t.Fatal(useErr)
	}
	if !portalFilled {
		t.Fatal("expected outward-facing completed frame ring to fill the 3x3 end portal")
	}
}

func TestEndGatewayTeleportsWithinEnd(t *testing.T) {
	finalisePlayerBlocksOnce.Do(worldFinaliseBlockRegistry)

	overworld, nether, end := newPortalTestWorlds()
	defer closePlayerTestWorld(t, overworld)
	defer closePlayerTestWorld(t, nether)
	defer closePlayerTestWorld(t, end)

	gatewayPos := cube.Pos{0, 64, 0}
	handle := world.NewEntity(Type, Config{Name: "gateway-tester", Position: gatewayPos.Vec3Middle()})

	var (
		manualGateway   bool
		detectedGateway bool
		gatewayBlock    bool
		gatewayInsider  bool
	)
	<-end.Exec(func(tx *world.Tx) {
		tx.SetBlock(gatewayPos, block.EndGateway{}, nil)
		_, gatewayBlock = tx.Block(gatewayPos).(block.EndGateway)
		_, gatewayInsider = tx.Block(gatewayPos).(block.EntityInsider)

		p := tx.AddEntity(handle).(*Player)
		(block.EndGateway{}).EntityInside(gatewayPos, tx, p)
		manualGateway = p.inEndGateway
		p.inEndGateway = false
		p.checkEntityInsiders(Type.BBox(p).Translate(p.Position()))
		detectedGateway = p.inEndGateway
		p.Tick(tx, 0)
	})
	if !gatewayBlock {
		t.Fatal("expected placed end gateway block to decode back as block.EndGateway")
	}
	if !gatewayInsider {
		t.Fatal("expected placed end gateway block to satisfy block.EntityInsider")
	}
	if !manualGateway {
		t.Fatal("expected direct end gateway EntityInside call to mark the player")
	}
	if !detectedGateway {
		t.Fatal("expected player collision checks to detect the end gateway block")
	}

	var (
		dimAfter world.Dimension
		posAfter mgl64.Vec3
	)
	handle.ExecWorld(func(tx *world.Tx, e world.Entity) {
		dimAfter = tx.World().Dimension()
		posAfter = e.Position()
	})
	if dimAfter != world.End {
		t.Fatalf("expected player to stay in End after entering end gateway, got %v", dimAfter)
	}
	if got, want := cube.PosFromVec3(posAfter), end.Spawn(); got != want {
		t.Fatalf("expected end gateway to teleport player to %v, got %v", want, got)
	}
}

func newPortalTestWorlds() (*world.World, *world.World, *world.World) {
	var overworld, nether, end *world.World
	overworld = world.Config{
		Dim:      world.Overworld,
		Provider: world.NopProvider{},
		PortalDestination: func(dim world.Dimension) *world.World {
			switch dim {
			case world.Nether:
				return nether
			case world.End:
				return end
			case world.Overworld:
				return overworld
			default:
				return nil
			}
		},
	}.New()
	nether = world.Config{
		Dim:      world.Nether,
		Provider: world.NopProvider{},
		PortalDestination: func(dim world.Dimension) *world.World {
			switch dim {
			case world.Nether, world.Overworld:
				return overworld
			case world.End:
				return end
			default:
				return nil
			}
		},
	}.New()
	end = world.Config{
		Dim:      world.End,
		Provider: world.NopProvider{},
		PortalDestination: func(dim world.Dimension) *world.World {
			switch dim {
			case world.End, world.Overworld:
				return overworld
			case world.Nether:
				return nether
			default:
				return nil
			}
		},
	}.New()
	return overworld, nether, end
}

func buildTestNetherPortal(tx *world.Tx, origin cube.Pos, axis cube.Axis) {
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
			tx.SetBlock(pos, block.Portal{Axis: axis}, nil)
		}
	}
}

func buildEndPortalRing(tx *world.Tx, center, missing cube.Pos) {
	buildEndPortalRingFacing(tx, center, missing, false)
}

func buildEndPortalRingFacing(tx *world.Tx, center, missing cube.Pos, outward bool) {
	for _, spec := range []struct {
		offset cube.Pos
		facing cube.Direction
		eye    bool
	}{
		{offset: cube.Pos{-2, 0, -1}, facing: cube.East, eye: true},
		{offset: cube.Pos{-2, 0, 0}, facing: cube.East, eye: true},
		{offset: cube.Pos{-2, 0, 1}, facing: cube.East, eye: true},
		{offset: cube.Pos{2, 0, -1}, facing: cube.West, eye: true},
		{offset: cube.Pos{2, 0, 0}, facing: cube.West, eye: true},
		{offset: cube.Pos{2, 0, 1}, facing: cube.West, eye: true},
		{offset: cube.Pos{-1, 0, -2}, facing: cube.South, eye: true},
		{offset: cube.Pos{0, 0, -2}, facing: cube.South, eye: true},
		{offset: cube.Pos{1, 0, -2}, facing: cube.South, eye: true},
		{offset: cube.Pos{-1, 0, 2}, facing: cube.North, eye: true},
		{offset: cube.Pos{0, 0, 2}, facing: cube.North, eye: false},
		{offset: cube.Pos{1, 0, 2}, facing: cube.North, eye: true},
	} {
		pos := center.Add(spec.offset)
		if pos == missing {
			spec.eye = false
		}
		if outward {
			spec.facing = spec.facing.Opposite()
		}
		tx.SetBlock(pos, block.EndPortalFrame{Facing: spec.facing, Eye: spec.eye}, nil)
	}
}

func isEndPortalFilled(tx *world.Tx, center cube.Pos) bool {
	for x := -1; x <= 1; x++ {
		for z := -1; z <= 1; z++ {
			if _, ok := tx.Block(center.Add(cube.Pos{x, 0, z})).(block.EndPortal); !ok {
				return false
			}
		}
	}
	return true
}

func lowestPortalPos(positions []cube.Pos) cube.Pos {
	best := positions[0]
	for _, pos := range positions[1:] {
		if pos.Y() < best.Y() {
			best = pos
		}
	}
	return best
}

func closePlayerTestWorld(t *testing.T, w *world.World) {
	t.Helper()
	if err := w.Close(); err != nil {
		t.Fatalf("failed closing world: %v", err)
	}
}

var finalisePlayerBlocksOnce sync.Once

//go:linkname worldFinaliseBlockRegistry github.com/df-mc/dragonfly/server/world.finaliseBlockRegistry
func worldFinaliseBlockRegistry()

type errString string

func (e errString) Error() string { return string(e) }

func errStringf(format string, args ...any) error {
	return errString(fmt.Sprintf(format, args...))
}

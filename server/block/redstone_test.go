package block

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

func runWorld(w *world.World, f func(*world.Tx)) {
	w.Do(f).Wait(context.Background())
}

func TestRedstoneWirePowersBlockBelowButNotAbove(t *testing.T) {
	wire := RedstoneWire{Power: 15}
	pos := cube.Pos{0, 64, 0}

	tests := []struct {
		name string
		face cube.Face
		want int
	}{
		{name: "top", face: cube.FaceUp},
		{name: "bottom", face: cube.FaceDown, want: 15},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if power := wire.RedstonePower(pos, nil, test.face); power != test.want {
				t.Fatalf("power from %s face = %d, want %d", test.face, power, test.want)
			}
		})
	}
}

func TestRedstoneWireVerticalTravel(t *testing.T) {
	tests := []struct {
		name         string
		upperSupport world.Block
		fromHigh     bool
		want         bool
	}{
		{name: "up glowstone", upperSupport: Glowstone{}, want: true},
		{name: "down glowstone", upperSupport: Glowstone{}, fromHigh: true},
		{name: "down glass", upperSupport: Glass{}, fromHigh: true, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := world.Config{Synchronous: true}.New()
			defer w.Close()

			low, high := cube.Pos{1, 64, 0}, cube.Pos{0, 65, 0}
			var neighbours []cube.Pos
			runWorld(w, func(tx *world.Tx) {
				tx.SetBlock(low.Side(cube.FaceDown), Stone{}, nil)
				tx.SetBlock(high.Side(cube.FaceDown), test.upperSupport, nil)
				tx.SetBlock(low, RedstoneWire{}, nil)
				tx.SetBlock(high, RedstoneWire{}, nil)

				from := low
				if test.fromHigh {
					from = high
				}
				neighbours = RedstoneWire{}.RedstoneRelayerNeighbours(from, tx)
			})

			to := high
			if test.fromHigh {
				to = low
			}
			if got := redstoneWireTestContains(neighbours, to); got != test.want {
				t.Fatalf("neighbours = %v, contains %v = %t, want %t", neighbours, to, got, test.want)
			}
		})
	}
}

func TestRedstoneWireBreaksWhenSupportRemoved(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: redstoneBreakDropTestEntityRegistry()}.New()
	defer w.Close()

	wirePos := cube.Pos{0, 64, 0}
	supportPos := wirePos.Side(cube.FaceDown)
	var blockAfter world.Block
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(supportPos, Stone{}, nil)
		tx.SetBlock(wirePos, RedstoneWire{}, nil)
		tx.SetBlock(supportPos, nil, nil)
	})
	w.AdvanceTick()
	runWorld(w, func(tx *world.Tx) {
		blockAfter = tx.Block(wirePos)
	})

	if _, ok := blockAfter.(Air); !ok {
		t.Fatalf("redstone wire after support removal = %T, want Air", blockAfter)
	}
}

func TestRedstoneWireGlowstoneLadderDoesNotOscillateAfterNeighbourBlockUpdate(t *testing.T) {
	for _, test := range []struct {
		name      string
		updatePos cube.Pos
		breaking  bool
	}{
		{name: "place adjacent top dust", updatePos: cube.Pos{0, 67, -1}},
		{name: "place adjacent support", updatePos: cube.Pos{0, 66, -1}},
		{name: "place diagonal top dust", updatePos: cube.Pos{1, 67, -1}},
		{name: "break adjacent top dust", updatePos: cube.Pos{0, 67, -1}, breaking: true},
		{name: "break adjacent support", updatePos: cube.Pos{0, 66, -1}, breaking: true},
		{name: "break diagonal top dust", updatePos: cube.Pos{1, 67, -1}, breaking: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := world.Config{Dim: world.End, Synchronous: true}.New()
			defer w.Close()

			viewer := &redstoneWireTestBlockUpdateViewer{}
			loader := world.NewLoader(2, w, viewer)
			defer func() {
				runWorld(w, func(tx *world.Tx) {
					loader.Close(tx)
				})
			}()

			sourcePos := cube.Pos{2, 64, 0}
			dustPositions := []cube.Pos{
				{1, 64, 0},
				{0, 65, 0},
				{1, 66, 0},
				{0, 67, 0},
			}
			supportPositions := []cube.Pos{
				dustPositions[0].Side(cube.FaceDown),
				{0, 64, 0},
				{1, 65, 0},
				{0, 66, 0},
			}
			topDustPos := dustPositions[len(dustPositions)-1]
			runWorld(w, func(tx *world.Tx) {
				loader.Move(tx, mgl64.Vec3{0, 64, 0})
				loader.Load(tx, 16)
			})
			redstoneWireTestSetBlockAndWait(t, w, sourcePos, RedstoneBlock{})
			for i, supportPos := range supportPositions {
				if i == 0 {
					redstoneWireTestSetBlockAndWait(t, w, supportPos, Stone{})
				} else {
					redstoneWireTestSetBlockAndWait(t, w, supportPos, Glowstone{})
				}
			}
			for _, dustPos := range dustPositions {
				redstoneWireTestSetBlockAndWait(t, w, dustPos, RedstoneWire{})
			}
			if test.breaking {
				redstoneWireTestSetBlockAndWait(t, w, test.updatePos, Stone{})
			}

			redstoneWireTestWaitFor(t, w, func(tx *world.Tx) bool {
				wire, ok := tx.Block(topDustPos).(RedstoneWire)
				return ok && wire.Power > 0
			})
			viewer.reset()

			var initialPower int
			runWorld(w, func(tx *world.Tx) {
				initialPower = tx.Block(topDustPos).(RedstoneWire).Power
				if test.breaking {
					tx.SetBlock(test.updatePos, nil, nil)
				} else {
					tx.SetBlock(test.updatePos, Stone{}, nil)
				}
			})

			lastPower := initialPower
			lastTick := int64(-1)
			powerChanges := 0
			for range 12 {
				lastTick = redstoneWireTestWaitNextTick(t, w, lastTick)
				var power int
				runWorld(w, func(tx *world.Tx) {
					power = tx.Block(topDustPos).(RedstoneWire).Power
				})
				if power != lastPower {
					powerChanges++
					lastPower = power
				}
			}
			if powerChanges != 0 {
				t.Fatalf("top glowstone ladder dust power changed %d times after neighbour update; initial=%d final=%d updatePos=%v breaking=%t", powerChanges, initialPower, lastPower, test.updatePos, test.breaking)
			}
			if updates := viewer.blockUpdateCount(topDustPos); updates != 0 {
				t.Fatalf("top glowstone ladder dust received %d block updates after neighbour update; initial=%d final=%d updatePos=%v breaking=%t", updates, initialPower, lastPower, test.updatePos, test.breaking)
			}
		})
	}
}

func TestRedstoneTorchAttachmentPower(t *testing.T) {
	tests := []struct {
		name  string
		setup func(tx *world.Tx, attachmentPos cube.Pos)
		want  bool
	}{
		{
			name: "ignores non-conductive attachment",
			setup: func(tx *world.Tx, attachmentPos cube.Pos) {
				tx.SetBlock(attachmentPos, Glass{}, nil)
				tx.SetBlock(attachmentPos.Side(cube.FaceNorth), RedstoneBlock{}, nil)
			},
		},
		{
			name: "powered conductive attachment",
			setup: func(tx *world.Tx, attachmentPos cube.Pos) {
				tx.SetBlock(attachmentPos, Stone{}, nil)
				tx.SetBlock(attachmentPos.Side(cube.FaceNorth), RedstoneWire{Power: 15}, nil)
				tx.SetBlock(attachmentPos.Side(cube.FaceNorth).Side(cube.FaceDown), Stone{}, nil)
			},
			want: true,
		},
		{
			name: "redstone block attachment",
			setup: func(tx *world.Tx, attachmentPos cube.Pos) {
				tx.SetBlock(attachmentPos, RedstoneBlock{}, nil)
			},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := world.Config{Synchronous: true}.New()
			defer w.Close()

			torchPos := cube.Pos{1, 64, 0}
			attachmentPos := torchPos.Side(cube.FaceWest)
			var powered bool
			runWorld(w, func(tx *world.Tx) {
				test.setup(tx, attachmentPos)

				torch := RedstoneTorch{Facing: cube.FaceWest, Lit: true}
				powered = torch.attachmentPowered(torchPos, tx)
			})

			if powered != test.want {
				t.Fatalf("attachment powered = %t, want %t", powered, test.want)
			}
		})
	}
}

func TestRedstoneTorchUnknownFacingDoesNotPowerAttachmentFace(t *testing.T) {
	torch := RedstoneTorch{Facing: unknownFace, Lit: true}
	pos := cube.Pos{1, 64, 0}

	tests := []struct {
		name string
		face cube.Face
		want int
	}{
		{name: "attachment", face: cube.FaceDown},
		{name: "top", face: cube.FaceUp, want: 15},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if power := torch.RedstonePower(pos, nil, test.face); power != test.want {
				t.Fatalf("unknown-facing torch power from %s face = %d, want %d", test.face, power, test.want)
			}
		})
	}
}

func TestRedstoneTorchUnknownFacingUsesBlockBelowAsAttachment(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceDown)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var unpoweredAttachment, poweredAttachment, supported bool
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: unknownFace, Lit: true}, nil)

		torch := tx.Block(torchPos).(RedstoneTorch)
		unpoweredAttachment = torch.attachmentPowered(torchPos, tx)

		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
		poweredAttachment = torch.attachmentPowered(torchPos, tx)

		torch.NeighbourUpdateTick(torchPos, attachmentPos, tx)
		_, supported = tx.Block(torchPos).(RedstoneTorch)
	})

	if unpoweredAttachment {
		t.Fatal("unknown-facing torch treated itself as a powered attachment")
	}
	if !poweredAttachment {
		t.Fatal("unknown-facing torch did not use the block below as its powered attachment")
	}
	if !supported {
		t.Fatal("unknown-facing torch broke instead of using the block below as support")
	}
}

func TestRedstoneBlockPowersAdjacentComponentsButNotThroughStone(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	sourcePos := cube.Pos{0, 64, 0}
	adjacentDustPos := sourcePos.Side(cube.FaceEast)
	stonePos := sourcePos.Side(cube.FaceWest)
	farDustPos := stonePos.Side(cube.FaceWest)
	var adjacentPower, farPower int
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(sourcePos, RedstoneBlock{}, nil)
		tx.SetBlock(adjacentDustPos, RedstoneWire{}, nil)
		tx.SetBlock(adjacentDustPos.Side(cube.FaceDown), Stone{}, nil)
		tx.SetBlock(stonePos, Stone{}, nil)
		tx.SetBlock(farDustPos, RedstoneWire{}, nil)
		tx.SetBlock(farDustPos.Side(cube.FaceDown), Stone{}, nil)

		adjacentPower = tx.RedstonePower(adjacentDustPos)
		farPower = tx.RedstonePower(farDustPos)
	})

	if adjacentPower == 0 {
		t.Fatal("redstone block did not power adjacent dust")
	}
	if farPower != 0 {
		t.Fatalf("redstone block powered dust through stone with %d, want 0", farPower)
	}
}

func TestRedstoneBlockDoesNotPowerTorchThroughStone(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	var powered bool
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(attachmentPos.Side(cube.FaceNorth), RedstoneBlock{}, nil)

		torch := RedstoneTorch{Facing: cube.FaceWest, Lit: true}
		powered = torch.attachmentPowered(torchPos, tx)
	})

	if powered {
		t.Fatal("torch attachment was powered through stone by redstone block")
	}
}

func TestLeverStrongPowersAttachedBlockFace(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	leverPos := cube.Pos{1, 64, 0}
	attachedPos := leverPos.Side(cube.FaceWest)
	unattachedPos := leverPos.Side(cube.FaceEast)
	var attachedPower, unattachedPower int
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachedPos, Stone{}, nil)
		tx.SetBlock(unattachedPos, Stone{}, nil)
		tx.SetBlock(leverPos, Lever{Powered: true, Facing: cube.FaceEast}, nil)

		attachedPower = tx.RedstoneStrongPower(attachedPos)
		unattachedPower = tx.RedstoneStrongPower(unattachedPos)
	})

	if attachedPower != 15 {
		t.Fatalf("attached block strong power = %d, want 15", attachedPower)
	}
	if unattachedPower != 0 {
		t.Fatalf("unattached block strong power = %d, want 0", unattachedPower)
	}
}

func TestLeverBreaksWhenSupportRemoved(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: redstoneBreakDropTestEntityRegistry()}.New()
	defer w.Close()

	leverPos := cube.Pos{1, 64, 0}
	supportPos := leverPos.Side(cube.FaceWest)
	var blockAfter world.Block
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(supportPos, Stone{}, nil)
		tx.SetBlock(leverPos, Lever{Facing: cube.FaceEast}, nil)
		tx.SetBlock(supportPos, nil, nil)
	})
	w.AdvanceTick()
	runWorld(w, func(tx *world.Tx) {
		blockAfter = tx.Block(leverPos)
	})

	if _, ok := blockAfter.(Air); !ok {
		t.Fatalf("lever after support removal = %T, want Air", blockAfter)
	}
}

func TestLeverUpdatesConsumerBehindAttachedBlock(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()
	loader := world.NewLoader(1, w, world.NopViewer{})
	runWorld(w, func(tx *world.Tx) {
		loader.Load(tx, 1)
	})
	defer func() {
		runWorld(w, func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	leverPos := cube.Pos{0, 64, 0}
	attachmentPos := leverPos.Side(cube.FaceWest)
	notePos := attachmentPos.Side(cube.FaceWest)
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(notePos, Note{}, nil)
		tx.SetBlock(leverPos, Lever{Facing: cube.FaceEast}, nil)
	})

	redstoneWireTestSetBlockAndWait(t, w, leverPos, Lever{Powered: true, Facing: cube.FaceEast})

	var powered bool
	runWorld(w, func(tx *world.Tx) {
		powered = tx.Block(notePos).(Note).Powered
	})
	if !powered {
		t.Fatal("lever did not update consumer behind its attached block")
	}
}

func TestNoteBlockPlaysOnRedstoneRisingEdgeOnly(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	handler := &redstoneSoundTestHandler{}
	w.Handle(handler)

	leverPos := cube.Pos{0, 64, 0}
	attachmentPos := leverPos.Side(cube.FaceWest)
	notePos := attachmentPos.Side(cube.FaceWest)
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(notePos, Note{}, nil)
		tx.SetBlock(leverPos, Lever{Facing: cube.FaceEast}, nil)
	})

	redstoneWireTestSetBlockAndWait(t, w, leverPos, Lever{Powered: true, Facing: cube.FaceEast})
	redstoneWireTestSetBlockAndWait(t, w, leverPos, Lever{Powered: false, Facing: cube.FaceEast})

	var powered bool
	runWorld(w, func(tx *world.Tx) {
		powered = tx.Block(notePos).(Note).Powered
	})
	if handler.noteSounds != 1 {
		t.Fatalf("note sounds after rising and falling edge = %d, want 1", handler.noteSounds)
	}
	if powered {
		t.Fatal("note block stayed powered after falling edge")
	}
}

func TestTNTDoesNotConductRedstonePower(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	sourcePos := cube.Pos{0, 64, 0}
	tntPos := sourcePos.Side(cube.FaceEast)
	dustPos := tntPos.Side(cube.FaceEast)
	var power int
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(sourcePos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(sourcePos.Side(cube.FaceDown), Stone{}, nil)
		tx.SetBlock(tntPos, TNT{}, nil)
		tx.SetBlock(dustPos, RedstoneWire{}, nil)
		tx.SetBlock(dustPos.Side(cube.FaceDown), Stone{}, nil)

		power = tx.RedstonePower(dustPos)
	})

	if power != 0 {
		t.Fatalf("redstone power conducted through TNT = %d, want 0", power)
	}
}

func TestTNTRedstoneEngineRisingEdgePrimes(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: redstoneTNTTestEntityRegistry()}.New()
	defer w.Close()

	sourcePos := cube.Pos{0, 64, 0}
	tntPos := sourcePos.Side(cube.FaceEast)
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(tntPos, TNT{}, nil)
		tx.SetBlock(sourcePos, RedstoneBlock{}, nil)
	})
	w.AdvanceTick()

	var blockAfter world.Block
	entities := 0
	runWorld(w, func(tx *world.Tx) {
		blockAfter = tx.Block(tntPos)
		for range tx.Entities() {
			entities++
		}
	})
	if _, ok := blockAfter.(Air); !ok {
		t.Fatalf("TNT after engine redstone update = %T, want Air", blockAfter)
	}
	if entities != 1 {
		t.Fatalf("entities after engine redstone update = %d, want 1 primed TNT", entities)
	}
}

func TestTNTRedstonePowerAction(t *testing.T) {
	tests := []struct {
		name         string
		oldPower     int
		newPower     int
		wantAir      bool
		wantEntities int
	}{
		{name: "rising edge primes", newPower: 15, wantAir: true, wantEntities: 1},
		{name: "falling edge ignored", oldPower: 15},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := world.Config{Synchronous: true, Entities: redstoneTNTTestEntityRegistry()}.New()
			defer w.Close()

			pos := cube.Pos{1, 64, 0}
			var blockAfter world.Block
			entities := 0
			runWorld(w, func(tx *world.Tx) {
				tx.SetBlock(pos, TNT{}, nil)

				(TNT{}).RedstonePowerAction(pos, tx, test.oldPower, test.newPower)
				blockAfter = tx.Block(pos)
				for range tx.Entities() {
					entities++
				}
			})

			_, air := blockAfter.(Air)
			if air != test.wantAir {
				t.Fatalf("block after TNT action = %T, air=%t, want air=%t", blockAfter, air, test.wantAir)
			}
			if entities != test.wantEntities {
				t.Fatalf("entities after TNT action = %d, want %d", entities, test.wantEntities)
			}
		})
	}
}

func TestRedstoneTorchBurnsOutAfterRapidSelfTriggeredTurnOffs(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit bool
	var burnedOut, recoverable bool
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		redstoneTorchBurnoutTestToggle(tx, torchPos, inputPos, nil, true)

		lit = tx.Block(torchPos).(RedstoneTorch).Lit
		burnedOut, recoverable = tx.Redstone().Torch(torchPos).BurnoutStatus()
	})

	if lit {
		t.Fatalf("redstone torch stayed lit after rapid turn-offs; burnedOut=%t recoverable=%t", burnedOut, recoverable)
	}
	if !burnedOut {
		t.Fatal("redstone torch turned off without recording burnout")
	}
}

func TestRedstoneTorchExternalTurnOffsDoNotBurnOut(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit bool
	var burnedOut bool
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		redstoneTorchBurnoutTestToggle(tx, torchPos, inputPos, nil, false)

		lit = tx.Block(torchPos).(RedstoneTorch).Lit
		burnedOut, _ = tx.Redstone().Torch(torchPos).BurnoutStatus()
	})

	if lit {
		t.Fatal("redstone torch stayed lit after external power was applied")
	}
	if burnedOut {
		t.Fatal("externally toggled redstone torch recorded burnout")
	}
}

func TestRedstoneTorchScheduledTickReloadsLiveState(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	var lit bool
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest}, &world.SetOpts{DisableRedstoneUpdates: true})

		stale := RedstoneTorch{Facing: cube.FaceWest, Lit: true}
		stale.ScheduledTick(torchPos, tx, nil)
		lit = tx.Block(torchPos).(RedstoneTorch).Lit
	})

	if !lit {
		t.Fatal("redstone torch scheduled tick used stale receiver state instead of live block state")
	}
}

func TestBurnedOutRedstoneTorchRelightsWhenInputIsRemoved(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()
	loader := world.NewLoader(1, w, world.NopViewer{})
	runWorld(w, func(tx *world.Tx) {
		loader.Load(tx, 1)
	})
	defer func() {
		runWorld(w, func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit bool
	var burnedOutTick int64
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		redstoneTorchBurnoutTestToggle(tx, torchPos, inputPos, nil, true)

		burnedOutTick = tx.CurrentTick()
	})
	redstoneWireTestWaitTick(t, w, burnedOutTick)
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(inputPos, nil, nil)
		torch := tx.Block(torchPos).(RedstoneTorch)
		torch.RedstonePowerActionUpdate(torchPos, tx, world.RedstoneUpdate{ChangedNeighbour: inputPos, HasChangedNeighbour: true, ChangedRedstoneRelevant: true})
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		lit = tx.Block(torchPos).(RedstoneTorch).Lit
	})

	if !lit {
		t.Fatal("burned-out redstone torch did not relight after its input was removed")
	}
}

func TestBurnedOutRedstoneTorchRecoversFromExternalScheduledUpdate(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()
	loader := world.NewLoader(1, w, world.NopViewer{})
	runWorld(w, func(tx *world.Tx) {
		loader.Load(tx, 1)
	})
	defer func() {
		runWorld(w, func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	sourcePos := inputPos.Side(cube.FaceNorth)
	var lit bool
	var burnedOutTick int64
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		redstoneTorchBurnoutTestToggle(tx, torchPos, inputPos, nil, true)

		burnedOutTick = tx.CurrentTick()
	})
	redstoneWireTestWaitTick(t, w, burnedOutTick)
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(inputPos, nil, nil)
		torch := tx.Block(torchPos).(RedstoneTorch)
		torch.RedstonePowerActionUpdate(torchPos, tx, world.RedstoneUpdate{
			ChangedNeighbour:        inputPos,
			HasChangedNeighbour:     true,
			ChangedRedstoneRelevant: true,
			Source:                  sourcePos,
			HasSource:               true,
			Cause:                   world.RedstoneUpdateCauseScheduledTick,
		})
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		lit = tx.Block(torchPos).(RedstoneTorch).Lit
	})

	if !lit {
		t.Fatal("burned-out redstone torch did not recover from an external scheduled update")
	}
}

func TestBurnedOutRedstoneTorchDoesNotRecoverFromInputWirePowerDrop(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit bool
	var burnedOut bool
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		redstoneTorchBurnoutTestToggle(tx, torchPos, inputPos, RedstoneWire{}, true)

		tx.SetBlock(inputPos, RedstoneWire{}, nil)
		torch := tx.Block(torchPos).(RedstoneTorch)
		torch.RedstonePowerActionUpdate(torchPos, tx, world.RedstoneUpdate{ChangedNeighbour: inputPos, HasChangedNeighbour: true})
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		lit = tx.Block(torchPos).(RedstoneTorch).Lit
		burnedOut, _ = tx.Redstone().Torch(torchPos).BurnoutStatus()
	})

	if lit || !burnedOut {
		t.Fatalf("burned-out redstone torch recovered from wire power drop; lit=%t burnedOut=%t", lit, burnedOut)
	}
}

func TestBurnedOutRedstoneTorchRecoversFromZeroPositionInputUpdate(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 0, 0}
	attachmentPos := cube.Pos{0, 0, 0}
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit bool
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		redstoneTorchBurnoutTestToggle(tx, torchPos, inputPos, nil, true)

		tx.SetBlock(inputPos, nil, nil)
		torch := tx.Block(torchPos).(RedstoneTorch)
		torch.RedstonePowerActionUpdate(torchPos, tx, world.RedstoneUpdate{ChangedNeighbour: attachmentPos, HasChangedNeighbour: true})
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		lit = tx.Block(torchPos).(RedstoneTorch).Lit
	})

	if !lit {
		t.Fatal("burned-out redstone torch did not recover from a valid zero-position input update")
	}
}

func TestBurnedOutRedstoneTorchDoesNotRelightFromDisconnectedUpdate(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()
	loader := world.NewLoader(1, w, world.NopViewer{})
	runWorld(w, func(tx *world.Tx) {
		loader.Load(tx, 1)
	})
	defer func() {
		runWorld(w, func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit, recoverable bool
	var burnedOutTick int64
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		redstoneTorchBurnoutTestToggle(tx, torchPos, inputPos, nil, true)

		burnedOutTick = tx.CurrentTick()
	})
	redstoneWireTestWaitTick(t, w, burnedOutTick)
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(inputPos, nil, &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true})
		torch := tx.Block(torchPos).(RedstoneTorch)
		torch.RedstonePowerActionUpdate(torchPos, tx, world.RedstoneUpdate{ChangedNeighbour: inputPos.Side(cube.FaceNorth).Side(cube.FaceNorth), HasChangedNeighbour: true})
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		lit = tx.Block(torchPos).(RedstoneTorch).Lit
		_, recoverable = tx.Redstone().Torch(torchPos).BurnoutStatus()
	})

	if lit || recoverable {
		t.Fatalf("burned-out redstone torch recovered from a disconnected update; lit=%t recoverable=%t", lit, recoverable)
	}
}

func TestBurnedOutRedstoneTorchDoesNotSelfRecoverWhenLoopUnpowersInput(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit bool
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		redstoneTorchBurnoutTestToggle(tx, torchPos, inputPos, nil, true)

		tx.SetBlock(inputPos, nil, nil)
		torch := tx.Block(torchPos).(RedstoneTorch)
		torch.RedstonePowerActionUpdate(torchPos, tx, world.RedstoneUpdate{ChangedNeighbour: torchPos, HasChangedNeighbour: true, Cause: world.RedstoneUpdateCauseScheduledTick})
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		lit = tx.Block(torchPos).(RedstoneTorch).Lit
	})

	if lit {
		t.Fatal("burned-out redstone torch self-recovered after its own loop unpowered the input")
	}
}

func redstoneWireTestWaitTick(t *testing.T, w *world.World, tick int64) {
	t.Helper()
	for range 200 {
		w.AdvanceTick()
		var current int64
		runWorld(w, func(tx *world.Tx) {
			current = tx.CurrentTick()
		})
		if current > tick {
			return
		}
	}
	t.Fatalf("world tick did not advance past %d", tick)
}

func redstoneWireTestWaitNextTick(t *testing.T, w *world.World, tick int64) int64 {
	t.Helper()
	for range 200 {
		w.AdvanceTick()
		var current int64
		runWorld(w, func(tx *world.Tx) {
			current = tx.CurrentTick()
		})
		if current > tick {
			return current
		}
	}
	t.Fatalf("world tick did not advance past %d", tick)
	return tick
}

func redstoneWireTestWaitFor(t *testing.T, w *world.World, ready func(tx *world.Tx) bool) {
	t.Helper()
	for range 200 {
		w.AdvanceTick()
		done := false
		runWorld(w, func(tx *world.Tx) {
			done = ready(tx)
		})
		if done {
			return
		}
	}
	t.Fatal("condition was not reached")
}

func redstoneWireTestSetBlockAndWait(t *testing.T, w *world.World, pos cube.Pos, b world.Block) {
	t.Helper()
	var tick int64
	runWorld(w, func(tx *world.Tx) {
		tick = tx.CurrentTick()
		tx.SetBlock(pos, b, nil)
	})
	redstoneWireTestWaitTick(t, w, tick)
}

func redstoneWireTestContains(positions []cube.Pos, pos cube.Pos) bool {
	for _, p := range positions {
		if p == pos {
			return true
		}
	}
	return false
}

type redstoneWireTestBlockUpdateViewer struct {
	world.NopViewer

	mu      sync.Mutex
	updates map[cube.Pos]int
}

func (v *redstoneWireTestBlockUpdateViewer) ViewBlockUpdate(pos cube.Pos, _ world.Block, _ int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.updates == nil {
		v.updates = make(map[cube.Pos]int)
	}
	v.updates[pos]++
}

func (v *redstoneWireTestBlockUpdateViewer) reset() {
	v.mu.Lock()
	defer v.mu.Unlock()
	clear(v.updates)
}

func (v *redstoneWireTestBlockUpdateViewer) blockUpdateCount(pos cube.Pos) int {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.updates[pos]
}

type redstoneSoundTestHandler struct {
	world.NopHandler

	noteSounds int
}

func (h *redstoneSoundTestHandler) HandleSound(_ *world.Context, s world.Sound, _ mgl64.Vec3) {
	if _, ok := s.(sound.Note); ok {
		h.noteSounds++
	}
}

func redstoneTNTTestEntityRegistry() world.EntityRegistry {
	return world.EntityRegistryConfig{
		TNT: func(opts world.EntitySpawnOpts, _ time.Duration) *world.EntityHandle {
			return opts.New(redstoneTNTTestEntityType{}, redstoneTNTTestEntityConfig{})
		},
	}.New([]world.EntityType{redstoneTNTTestEntityType{}})
}

func redstoneBreakDropTestEntityRegistry() world.EntityRegistry {
	return world.EntityRegistryConfig{
		Item: func(opts world.EntitySpawnOpts, _ any) *world.EntityHandle {
			return opts.New(redstoneTNTTestEntityType{}, redstoneTNTTestEntityConfig{})
		},
	}.New([]world.EntityType{redstoneTNTTestEntityType{}})
}

type redstoneTNTTestEntityConfig struct{}

func (redstoneTNTTestEntityConfig) Apply(*world.EntityData) {}

type redstoneTNTTestEntityType struct{}

func (redstoneTNTTestEntityType) Open(_ *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return redstoneTNTTestEntity{handle: handle, data: data}
}

func (redstoneTNTTestEntityType) EncodeEntity() string { return "test:tnt" }
func (redstoneTNTTestEntityType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.49, 0, -0.49, 0.49, 0.98, 0.49)
}
func (redstoneTNTTestEntityType) DecodeNBT(map[string]any, *world.EntityData) {}
func (redstoneTNTTestEntityType) EncodeNBT(*world.EntityData) map[string]any {
	return nil
}

type redstoneTNTTestEntity struct {
	handle *world.EntityHandle
	data   *world.EntityData
}

func (e redstoneTNTTestEntity) Close() error {
	return nil
}

func (e redstoneTNTTestEntity) H() *world.EntityHandle {
	return e.handle
}

func (e redstoneTNTTestEntity) Position() mgl64.Vec3 {
	return e.data.Pos
}

func (e redstoneTNTTestEntity) Rotation() cube.Rotation {
	return e.data.Rot
}
func TestRedstoneTorchLoopBurnsOutThroughWorldScheduler(t *testing.T) {
	w := world.Config{Dim: world.End, Synchronous: true}.New()
	defer w.Close()

	loader := world.NewLoader(2, w, world.NopViewer{})
	defer func() {
		runWorld(w, func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := cube.Pos{0, 64, 0}
	dustPositions := []cube.Pos{
		{1, 66, 0},
		{0, 65, 0},
	}
	runWorld(w, func(tx *world.Tx) {
		loader.Move(tx, mgl64.Vec3{0, 64, 0})
		loader.Load(tx, 16)

		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos.Side(cube.FaceUp), Stone{}, nil)
		for _, pos := range dustPositions {
			tx.SetBlock(pos, RedstoneWire{}, nil)
		}
		tx.SetBlock(torchPos.Side(cube.FaceDown), Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)
	})

	var lit, burnedOut, attachmentPowered bool
	var currentTick, burnedOutTick int64
	dustPower := make(map[cube.Pos]int, len(dustPositions))

	for range 200 {
		w.AdvanceTick()
		redstoneTorchBurnoutTestSnapshot(w, torchPos, dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
		if burnedOut && !lit {
			burnedOutTick = currentTick
			break
		}
	}
	if burnedOutTick == 0 {
		t.Fatalf("redstone torch loop did not burn out through world scheduler; tick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, lit, burnedOut, attachmentPowered, dustPower)
	}
	for currentTick < burnedOutTick+100 {
		w.AdvanceTick()
		redstoneTorchBurnoutTestSnapshot(w, torchPos, dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
		if !burnedOut || lit {
			t.Fatalf("redstone torch loop recovered without an external update; tick=%d burnedOutTick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, burnedOutTick, lit, burnedOut, attachmentPowered, dustPower)
		}
	}
}

func TestBurnedOutRedstoneTorchRelightsWhenLoopWireBreaks(t *testing.T) {
	w := world.Config{Dim: world.End, Synchronous: true}.New()
	defer w.Close()

	loader := world.NewLoader(2, w, world.NopViewer{})
	defer func() {
		runWorld(w, func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := cube.Pos{0, 64, 0}
	loopWirePos := cube.Pos{0, 65, 0}
	dustPositions := []cube.Pos{
		{1, 66, 0},
		loopWirePos,
	}
	runWorld(w, func(tx *world.Tx) {
		loader.Move(tx, mgl64.Vec3{0, 64, 0})
		loader.Load(tx, 16)

		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos.Side(cube.FaceUp), Stone{}, nil)
		for _, pos := range dustPositions {
			tx.SetBlock(pos, RedstoneWire{}, nil)
		}
		tx.SetBlock(torchPos.Side(cube.FaceDown), Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)
	})

	var lit, burnedOut, attachmentPowered bool
	var currentTick int64
	dustPower := make(map[cube.Pos]int, len(dustPositions))

	redstoneTorchBurnoutTestWaitFor(t, w, func() bool {
		redstoneTorchBurnoutTestSnapshot(w, torchPos, dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
		return burnedOut && !lit
	}, func() string {
		return fmt.Sprintf("torch did not burn out before wire break; tick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, lit, burnedOut, attachmentPowered, dustPower)
	})
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(loopWirePos, nil, nil)
	})

	redstoneTorchBurnoutTestWaitFor(t, w, func() bool {
		redstoneTorchBurnoutTestSnapshot(w, torchPos, dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
		return lit && !burnedOut && !attachmentPowered
	}, func() string {
		return fmt.Sprintf("torch did not relight after loop wire broke; tick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, lit, burnedOut, attachmentPowered, dustPower)
	})
}

func TestBurnedOutRedstoneTorchRecoveryUpdates(t *testing.T) {
	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := cube.Pos{0, 64, 0}
	east := torchPos.Side(cube.FaceEast)
	eastTwo := east.Side(cube.FaceEast)
	eastThree := eastTwo.Side(cube.FaceEast)
	unrelatedWire := cube.Pos{10, 64, 10}

	tests := []struct {
		name            string
		dustPositions   []cube.Pos
		waitRecoverable bool
		wantRecover     bool
		setup           func(tx *world.Tx, opts *world.SetOpts)
		update          func(tx *world.Tx)
	}{
		{
			name:            "TestBurnedOutRedstoneTorchRecoversFromAdjacentBlockUpdate",
			waitRecoverable: true,
			wantRecover:     true,
			update: func(tx *world.Tx) {
				tx.SetBlock(east, Stone{}, nil)
			},
		},
		{
			name:            "TestBurnedOutRedstoneTorchRecoversFromVerticalBlockUpdate/up",
			waitRecoverable: true,
			wantRecover:     true,
			update: func(tx *world.Tx) {
				tx.SetBlock(torchPos.Side(cube.FaceUp), Stone{}, nil)
			},
		},
		{
			name:            "TestBurnedOutRedstoneTorchRecoversFromVerticalBlockUpdate/down",
			waitRecoverable: true,
			wantRecover:     true,
			update: func(tx *world.Tx) {
				tx.SetBlock(torchPos.Side(cube.FaceDown), Stone{}, nil)
			},
		},
		{
			name:            "TestBurnedOutRedstoneTorchRecoversFromWireNeighbourUpdate",
			dustPositions:   []cube.Pos{east},
			waitRecoverable: true,
			wantRecover:     true,
			setup: func(tx *world.Tx, opts *world.SetOpts) {
				tx.SetBlock(east.Side(cube.FaceDown), Stone{}, opts)
				tx.SetBlock(east, RedstoneWire{}, opts)
			},
			update: func(tx *world.Tx) {
				tx.SetBlock(east.Side(cube.FaceNorth), Stone{}, nil)
			},
		},
		{
			name:            "TestBurnedOutRedstoneTorchDoesNotRecoverFromRedstoneDustPastAdjacentWire",
			dustPositions:   []cube.Pos{east, eastTwo},
			waitRecoverable: true,
			setup: func(tx *world.Tx, opts *world.SetOpts) {
				tx.SetBlock(east.Side(cube.FaceDown), Stone{}, opts)
				tx.SetBlock(eastTwo.Side(cube.FaceDown), Stone{}, opts)
				tx.SetBlock(east, RedstoneWire{}, opts)
			},
			update: func(tx *world.Tx) {
				tx.SetBlock(eastTwo, RedstoneWire{}, nil)
			},
		},
		{
			name:            "TestBurnedOutRedstoneTorchDoesNotRecoverFromDistantPathWireNeighbourUpdate",
			dustPositions:   []cube.Pos{east, eastTwo, eastThree},
			waitRecoverable: true,
			setup: func(tx *world.Tx, opts *world.SetOpts) {
				for _, pos := range []cube.Pos{east, eastTwo, eastThree} {
					tx.SetBlock(pos.Side(cube.FaceDown), Stone{}, opts)
					tx.SetBlock(pos, RedstoneWire{}, opts)
				}
			},
			update: func(tx *world.Tx) {
				tx.SetBlock(eastThree.Side(cube.FaceNorth), Stone{}, nil)
			},
		},
		{
			name:            "TestBurnedOutRedstoneTorchDoesNotRecoverFromDistantWireUpdate",
			dustPositions:   []cube.Pos{east},
			waitRecoverable: true,
			setup: func(tx *world.Tx, opts *world.SetOpts) {
				tx.SetBlock(east.Side(cube.FaceDown), Stone{}, opts)
				tx.SetBlock(east, RedstoneWire{}, opts)
			},
			update: func(tx *world.Tx) {
				tx.SetBlock(east.Side(cube.FaceNorth).Side(cube.FaceNorth), Stone{}, nil)
			},
		},
		{
			name:          "TestBurnedOutRedstoneTorchDoesNotRecoverFromDistantPathWireBreak",
			dustPositions: []cube.Pos{east, eastTwo, eastThree},
			setup: func(tx *world.Tx, opts *world.SetOpts) {
				for _, pos := range []cube.Pos{east, eastTwo, eastThree} {
					tx.SetBlock(pos.Side(cube.FaceDown), Stone{}, opts)
					tx.SetBlock(pos, RedstoneWire{}, opts)
				}
			},
			update: func(tx *world.Tx) {
				tx.SetBlock(eastThree, nil, nil)
			},
		},
		{
			name: "TestBurnedOutRedstoneTorchDoesNotRecoverFromUnrelatedWireBreak",
			setup: func(tx *world.Tx, opts *world.SetOpts) {
				tx.SetBlock(unrelatedWire.Side(cube.FaceDown), Stone{}, opts)
				tx.SetBlock(unrelatedWire, RedstoneWire{}, opts)
			},
			update: func(tx *world.Tx) {
				tx.SetBlock(unrelatedWire, nil, nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := world.Config{Dim: world.End, Synchronous: true}.New()
			defer w.Close()

			loader := world.NewLoader(2, w, world.NopViewer{})
			defer func() {
				runWorld(w, func(tx *world.Tx) {
					loader.Close(tx)
				})
			}()

			runWorld(w, func(tx *world.Tx) {
				loader.Move(tx, mgl64.Vec3{0, 64, 0})
				loader.Load(tx, 16)

				opts := &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true}
				tx.SetBlock(attachmentPos, Stone{}, opts)
				if test.setup != nil {
					test.setup(tx, opts)
				}
				tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest}, opts)
				redstoneTorchBurnoutTestForceBurnedOut(tx, torchPos)
			})

			var lit, burnedOut, recoverable, attachmentPowered bool
			var currentTick int64
			dustPower := make(map[cube.Pos]int, len(test.dustPositions))

			if test.waitRecoverable {
				redstoneTorchBurnoutTestWaitFor(t, w, func() bool {
					runWorld(w, func(tx *world.Tx) {
						currentTick = tx.CurrentTick()
						burnedOut, recoverable = tx.Redstone().Torch(torchPos).BurnoutStatus()
					})
					return burnedOut && recoverable
				}, func() string {
					return fmt.Sprintf("torch did not become recoverable; tick=%d burnedOut=%t recoverable=%t", currentTick, burnedOut, recoverable)
				})
			} else {
				runWorld(w, func(tx *world.Tx) {
					currentTick = tx.CurrentTick()
					burnedOut, recoverable = tx.Redstone().Torch(torchPos).BurnoutStatus()
				})
				if !burnedOut || recoverable {
					t.Fatalf("torch was not in immediate burnout window; tick=%d burnedOut=%t recoverable=%t", currentTick, burnedOut, recoverable)
				}
			}

			var updateTick int64
			runWorld(w, func(tx *world.Tx) {
				updateTick = tx.CurrentTick()
				test.update(tx)
			})

			if test.wantRecover {
				redstoneTorchBurnoutTestWaitFor(t, w, func() bool {
					redstoneTorchBurnoutTestSnapshot(w, torchPos, test.dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
					return lit && !burnedOut && !attachmentPowered
				}, func() string {
					return fmt.Sprintf("torch did not recover; tick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, lit, burnedOut, attachmentPowered, dustPower)
				})
				return
			}

			for currentTick <= updateTick+10 {
				w.AdvanceTick()
				redstoneTorchBurnoutTestSnapshot(w, torchPos, test.dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
				if lit || !burnedOut {
					t.Fatalf("torch recovered from non-local update; tick=%d updateTick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, updateTick, lit, burnedOut, attachmentPowered, dustPower)
				}
			}
		})
	}
}
func redstoneTorchBurnoutTestSnapshot(w *world.World, torchPos cube.Pos, dustPositions []cube.Pos, currentTick *int64, lit, burnedOut, attachmentPowered *bool, dustPower map[cube.Pos]int) {
	runWorld(w, func(tx *world.Tx) {
		*currentTick = tx.CurrentTick()
		if torch, ok := tx.Block(torchPos).(RedstoneTorch); ok {
			*lit = torch.Lit
			*attachmentPowered = torch.attachmentPowered(torchPos, tx)
		}
		for _, pos := range dustPositions {
			if wire, ok := tx.Block(pos).(RedstoneWire); ok {
				if dustPower != nil {
					dustPower[pos] = wire.Power
				}
			} else {
				if dustPower != nil {
					delete(dustPower, pos)
				}
			}
		}
		*burnedOut, _ = tx.Redstone().Torch(torchPos).BurnoutStatus()
	})
}

func redstoneTorchBurnoutTestForceBurnedOut(tx *world.Tx, pos cube.Pos) {
	for range 10 {
		tx.Redstone().Torch(pos).MarkSelfTriggered()
		tx.Redstone().Torch(pos).RecordTurnOff()
	}
}

func redstoneTorchBurnoutTestToggle(tx *world.Tx, torchPos, inputPos cube.Pos, unpoweredInput world.Block, selfTriggered bool) {
	setPoweredInput := func() {
		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
	}
	tick := func() {
		if selfTriggered {
			tx.Redstone().Torch(torchPos).MarkSelfTriggered()
		}
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
	}
	for range 8 {
		setPoweredInput()
		tick()
		tx.SetBlock(inputPos, unpoweredInput, nil)
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
	}
	setPoweredInput()
	tick()
}

func redstoneTorchBurnoutTestWaitFor(t *testing.T, w *world.World, ready func() bool, fail func() string) {
	t.Helper()
	for range 200 {
		w.AdvanceTick()
		if ready() {
			return
		}
	}
	t.Fatal(fail())
}

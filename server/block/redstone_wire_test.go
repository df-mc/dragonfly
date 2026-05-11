package block

import (
	"sync"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestRedstoneWirePowersBlockBelowButNotAbove(t *testing.T) {
	wire := RedstoneWire{Power: 15}
	pos := cube.Pos{0, 64, 0}

	if power := wire.RedstonePower(pos, nil, cube.FaceUp); power != 0 {
		t.Fatalf("power from wire top face = %d, want 0", power)
	}
	if power := wire.RedstonePower(pos, nil, cube.FaceDown); power != 15 {
		t.Fatalf("power from wire bottom face = %d, want 15", power)
	}
}

func TestRedstoneWireTransmitsUpButNotDownGlowstone(t *testing.T) {
	w := world.Config{}.New()
	defer w.Close()

	low, high := cube.Pos{1, 64, 0}, cube.Pos{0, 65, 0}
	var lowNeighbours, highNeighbours []cube.Pos
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(low.Side(cube.FaceDown), Stone{}, nil)
		tx.SetBlock(high.Side(cube.FaceDown), Glowstone{}, nil)
		tx.SetBlock(low, RedstoneWire{}, nil)
		tx.SetBlock(high, RedstoneWire{}, nil)

		wire := RedstoneWire{}
		lowNeighbours = wire.RedstoneRelayerNeighbours(low, tx)
		highNeighbours = wire.RedstoneRelayerNeighbours(high, tx)
	})

	if !redstoneWireTestContains(lowNeighbours, high) {
		t.Fatalf("lower wire neighbours = %v, want high wire %v", lowNeighbours, high)
	}
	if redstoneWireTestContains(highNeighbours, low) {
		t.Fatalf("upper wire neighbours = %v, did not want lower wire %v", highNeighbours, low)
	}
}

func TestRedstoneWireTransmitsDownGlassInBedrock(t *testing.T) {
	w := world.Config{}.New()
	defer w.Close()

	low, high := cube.Pos{1, 64, 0}, cube.Pos{0, 65, 0}
	var highNeighbours []cube.Pos
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(low.Side(cube.FaceDown), Stone{}, nil)
		tx.SetBlock(high.Side(cube.FaceDown), Glass{}, nil)
		tx.SetBlock(low, RedstoneWire{}, nil)
		tx.SetBlock(high, RedstoneWire{}, nil)

		highNeighbours = RedstoneWire{}.RedstoneRelayerNeighbours(high, tx)
	})

	if !redstoneWireTestContains(highNeighbours, low) {
		t.Fatalf("upper wire neighbours = %v, want lower wire %v", highNeighbours, low)
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
			w := world.Config{Dim: world.End}.New()
			defer w.Close()

			viewer := &redstoneWireTestBlockUpdateViewer{}
			loader := world.NewLoader(2, w, viewer)
			defer func() {
				<-w.Exec(func(tx *world.Tx) {
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
			<-w.Exec(func(tx *world.Tx) {
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
			<-w.Exec(func(tx *world.Tx) {
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
				<-w.Exec(func(tx *world.Tx) {
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

func TestRedstoneTorchIgnoresPowerOnNonConductiveAttachment(t *testing.T) {
	w := world.Config{}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	var powered bool
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Glass{}, nil)
		tx.SetBlock(attachmentPos.Side(cube.FaceNorth), RedstoneBlock{}, nil)

		torch := RedstoneTorch{Facing: cube.FaceWest, Lit: true}
		powered = torch.attachmentPowered(torchPos, tx)
	})

	if powered {
		t.Fatal("torch attachment was powered through non-conductive glass")
	}
}

func TestRedstoneTorchTurnsOffOnPoweredConductiveAttachment(t *testing.T) {
	w := world.Config{}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	var powered bool
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(attachmentPos.Side(cube.FaceNorth), RedstoneWire{Power: 15}, nil)
		tx.SetBlock(attachmentPos.Side(cube.FaceNorth).Side(cube.FaceDown), Stone{}, nil)

		torch := RedstoneTorch{Facing: cube.FaceWest, Lit: true}
		powered = torch.attachmentPowered(torchPos, tx)
	})

	if !powered {
		t.Fatal("torch attachment was not powered through conductive stone")
	}
}

func TestRedstoneTorchTurnsOffOnRedstoneBlockAttachment(t *testing.T) {
	w := world.Config{}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	var powered bool
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, RedstoneBlock{}, nil)

		torch := RedstoneTorch{Facing: cube.FaceWest, Lit: true}
		powered = torch.attachmentPowered(torchPos, tx)
	})

	if !powered {
		t.Fatal("torch attachment was not powered by redstone block")
	}
}

func TestRedstoneTorchUnknownFacingDoesNotPowerAttachmentFace(t *testing.T) {
	torch := RedstoneTorch{Facing: unknownFace, Lit: true}
	pos := cube.Pos{1, 64, 0}

	if power := torch.RedstonePower(pos, nil, cube.FaceDown); power != 0 {
		t.Fatalf("unknown-facing torch power from attachment face = %d, want 0", power)
	}
	if power := torch.RedstonePower(pos, nil, cube.FaceUp); power != 15 {
		t.Fatalf("unknown-facing torch power from top face = %d, want 15", power)
	}
}

func TestRedstoneTorchUnknownFacingUsesBlockBelowAsAttachment(t *testing.T) {
	w := world.Config{}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceDown)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var unpoweredAttachment, poweredAttachment, supported bool
	<-w.Exec(func(tx *world.Tx) {
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
	w := world.Config{}.New()
	defer w.Close()

	sourcePos := cube.Pos{0, 64, 0}
	adjacentDustPos := sourcePos.Side(cube.FaceEast)
	stonePos := sourcePos.Side(cube.FaceWest)
	farDustPos := stonePos.Side(cube.FaceWest)
	var adjacentPower, farPower int
	<-w.Exec(func(tx *world.Tx) {
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
	w := world.Config{}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	var powered bool
	<-w.Exec(func(tx *world.Tx) {
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
	w := world.Config{}.New()
	defer w.Close()

	leverPos := cube.Pos{1, 64, 0}
	attachedPos := leverPos.Side(cube.FaceWest)
	unattachedPos := leverPos.Side(cube.FaceEast)
	var attachedPower, unattachedPower int
	<-w.Exec(func(tx *world.Tx) {
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

func TestLeverUpdatesConsumerBehindAttachedBlock(t *testing.T) {
	w := world.Config{}.New()
	defer w.Close()
	loader := world.NewLoader(1, w, world.NopViewer{})
	<-w.Exec(func(tx *world.Tx) {
		loader.Load(tx, 1)
	})
	defer func() {
		<-w.Exec(func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	leverPos := cube.Pos{0, 64, 0}
	attachmentPos := leverPos.Side(cube.FaceWest)
	notePos := attachmentPos.Side(cube.FaceWest)
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(notePos, Note{}, nil)
		tx.SetBlock(leverPos, Lever{Facing: cube.FaceEast}, nil)
	})

	redstoneWireTestSetBlockAndWait(t, w, leverPos, Lever{Powered: true, Facing: cube.FaceEast})

	var powered bool
	<-w.Exec(func(tx *world.Tx) {
		powered = tx.Block(notePos).(Note).Powered
	})
	if !powered {
		t.Fatal("lever did not update consumer behind its attached block")
	}
}

func TestTNTDoesNotConductRedstonePower(t *testing.T) {
	w := world.Config{}.New()
	defer w.Close()

	sourcePos := cube.Pos{0, 64, 0}
	tntPos := sourcePos.Side(cube.FaceEast)
	dustPos := tntPos.Side(cube.FaceEast)
	var power int
	<-w.Exec(func(tx *world.Tx) {
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

func TestTNTRedstonePowerPrimesOnRisingEdge(t *testing.T) {
	w := world.Config{Entities: redstoneTNTTestEntityRegistry()}.New()
	defer w.Close()

	pos := cube.Pos{1, 64, 0}
	var acted bool
	var blockAfter world.Block
	entities := 0
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(pos, TNT{}, nil)

		acted = (TNT{}).RedstonePowerAction(pos, tx, 0, 15)
		blockAfter = tx.Block(pos)
		for range tx.Entities() {
			entities++
		}
	})

	if !acted {
		t.Fatal("TNT redstone action returned false on rising edge")
	}
	if _, ok := blockAfter.(Air); !ok {
		t.Fatalf("block after powered TNT action = %T, want Air", blockAfter)
	}
	if entities != 1 {
		t.Fatalf("entities after powered TNT action = %d, want 1", entities)
	}
}

func TestTNTRedstonePowerIgnoresFallingEdge(t *testing.T) {
	w := world.Config{Entities: redstoneTNTTestEntityRegistry()}.New()
	defer w.Close()

	pos := cube.Pos{1, 64, 0}
	var acted bool
	var blockAfter world.Block
	entities := 0
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(pos, TNT{}, nil)

		acted = (TNT{}).RedstonePowerAction(pos, tx, 15, 0)
		blockAfter = tx.Block(pos)
		for range tx.Entities() {
			entities++
		}
	})

	if acted {
		t.Fatal("TNT redstone action returned true on falling edge")
	}
	if _, ok := blockAfter.(TNT); !ok {
		t.Fatalf("block after unpowered TNT action = %T, want TNT", blockAfter)
	}
	if entities != 0 {
		t.Fatalf("entities after unpowered TNT action = %d, want 0", entities)
	}
}

func TestRedstoneTorchBurnsOutAfterRapidSelfTriggeredTurnOffs(t *testing.T) {
	w := world.Config{}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit bool
	var burnedOut, recoverable bool
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		for range 8 {
			tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
			tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
			tx.Redstone().Torch(torchPos).MarkSelfTriggered()
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
			tx.SetBlock(inputPos, nil, nil)
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		}
		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
		tx.Redstone().Torch(torchPos).MarkSelfTriggered()
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)

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
	w := world.Config{}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit bool
	var burnedOut bool
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		for range 8 {
			tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
			tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
			tx.SetBlock(inputPos, nil, nil)
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		}
		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)

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
	w := world.Config{}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	var lit bool
	<-w.Exec(func(tx *world.Tx) {
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
	w := world.Config{}.New()
	defer w.Close()
	loader := world.NewLoader(1, w, world.NopViewer{})
	<-w.Exec(func(tx *world.Tx) {
		loader.Load(tx, 1)
	})
	defer func() {
		<-w.Exec(func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit bool
	var burnedOutTick int64
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		for range 8 {
			tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
			tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
			tx.Redstone().Torch(torchPos).MarkSelfTriggered()
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
			tx.SetBlock(inputPos, nil, nil)
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		}
		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
		tx.Redstone().Torch(torchPos).MarkSelfTriggered()
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)

		burnedOutTick = tx.CurrentTick()
	})
	redstoneWireTestWaitTick(t, w, burnedOutTick)
	<-w.Exec(func(tx *world.Tx) {
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

func TestBurnedOutRedstoneTorchDoesNotRecoverFromInputWirePowerDrop(t *testing.T) {
	w := world.Config{}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit bool
	var burnedOut bool
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		for range 8 {
			tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
			tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
			tx.Redstone().Torch(torchPos).MarkSelfTriggered()
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
			tx.SetBlock(inputPos, RedstoneWire{}, nil)
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		}
		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
		tx.Redstone().Torch(torchPos).MarkSelfTriggered()
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)

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
	w := world.Config{}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 0, 0}
	attachmentPos := cube.Pos{0, 0, 0}
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit bool
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		for range 8 {
			tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
			tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
			tx.Redstone().Torch(torchPos).MarkSelfTriggered()
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
			tx.SetBlock(inputPos, nil, nil)
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		}
		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
		tx.Redstone().Torch(torchPos).MarkSelfTriggered()
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)

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
	w := world.Config{}.New()
	defer w.Close()
	loader := world.NewLoader(1, w, world.NopViewer{})
	<-w.Exec(func(tx *world.Tx) {
		loader.Load(tx, 1)
	})
	defer func() {
		<-w.Exec(func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit, recoverable bool
	var burnedOutTick int64
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		for range 8 {
			tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
			tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
			tx.Redstone().Torch(torchPos).MarkSelfTriggered()
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
			tx.SetBlock(inputPos, nil, nil)
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		}
		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
		tx.Redstone().Torch(torchPos).MarkSelfTriggered()
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)

		burnedOutTick = tx.CurrentTick()
	})
	redstoneWireTestWaitTick(t, w, burnedOutTick)
	<-w.Exec(func(tx *world.Tx) {
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

func TestBurnedOutRedstoneTorchDoesNotRecoverOnSameTickInputDrops(t *testing.T) {
	w := world.Config{}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit bool
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		for range 8 {
			tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
			tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
			tx.Redstone().Torch(torchPos).MarkSelfTriggered()
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
			tx.SetBlock(inputPos, nil, nil)
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		}
		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
		tx.Redstone().Torch(torchPos).MarkSelfTriggered()
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)

		tx.SetBlock(inputPos, nil, nil)
		torch := tx.Block(torchPos).(RedstoneTorch)
		torch.RedstonePowerActionUpdate(torchPos, tx, world.RedstoneUpdate{ChangedNeighbour: torchPos, HasChangedNeighbour: true, Cause: world.RedstoneUpdateCauseScheduledTick})
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		lit = tx.Block(torchPos).(RedstoneTorch).Lit
	})

	if lit {
		t.Fatal("burned-out redstone torch recovered on the same tick its input dropped")
	}
}

func TestBurnedOutRedstoneTorchDoesNotSelfRecoverWhenLoopUnpowersInput(t *testing.T) {
	w := world.Config{}.New()
	defer w.Close()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := torchPos.Side(cube.FaceWest)
	inputPos := attachmentPos.Side(cube.FaceNorth)
	var lit bool
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(attachmentPos, Stone{}, nil)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest, Lit: true}, nil)

		for range 8 {
			tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
			tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
			tx.Redstone().Torch(torchPos).MarkSelfTriggered()
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
			tx.SetBlock(inputPos, nil, nil)
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		}
		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
		tx.Redstone().Torch(torchPos).MarkSelfTriggered()
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)

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
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		var current int64
		<-w.Exec(func(tx *world.Tx) {
			current = tx.CurrentTick()
		})
		if current > tick {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("world tick did not advance past %d", tick)
}

func redstoneWireTestWaitNextTick(t *testing.T, w *world.World, tick int64) int64 {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		var current int64
		<-w.Exec(func(tx *world.Tx) {
			current = tx.CurrentTick()
		})
		if current > tick {
			return current
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("world tick did not advance past %d", tick)
	return tick
}

func redstoneWireTestWaitFor(t *testing.T, w *world.World, ready func(tx *world.Tx) bool) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		done := false
		<-w.Exec(func(tx *world.Tx) {
			done = ready(tx)
		})
		if done {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition was not reached")
}

func redstoneWireTestSetBlockAndWait(t *testing.T, w *world.World, pos cube.Pos, b world.Block) {
	t.Helper()
	var tick int64
	<-w.Exec(func(tx *world.Tx) {
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

func redstoneTNTTestEntityRegistry() world.EntityRegistry {
	return world.EntityRegistryConfig{
		TNT: func(opts world.EntitySpawnOpts, _ time.Duration) *world.EntityHandle {
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

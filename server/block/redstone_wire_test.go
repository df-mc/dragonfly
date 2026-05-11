package block

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
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

func TestRedstoneTorchBurnsOutAfterRapidTurnOffs(t *testing.T) {
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
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
			tx.SetBlock(inputPos, nil, nil)
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		}
		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)

		lit = tx.Block(torchPos).(RedstoneTorch).Lit
		burnedOut, recoverable = tx.RedstoneTorchBurnoutStatus(torchPos)
	})

	if lit {
		t.Fatalf("redstone torch stayed lit after rapid turn-offs; burnedOut=%t recoverable=%t", burnedOut, recoverable)
	}
	if !burnedOut {
		t.Fatal("redstone torch turned off without recording burnout")
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
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
			tx.SetBlock(inputPos, nil, nil)
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		}
		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)

		burnedOutTick = tx.CurrentTick()
	})
	redstoneWireTestWaitTick(t, w, burnedOutTick)
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(inputPos, nil, nil)
		torch := tx.Block(torchPos).(RedstoneTorch)
		torch.RedstonePowerActionUpdate(torchPos, tx, world.RedstoneUpdate{ChangedNeighbour: inputPos})
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		lit = tx.Block(torchPos).(RedstoneTorch).Lit
	})

	if !lit {
		t.Fatal("burned-out redstone torch did not relight after its input was removed")
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
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
			tx.SetBlock(inputPos, nil, nil)
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		}
		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)

		tx.SetBlock(inputPos, nil, nil)
		torch := tx.Block(torchPos).(RedstoneTorch)
		torch.RedstonePowerActionUpdate(torchPos, tx, world.RedstoneUpdate{ChangedNeighbour: inputPos})
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
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
			tx.SetBlock(inputPos, nil, nil)
			tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		}
		tx.SetBlock(inputPos, RedstoneWire{Power: 15}, nil)
		tx.SetBlock(inputPos.Side(cube.FaceDown), Stone{}, nil)
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)

		tx.SetBlock(inputPos, nil, nil)
		torch := tx.Block(torchPos).(RedstoneTorch)
		torch.RedstonePowerActionUpdate(torchPos, tx, world.RedstoneUpdate{ChangedNeighbour: torchPos})
		tx.Block(torchPos).(RedstoneTorch).ScheduledTick(torchPos, tx, nil)
		lit = tx.Block(torchPos).(RedstoneTorch).Lit
	})

	if lit {
		t.Fatal("burned-out redstone torch self-recovered after its own loop unpowered the input")
	}
}

func redstoneWireTestWaitTick(t *testing.T, w *world.World, tick int64) {
	t.Helper()
	for deadline := time.Now().Add(time.Second); time.Now().Before(deadline); {
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

func redstoneWireTestContains(positions []cube.Pos, pos cube.Pos) bool {
	for _, p := range positions {
		if p == pos {
			return true
		}
	}
	return false
}

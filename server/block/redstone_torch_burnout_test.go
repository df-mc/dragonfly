package block

import (
	"fmt"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestRedstoneTorchLoopBurnsOutThroughWorldScheduler(t *testing.T) {
	w := world.Config{Dim: world.End}.New()
	defer w.Close()

	loader := world.NewLoader(2, w, world.NopViewer{})
	defer func() {
		<-w.Exec(func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := cube.Pos{0, 64, 0}
	dustPositions := []cube.Pos{
		{1, 66, 0},
		{0, 65, 0},
	}
	<-w.Exec(func(tx *world.Tx) {
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

	ticker := time.NewTicker(time.Second / 40)
	defer ticker.Stop()

	var lit, burnedOut, attachmentPowered bool
	var currentTick, burnedOutTick int64
	dustPower := make(map[cube.Pos]int, len(dustPositions))

	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		<-ticker.C
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
		<-ticker.C
		redstoneTorchBurnoutTestSnapshot(w, torchPos, dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
		if !burnedOut || lit {
			t.Fatalf("redstone torch loop recovered without an external update; tick=%d burnedOutTick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, burnedOutTick, lit, burnedOut, attachmentPowered, dustPower)
		}
	}
}

func TestBurnedOutRedstoneTorchRelightsWhenLoopWireBreaks(t *testing.T) {
	w := world.Config{Dim: world.End}.New()
	defer w.Close()

	loader := world.NewLoader(2, w, world.NopViewer{})
	defer func() {
		<-w.Exec(func(tx *world.Tx) {
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
	<-w.Exec(func(tx *world.Tx) {
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

	ticker := time.NewTicker(time.Second / 40)
	defer ticker.Stop()

	var lit, burnedOut, attachmentPowered bool
	var currentTick int64
	dustPower := make(map[cube.Pos]int, len(dustPositions))

	redstoneTorchBurnoutTestWaitFor(t, ticker, w, func() bool {
		redstoneTorchBurnoutTestSnapshot(w, torchPos, dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
		return burnedOut && !lit
	}, func() string {
		return fmt.Sprintf("torch did not burn out before wire break; tick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, lit, burnedOut, attachmentPowered, dustPower)
	})

	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(loopWirePos, nil, nil)
	})

	redstoneTorchBurnoutTestWaitFor(t, ticker, w, func() bool {
		redstoneTorchBurnoutTestSnapshot(w, torchPos, dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
		return lit && !burnedOut && !attachmentPowered
	}, func() string {
		return fmt.Sprintf("torch did not relight after loop wire broke; tick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, lit, burnedOut, attachmentPowered, dustPower)
	})
}

func TestBurnedOutRedstoneTorchDoesNotRecoverFromDistantPathWireBreak(t *testing.T) {
	w := world.Config{Dim: world.End}.New()
	defer w.Close()

	loader := world.NewLoader(2, w, world.NopViewer{})
	defer func() {
		<-w.Exec(func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := cube.Pos{0, 64, 0}
	dustPositions := []cube.Pos{
		torchPos.Side(cube.FaceEast),
		torchPos.Side(cube.FaceEast).Side(cube.FaceEast),
		torchPos.Side(cube.FaceEast).Side(cube.FaceEast).Side(cube.FaceEast),
	}
	breakPos := dustPositions[len(dustPositions)-1]
	<-w.Exec(func(tx *world.Tx) {
		loader.Move(tx, mgl64.Vec3{0, 64, 0})
		loader.Load(tx, 16)

		setupOpts := &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true}
		tx.SetBlock(attachmentPos, Stone{}, setupOpts)
		for _, pos := range dustPositions {
			tx.SetBlock(pos.Side(cube.FaceDown), Stone{}, setupOpts)
			tx.SetBlock(pos, RedstoneWire{}, setupOpts)
		}
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest}, setupOpts)
		redstoneTorchBurnoutTestForceBurnedOut(tx, torchPos)
	})

	ticker := time.NewTicker(time.Second / 40)
	defer ticker.Stop()

	var lit, burnedOut, recoverable, attachmentPowered bool
	var currentTick int64
	dustPower := make(map[cube.Pos]int, len(dustPositions))
	<-w.Exec(func(tx *world.Tx) {
		currentTick = tx.CurrentTick()
		burnedOut, recoverable = tx.Redstone().Torch(torchPos).BurnoutStatus()
	})
	if !burnedOut || recoverable {
		t.Fatalf("torch was not in immediate burnout window before wire break; tick=%d burnedOut=%t recoverable=%t", currentTick, burnedOut, recoverable)
	}

	var updateTick int64
	<-w.Exec(func(tx *world.Tx) {
		updateTick = tx.CurrentTick()
		tx.SetBlock(breakPos, nil, nil)
	})

	for currentTick <= updateTick+10 {
		<-ticker.C
		redstoneTorchBurnoutTestSnapshot(w, torchPos, dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
		if lit || !burnedOut {
			t.Fatalf("torch relit from distant path wire break; tick=%d updateTick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, updateTick, lit, burnedOut, attachmentPowered, dustPower)
		}
	}
}

func TestBurnedOutRedstoneTorchDoesNotRecoverFromUnrelatedWireBreak(t *testing.T) {
	w := world.Config{Dim: world.End}.New()
	defer w.Close()

	loader := world.NewLoader(2, w, world.NopViewer{})
	defer func() {
		<-w.Exec(func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := cube.Pos{0, 64, 0}
	wirePos := cube.Pos{10, 64, 10}
	<-w.Exec(func(tx *world.Tx) {
		loader.Move(tx, mgl64.Vec3{0, 64, 0})
		loader.Load(tx, 16)

		setupOpts := &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true}
		tx.SetBlock(attachmentPos, Stone{}, setupOpts)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest}, setupOpts)
		tx.SetBlock(wirePos.Side(cube.FaceDown), Stone{}, setupOpts)
		tx.SetBlock(wirePos, RedstoneWire{}, setupOpts)
		redstoneTorchBurnoutTestForceBurnedOut(tx, torchPos)
	})

	ticker := time.NewTicker(time.Second / 40)
	defer ticker.Stop()

	var lit, burnedOut, attachmentPowered bool
	var currentTick int64
	var updateTick int64
	<-w.Exec(func(tx *world.Tx) {
		updateTick = tx.CurrentTick()
		tx.SetBlock(wirePos, nil, nil)
	})

	for currentTick <= updateTick+10 {
		<-ticker.C
		redstoneTorchBurnoutTestSnapshot(w, torchPos, nil, &currentTick, &lit, &burnedOut, &attachmentPowered, nil)
		if lit || !burnedOut {
			t.Fatalf("torch relit from unrelated wire break; tick=%d updateTick=%d lit=%t burnedOut=%t attachmentPowered=%t", currentTick, updateTick, lit, burnedOut, attachmentPowered)
		}
	}
}

func TestBurnedOutRedstoneTorchRecoversFromAdjacentBlockUpdate(t *testing.T) {
	w := world.Config{Dim: world.End}.New()
	defer w.Close()

	loader := world.NewLoader(2, w, world.NopViewer{})
	defer func() {
		<-w.Exec(func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := cube.Pos{0, 64, 0}
	updatePos := torchPos.Side(cube.FaceEast)
	<-w.Exec(func(tx *world.Tx) {
		loader.Move(tx, mgl64.Vec3{0, 64, 0})
		loader.Load(tx, 16)

		setupOpts := &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true}
		tx.SetBlock(attachmentPos, Stone{}, setupOpts)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest}, &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true})
		redstoneTorchBurnoutTestForceBurnedOut(tx, torchPos)
	})

	ticker := time.NewTicker(time.Second / 40)
	defer ticker.Stop()

	var lit, burnedOut, recoverable, attachmentPowered bool
	var currentTick int64

	redstoneTorchBurnoutTestWaitFor(t, ticker, w, func() bool {
		<-w.Exec(func(tx *world.Tx) {
			currentTick = tx.CurrentTick()
			burnedOut, recoverable = tx.Redstone().Torch(torchPos).BurnoutStatus()
		})
		return burnedOut && recoverable
	}, func() string {
		return fmt.Sprintf("torch did not become recoverable; tick=%d burnedOut=%t recoverable=%t", currentTick, burnedOut, recoverable)
	})

	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(updatePos, Stone{}, nil)
	})

	redstoneTorchBurnoutTestWaitFor(t, ticker, w, func() bool {
		redstoneTorchBurnoutTestSnapshot(w, torchPos, nil, &currentTick, &lit, &burnedOut, &attachmentPowered, nil)
		return lit && !burnedOut && !attachmentPowered
	}, func() string {
		return fmt.Sprintf("torch did not relight after adjacent block update; tick=%d lit=%t burnedOut=%t attachmentPowered=%t", currentTick, lit, burnedOut, attachmentPowered)
	})
}

func TestBurnedOutRedstoneTorchRecoversFromVerticalBlockUpdate(t *testing.T) {
	for _, face := range []cube.Face{cube.FaceUp, cube.FaceDown} {
		t.Run(face.String(), func(t *testing.T) {
			w := world.Config{Dim: world.End}.New()
			defer w.Close()

			loader := world.NewLoader(2, w, world.NopViewer{})
			defer func() {
				<-w.Exec(func(tx *world.Tx) {
					loader.Close(tx)
				})
			}()

			torchPos := cube.Pos{1, 64, 0}
			attachmentPos := cube.Pos{0, 64, 0}
			updatePos := torchPos.Side(face)
			<-w.Exec(func(tx *world.Tx) {
				loader.Move(tx, mgl64.Vec3{0, 64, 0})
				loader.Load(tx, 16)

				setupOpts := &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true}
				tx.SetBlock(attachmentPos, Stone{}, setupOpts)
				tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest}, setupOpts)
				redstoneTorchBurnoutTestForceBurnedOut(tx, torchPos)
			})

			ticker := time.NewTicker(time.Second / 40)
			defer ticker.Stop()

			var lit, burnedOut, recoverable, attachmentPowered bool
			var currentTick int64

			redstoneTorchBurnoutTestWaitFor(t, ticker, w, func() bool {
				<-w.Exec(func(tx *world.Tx) {
					currentTick = tx.CurrentTick()
					burnedOut, recoverable = tx.Redstone().Torch(torchPos).BurnoutStatus()
				})
				return burnedOut && recoverable
			}, func() string {
				return fmt.Sprintf("torch did not become recoverable; tick=%d burnedOut=%t recoverable=%t", currentTick, burnedOut, recoverable)
			})

			<-w.Exec(func(tx *world.Tx) {
				tx.SetBlock(updatePos, Stone{}, nil)
			})

			redstoneTorchBurnoutTestWaitFor(t, ticker, w, func() bool {
				redstoneTorchBurnoutTestSnapshot(w, torchPos, nil, &currentTick, &lit, &burnedOut, &attachmentPowered, nil)
				return lit && !burnedOut && !attachmentPowered
			}, func() string {
				return fmt.Sprintf("torch did not relight after vertical block update; tick=%d face=%s lit=%t burnedOut=%t attachmentPowered=%t", currentTick, face, lit, burnedOut, attachmentPowered)
			})
		})
	}
}

func TestBurnedOutRedstoneTorchRecoversFromWireNeighbourUpdate(t *testing.T) {
	w := world.Config{Dim: world.End}.New()
	defer w.Close()

	loader := world.NewLoader(2, w, world.NopViewer{})
	defer func() {
		<-w.Exec(func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := cube.Pos{0, 64, 0}
	wirePos := torchPos.Side(cube.FaceEast)
	updatePos := wirePos.Side(cube.FaceNorth)
	dustPositions := []cube.Pos{wirePos}
	<-w.Exec(func(tx *world.Tx) {
		loader.Move(tx, mgl64.Vec3{0, 64, 0})
		loader.Load(tx, 16)

		setupOpts := &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true}
		tx.SetBlock(attachmentPos, Stone{}, setupOpts)
		tx.SetBlock(wirePos.Side(cube.FaceDown), Stone{}, setupOpts)
		tx.SetBlock(wirePos, RedstoneWire{}, setupOpts)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest}, &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true})
		redstoneTorchBurnoutTestForceBurnedOut(tx, torchPos)
	})

	ticker := time.NewTicker(time.Second / 40)
	defer ticker.Stop()

	var lit, burnedOut, recoverable, attachmentPowered bool
	var currentTick int64
	dustPower := make(map[cube.Pos]int, len(dustPositions))

	redstoneTorchBurnoutTestWaitFor(t, ticker, w, func() bool {
		<-w.Exec(func(tx *world.Tx) {
			currentTick = tx.CurrentTick()
			burnedOut, recoverable = tx.Redstone().Torch(torchPos).BurnoutStatus()
		})
		return burnedOut && recoverable
	}, func() string {
		return fmt.Sprintf("torch did not become recoverable; tick=%d burnedOut=%t recoverable=%t", currentTick, burnedOut, recoverable)
	})

	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(updatePos, Stone{}, nil)
	})

	redstoneTorchBurnoutTestWaitFor(t, ticker, w, func() bool {
		redstoneTorchBurnoutTestSnapshot(w, torchPos, dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
		return lit && !burnedOut && !attachmentPowered
	}, func() string {
		return fmt.Sprintf("torch did not relight after wire-neighbour block update; tick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, lit, burnedOut, attachmentPowered, dustPower)
	})
}

func TestBurnedOutRedstoneTorchDoesNotRecoverFromRedstoneDustPastAdjacentWire(t *testing.T) {
	w := world.Config{Dim: world.End}.New()
	defer w.Close()

	loader := world.NewLoader(2, w, world.NopViewer{})
	defer func() {
		<-w.Exec(func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := cube.Pos{0, 64, 0}
	wirePos := torchPos.Side(cube.FaceEast)
	updatePos := wirePos.Side(cube.FaceEast)
	dustPositions := []cube.Pos{wirePos, updatePos}
	<-w.Exec(func(tx *world.Tx) {
		loader.Move(tx, mgl64.Vec3{0, 64, 0})
		loader.Load(tx, 16)

		setupOpts := &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true}
		tx.SetBlock(attachmentPos, Stone{}, setupOpts)
		tx.SetBlock(wirePos.Side(cube.FaceDown), Stone{}, setupOpts)
		tx.SetBlock(updatePos.Side(cube.FaceDown), Stone{}, setupOpts)
		tx.SetBlock(wirePos, RedstoneWire{}, setupOpts)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest}, setupOpts)
		redstoneTorchBurnoutTestForceBurnedOut(tx, torchPos)
	})

	ticker := time.NewTicker(time.Second / 40)
	defer ticker.Stop()

	var lit, burnedOut, recoverable, attachmentPowered bool
	var currentTick int64
	dustPower := make(map[cube.Pos]int, len(dustPositions))

	redstoneTorchBurnoutTestWaitFor(t, ticker, w, func() bool {
		<-w.Exec(func(tx *world.Tx) {
			currentTick = tx.CurrentTick()
			burnedOut, recoverable = tx.Redstone().Torch(torchPos).BurnoutStatus()
		})
		return burnedOut && recoverable
	}, func() string {
		return fmt.Sprintf("torch did not become recoverable; tick=%d burnedOut=%t recoverable=%t", currentTick, burnedOut, recoverable)
	})

	var updateTick int64
	<-w.Exec(func(tx *world.Tx) {
		updateTick = tx.CurrentTick()
		tx.SetBlock(updatePos, RedstoneWire{}, nil)
	})

	for currentTick <= updateTick+10 {
		<-ticker.C
		redstoneTorchBurnoutTestSnapshot(w, torchPos, dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
		if lit || !burnedOut {
			t.Fatalf("torch relit from redstone dust placed past adjacent wire; tick=%d updateTick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, updateTick, lit, burnedOut, attachmentPowered, dustPower)
		}
	}
}

func TestBurnedOutRedstoneTorchDoesNotRecoverFromDistantPathWireNeighbourUpdate(t *testing.T) {
	w := world.Config{Dim: world.End}.New()
	defer w.Close()

	loader := world.NewLoader(2, w, world.NopViewer{})
	defer func() {
		<-w.Exec(func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := cube.Pos{0, 64, 0}
	dustPositions := []cube.Pos{
		torchPos.Side(cube.FaceEast),
		torchPos.Side(cube.FaceEast).Side(cube.FaceEast),
		torchPos.Side(cube.FaceEast).Side(cube.FaceEast).Side(cube.FaceEast),
	}
	updatePos := dustPositions[len(dustPositions)-1].Side(cube.FaceNorth)
	<-w.Exec(func(tx *world.Tx) {
		loader.Move(tx, mgl64.Vec3{0, 64, 0})
		loader.Load(tx, 16)

		setupOpts := &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true}
		tx.SetBlock(attachmentPos, Stone{}, setupOpts)
		for _, pos := range dustPositions {
			tx.SetBlock(pos.Side(cube.FaceDown), Stone{}, setupOpts)
			tx.SetBlock(pos, RedstoneWire{}, setupOpts)
		}
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest}, &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true})
		redstoneTorchBurnoutTestForceBurnedOut(tx, torchPos)
	})

	ticker := time.NewTicker(time.Second / 40)
	defer ticker.Stop()

	var lit, burnedOut, recoverable, attachmentPowered bool
	var currentTick int64
	dustPower := make(map[cube.Pos]int, len(dustPositions))

	redstoneTorchBurnoutTestWaitFor(t, ticker, w, func() bool {
		<-w.Exec(func(tx *world.Tx) {
			currentTick = tx.CurrentTick()
			burnedOut, recoverable = tx.Redstone().Torch(torchPos).BurnoutStatus()
		})
		return burnedOut && recoverable
	}, func() string {
		return fmt.Sprintf("torch did not become recoverable; tick=%d burnedOut=%t recoverable=%t", currentTick, burnedOut, recoverable)
	})

	var updateTick int64
	<-w.Exec(func(tx *world.Tx) {
		updateTick = tx.CurrentTick()
		tx.SetBlock(updatePos, Stone{}, nil)
	})

	for currentTick <= updateTick+10 {
		<-ticker.C
		redstoneTorchBurnoutTestSnapshot(w, torchPos, dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
		if lit || !burnedOut {
			t.Fatalf("torch relit from block update beside distant path wire; tick=%d updateTick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, updateTick, lit, burnedOut, attachmentPowered, dustPower)
		}
	}
}

func TestBurnedOutRedstoneTorchDoesNotRecoverFromDistantWireUpdate(t *testing.T) {
	w := world.Config{Dim: world.End}.New()
	defer w.Close()

	loader := world.NewLoader(2, w, world.NopViewer{})
	defer func() {
		<-w.Exec(func(tx *world.Tx) {
			loader.Close(tx)
		})
	}()

	torchPos := cube.Pos{1, 64, 0}
	attachmentPos := cube.Pos{0, 64, 0}
	wirePos := torchPos.Side(cube.FaceEast)
	updatePos := wirePos.Side(cube.FaceNorth).Side(cube.FaceNorth)
	dustPositions := []cube.Pos{wirePos}
	<-w.Exec(func(tx *world.Tx) {
		loader.Move(tx, mgl64.Vec3{0, 64, 0})
		loader.Load(tx, 16)

		setupOpts := &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true}
		tx.SetBlock(attachmentPos, Stone{}, setupOpts)
		tx.SetBlock(wirePos.Side(cube.FaceDown), Stone{}, setupOpts)
		tx.SetBlock(wirePos, RedstoneWire{}, setupOpts)
		tx.SetBlock(torchPos, RedstoneTorch{Facing: cube.FaceWest}, &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true})
		redstoneTorchBurnoutTestForceBurnedOut(tx, torchPos)
	})

	ticker := time.NewTicker(time.Second / 40)
	defer ticker.Stop()

	var lit, burnedOut, recoverable, attachmentPowered bool
	var currentTick int64
	dustPower := make(map[cube.Pos]int, len(dustPositions))

	redstoneTorchBurnoutTestWaitFor(t, ticker, w, func() bool {
		<-w.Exec(func(tx *world.Tx) {
			currentTick = tx.CurrentTick()
			burnedOut, recoverable = tx.Redstone().Torch(torchPos).BurnoutStatus()
		})
		return burnedOut && recoverable
	}, func() string {
		return fmt.Sprintf("torch did not become recoverable; tick=%d burnedOut=%t recoverable=%t", currentTick, burnedOut, recoverable)
	})

	var updateTick int64
	<-w.Exec(func(tx *world.Tx) {
		updateTick = tx.CurrentTick()
		tx.SetBlock(updatePos, Stone{}, nil)
	})

	for currentTick <= updateTick+10 {
		<-ticker.C
		redstoneTorchBurnoutTestSnapshot(w, torchPos, dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
		if lit || !burnedOut {
			t.Fatalf("torch relit from block update one block away from wire; tick=%d updateTick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, updateTick, lit, burnedOut, attachmentPowered, dustPower)
		}
	}
}

func redstoneTorchBurnoutTestSnapshot(w *world.World, torchPos cube.Pos, dustPositions []cube.Pos, currentTick *int64, lit, burnedOut, attachmentPowered *bool, dustPower map[cube.Pos]int) {
	<-w.Exec(func(tx *world.Tx) {
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

func redstoneTorchBurnoutTestWaitFor(t *testing.T, ticker *time.Ticker, w *world.World, ready func() bool, fail func() string) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		<-ticker.C
		if ready() {
			return
		}
	}
	t.Fatal(fail())
}

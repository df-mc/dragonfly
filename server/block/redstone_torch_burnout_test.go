package block

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestRedstoneTorchLoopBurnsOutThroughWorldScheduler(t *testing.T) {
	w := world.Config{}.New()
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

	for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); {
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
	for currentTick < burnedOutTick+10 {
		<-ticker.C
		redstoneTorchBurnoutTestSnapshot(w, torchPos, dustPositions, &currentTick, &lit, &burnedOut, &attachmentPowered, dustPower)
	}
	if !burnedOut || lit {
		t.Fatalf("redstone torch loop did not remain burned out through world scheduler; tick=%d lit=%t burnedOut=%t attachmentPowered=%t dust=%v", currentTick, lit, burnedOut, attachmentPowered, dustPower)
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
				dustPower[pos] = wire.Power
			} else {
				delete(dustPower, pos)
			}
		}
		*burnedOut, _ = tx.RedstoneTorchBurnoutStatus(torchPos)
	})
}

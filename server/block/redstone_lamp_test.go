package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

func TestRedstoneLampCancelledPowerLoss(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()
	w.Handle(lampPowerLossHandler{})
	pos := cube.Pos{0, 64, 0}
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(pos, RedstoneLamp{}, nil)
		tx.SetBlock(pos.Side(cube.FaceEast), RedstoneBlock{}, nil)
	})
	w.AdvanceTick()
	runWorld(w, func(tx *world.Tx) {
		if !tx.Block(pos).(RedstoneLamp).Lit {
			t.Fatal("powered lamp did not light")
		}
		tx.SetBlock(pos.Side(cube.FaceEast), Air{}, nil)
	})
	for range 6 {
		w.AdvanceTick()
	}
	runWorld(w, func(tx *world.Tx) {
		if !tx.Block(pos).(RedstoneLamp).Lit {
			t.Error("cancelled power loss still turned the lamp off")
		}
	})
}

func TestRedstoneLampOffDelayRestartsAfterPowerReturns(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()
	pos := cube.Pos{0, 64, 0}
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(pos, RedstoneLamp{}, nil)
		tx.SetBlock(pos.Side(cube.FaceEast), RedstoneBlock{}, nil)
	})
	w.AdvanceTick()
	runWorld(w, func(tx *world.Tx) { tx.SetBlock(pos.Side(cube.FaceEast), Air{}, nil) })
	w.AdvanceTick()
	w.AdvanceTick()
	runWorld(w, func(tx *world.Tx) { tx.SetBlock(pos.Side(cube.FaceEast), RedstoneBlock{}, nil) })
	w.AdvanceTick()
	runWorld(w, func(tx *world.Tx) { tx.SetBlock(pos.Side(cube.FaceEast), Air{}, nil) })
	w.AdvanceTick()
	for tick := 1; tick <= 4; tick++ {
		w.AdvanceTick()
		runWorld(w, func(tx *world.Tx) {
			if lit := tx.Block(pos).(RedstoneLamp).Lit; lit != (tick < 4) {
				t.Errorf("tick %d after power loss: lit=%v, want %v", tick, lit, tick < 4)
			}
		})
	}
}

func TestRedstoneLampInitiallyLitWithoutPower(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()
	pos := cube.Pos{0, 64, 0}
	runWorld(w, func(tx *world.Tx) { tx.SetBlock(pos, RedstoneLamp{Lit: true}, nil) })
	for range 5 {
		w.AdvanceTick()
	}
	runWorld(w, func(tx *world.Tx) {
		if tx.Block(pos).(RedstoneLamp).Lit {
			t.Error("initially lit lamp remained lit without power")
		}
	})
}

type lampPowerLossHandler struct{ world.NopHandler }

// HandleRedstoneUpdate prevents lamps from accepting a loss of power.
func (lampPowerLossHandler) HandleRedstoneUpdate(ctx *world.Context, update world.RedstoneUpdate) {
	if _, ok := update.Before.(RedstoneLamp); ok && update.NewPower == 0 {
		ctx.Cancel()
	}
}

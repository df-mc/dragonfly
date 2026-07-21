package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

func TestDispenserStatesRegistered(t *testing.T) {
	w := world.New()
	defer func() { _ = w.Close() }()

	for _, facing := range cube.Faces() {
		for _, triggered := range []uint8{0, 1} {
			properties := map[string]any{
				"facing_direction": int32(facing),
				"triggered_bit":    triggered,
			}
			b, ok := world.BlockByName("minecraft:dispenser", properties)
			if !ok {
				t.Errorf("dispenser state facing=%v triggered=%d is not registered", facing, triggered)
				continue
			}
			if _, ok := b.(Container); !ok {
				t.Errorf("dispenser state facing=%v triggered=%d resolved to %T, which is not a container", facing, triggered, b)
			}
		}
	}
}

func TestDispenserUsesQuasiConnectivity(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer func() { _ = w.Close() }()

	dispenserPos := cube.Pos{0, 0, 0}
	powerPos := cube.Pos{1, 1, 0}
	runWorld(w, func(tx *world.Tx) {
		d := NewDispenser()
		d.Facing = cube.FaceNorth
		tx.SetBlock(dispenserPos, d, nil)
		tx.SetBlock(powerPos, RedstoneBlock{}, nil)
	})
	w.AdvanceTick()
	runWorld(w, func(tx *world.Tx) {
		got := tx.Block(dispenserPos).(Dispenser)
		if !got.Triggered {
			t.Fatal("expected dispenser to be triggered by power adjacent to the block above it")
		}

		tx.SetBlock(powerPos, Air{}, nil)
	})
	w.AdvanceTick()
	runWorld(w, func(tx *world.Tx) {
		if got := tx.Block(dispenserPos).(Dispenser); got.Triggered {
			t.Fatal("expected dispenser to reset after quasi-connected power was removed")
		}
	})
}

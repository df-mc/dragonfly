package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

func TestHopperInsertUsesDestinationPosition(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer func() { _ = w.Close() }()

	hopperPos := cube.Pos{0, 1, 0}
	destinationPos := cube.Pos{0, 0, 0}
	pairPos := cube.Pos{1, 0, 0}

	var inserted bool
	runWorld(w, func(tx *world.Tx) {
		destination := NewChest()
		destination.Facing = cube.North
		destination.paired = true
		destination.pairX, destination.pairZ = pairPos[0], pairPos[2]

		pair := NewChest()
		pair.Facing = cube.North
		pair.paired = true
		pair.pairX, pair.pairZ = destinationPos[0], destinationPos[2]

		hopper := NewHopper()
		hopper.Facing = cube.FaceDown
		_ = hopper.inventory.SetItem(0, item.NewStack(item.Stick{}, 1))

		tx.SetBlock(destinationPos, destination, nil)
		tx.SetBlock(pairPos, pair, nil)
		tx.SetBlock(hopperPos, hopper, nil)

		inserted = hopper.insertItem(hopperPos, tx)
		if _, ok := tx.Block(hopperPos).(Hopper); !ok {
			t.Errorf("hopper block replaced with %T while resolving destination inventory", tx.Block(hopperPos))
		}
	})

	if !inserted {
		t.Fatal("hopper did not insert into the destination chest")
	}
}

func TestHopperTracksRedstoneLock(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer func() { _ = w.Close() }()

	hopperPos := cube.Pos{0, 0, 0}
	powerPos := hopperPos.Side(cube.FaceEast)

	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(hopperPos, NewHopper(), nil)
		tx.SetBlock(powerPos, RedstoneBlock{}, nil)
	})
	w.AdvanceTick()
	runWorld(w, func(tx *world.Tx) {
		hopper := tx.Block(hopperPos).(Hopper)
		if !hopper.Powered {
			t.Error("hopper remained unlocked while receiving redstone power")
		}

		tx.SetBlock(powerPos, Air{}, nil)
	})
	w.AdvanceTick()
	runWorld(w, func(tx *world.Tx) {
		hopper := tx.Block(hopperPos).(Hopper)
		if hopper.Powered {
			t.Error("hopper remained locked after redstone power was removed")
		}
	})
}

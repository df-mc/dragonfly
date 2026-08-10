package session

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// TestInvByIDRefusesStaleWindow verifies that the window opened for a container is refused once the block it was
// opened for is gone. Breaking the block leaves the inventory the window holds detached from the world but still
// full, so a request that reached it would take a second copy of everything in it.
func TestInvByIDRefusesStaleWindow(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	pos := cube.Pos{0, 64, 0}
	w.Do(func(tx *world.Tx) {
		chest := block.NewChest()
		chest.Facing = cube.South
		tx.SetBlock(pos, chest, nil)

		s := &Session{}
		s.openedPos.Store(&pos)
		s.openedWindow.Store(chest.Inventory(tx, pos))
		s.containerOpened.Store(true)

		if _, ok := s.invByID(protocol.ContainerLevelEntity, tx); !ok {
			t.Fatal("the window of a chest that is still there was refused")
		}

		tx.SetBlock(pos, nil, nil)

		if _, ok := s.invByID(protocol.ContainerLevelEntity, tx); ok {
			t.Error("the window of a chest that was broken is still usable, so its contents can be taken twice")
		}
	}).Wait(context.Background())
}

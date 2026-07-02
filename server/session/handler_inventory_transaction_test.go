package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestInventoryTransactionDropHeldSlotUpdatesHeldItemState(t *testing.T) {
	var updates int
	handle := world.EntitySpawnOpts{}.New(heldItemStateTestType{}, heldItemStateTestConfig{updates: &updates})
	heldSlot := uint32(2)
	s := &Session{
		ent:      handle,
		heldSlot: &heldSlot,
		inv:      inventory.New(36, nil),
	}
	held := item.NewStack(item.Apple{}, 2)
	_ = s.inv.SetItem(int(heldSlot), held)
	w := world.Config{Entities: world.EntityRegistryConfig{}.New([]world.EntityType{heldItemStateTestType{}})}.New()
	defer func() {
		_ = w.Close()
	}()

	var err error
	<-w.Exec(func(tx *world.Tx) {
		tx.AddEntity(handle)
		err = (&InventoryTransactionHandler{}).handleNormalTransaction(&packet.InventoryTransaction{Actions: []protocol.InventoryAction{
			{
				SourceType:    protocol.InventoryActionSourceWorld,
				InventorySlot: 0,
				NewItem:       instanceFromItem(s.br, item.NewStack(item.Apple{}, 1)),
			},
			{
				SourceType:    protocol.InventoryActionSourceContainer,
				WindowID:      protocol.WindowIDInventory,
				InventorySlot: heldSlot,
				OldItem:       instanceFromItem(s.br, held),
			},
		}}, s, tx, heldItemDropper{})
	})
	if err != nil {
		t.Fatalf("expected drop transaction to succeed: %v", err)
	}
	if updates != 1 {
		t.Fatalf("expected held slot drop to update held item state once, got %v", updates)
	}
}

type heldItemDropper struct{}

func (heldItemDropper) Drop(s item.Stack) int {
	return s.Count()
}

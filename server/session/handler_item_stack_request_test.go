package session

import (
	"io"
	"log/slog"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// testSession builds the minimum Session needed to drive the item stack
// request handler: the player inventory, the UI inventory and an outgoing
// packet buffer.
func testSession() *Session {
	s := &Session{
		inv:             inventory.New(36, nil),
		ui:              inventory.New(54, nil),
		offHand:         inventory.New(1, nil),
		enderChest:      inventory.New(27, nil),
		packets:         make(chan packet.Packet, 256),
		closeBackground: make(chan struct{}),
		recipes:         map[uint32]recipe.Recipe{},
		conf:            Config{Log: slog.New(slog.NewTextHandler(io.Discard, nil))},
	}
	s.openedWindow.Store(inventory.New(1, nil))
	s.openedPos.Store(&cube.Pos{})
	return s
}

func testHandler() *ItemStackRequestHandler {
	return &ItemStackRequestHandler{
		changes:         map[byte]map[byte]changeInfo{},
		responseChanges: map[int32]map[*inventory.Inventory]map[byte]responseChange{},
	}
}

// invTotal counts every item of the type passed held across the session's
// inventories.
func invTotal(s *Session) int {
	var n int
	for _, inv := range []*inventory.Inventory{s.inv, s.ui, s.offHand} {
		for _, it := range inv.Slots() {
			n += it.Count()
		}
	}
	return n
}

func playerSlot(slot byte, id int32) protocol.StackRequestSlotInfo {
	return protocol.StackRequestSlotInfo{
		Container:      protocol.FullContainerName{ContainerID: protocol.ContainerCombinedHotBarAndInventory},
		Slot:           slot,
		StackNetworkID: id,
	}
}

// TestRejectRestoresItemCount verifies that undoing a request that failed part way through leaves the same number
// of items as before it started. Each slot is restored to what it held before the request, not to what it held
// between two writes of the same request.
func TestRejectRestoresItemCount(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	s, h := testSession(), testHandler()
	stack := item.NewStack(item.Diamond{}, 64)
	if err := s.inv.SetItem(0, stack); err != nil {
		t.Fatal(err)
	}
	before := invTotal(s)

	cur, _ := s.inv.Item(0)
	// 0 -> 1
	first := &protocol.PlaceStackRequestAction{}
	first.Count, first.Source, first.Destination = 1, playerSlot(0, item_id(cur)), playerSlot(1, 0)
	// 1 -> 2, referring to the change made by the previous action.
	second := &protocol.PlaceStackRequestAction{}
	second.Count, second.Source, second.Destination = 1, playerSlot(1, -1), playerSlot(2, 0)
	// Bogus action that fails verification and triggers reject().
	bogus := &protocol.TakeStackRequestAction{}
	bogus.Count, bogus.Source, bogus.Destination = 1, playerSlot(3, 987654), playerSlot(4, 0)

	req := protocol.ItemStackRequest{
		RequestID: -1,
		Actions:   []protocol.StackRequestAction{first, second, bogus},
	}

	w.Do(func(tx *world.Tx) {
		if err := h.Handle(&packet.ItemStackRequest{Requests: []protocol.ItemStackRequest{req}}, s, tx, nil); err != nil {
			t.Fatal(err)
		}
	})

	if after := invTotal(s); after != before {
		t.Fatalf("item count changed over a rejected request: want %d, got %d (slots 0..2: %d, %d, %d)",
			before, after, slotCount(s, 0), slotCount(s, 1), slotCount(s, 2))
	}
}

func slotCount(s *Session, slot int) int {
	it, _ := s.inv.Item(slot)
	return it.Count()
}

// TestTransferOntoSameSlot verifies that a request moving a slot onto itself does not change how many items it
// holds. Both slots are read before either is written, so such a request would write the count twice.
func TestTransferOntoSameSlot(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	s, h := testSession(), testHandler()
	if err := s.inv.SetItem(0, item.NewStack(item.Diamond{}, 32)); err != nil {
		t.Fatal(err)
	}
	before := invTotal(s)

	cur, _ := s.inv.Item(0)
	self := &protocol.PlaceStackRequestAction{}
	self.Count, self.Source, self.Destination = 32, playerSlot(0, item_id(cur)), playerSlot(0, item_id(cur))

	w.Do(func(tx *world.Tx) {
		if err := h.Handle(&packet.ItemStackRequest{Requests: []protocol.ItemStackRequest{{
			RequestID: -1,
			Actions:   []protocol.StackRequestAction{self},
		}}}, s, tx, nil); err != nil {
			t.Fatal(err)
		}
	})

	if after := invTotal(s); after != before {
		t.Fatalf("item count changed over a self-transfer: want %d, got %d (slot 0: %d)", before, after, slotCount(s, 0))
	}
}

// TestTransferOntoSameSlotUnderAnotherContainerID verifies that one slot named under two container IDs is refused.
// The crafting input, the created output and the cursor all address the same inventory, as do the hotbar and the rest
// of the inventory, so comparing the container IDs of a request cannot tell two slots apart.
func TestTransferOntoSameSlotUnderAnotherContainerID(t *testing.T) {
	tests := []struct {
		name     string
		from, to byte
	}{
		{name: "crafting input and cursor", from: protocol.ContainerCraftingInput, to: protocol.ContainerCursor},
		{name: "crafting input and created output", from: protocol.ContainerCraftingInput, to: protocol.ContainerCreatedOutput},
		{name: "hotbar and inventory", from: protocol.ContainerHotBar, to: protocol.ContainerInventory},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := world.Config{Synchronous: true}.New()
			defer w.Close()

			s, h := testSession(), testHandler()
			stack := item.NewStack(item.Diamond{}, 32)
			inv, _ := s.invByID(int32(tt.from), nil)
			if err := inv.SetItem(0, stack); err != nil {
				t.Fatalf("SetItem() = %v, want nil", err)
			}
			before := invTotal(s)

			id := item_id(stack)
			from := protocol.StackRequestSlotInfo{Container: protocol.FullContainerName{ContainerID: tt.from}, Slot: 0, StackNetworkID: id}
			to := protocol.StackRequestSlotInfo{Container: protocol.FullContainerName{ContainerID: tt.to}, Slot: 0, StackNetworkID: id}

			w.Do(func(tx *world.Tx) {
				_ = h.handleTransfer(from, to, 32, s, tx, nil)
			})

			if got := invTotal(s); got != before {
				t.Errorf("item count changed from %v to %v naming one slot under two container IDs", before, got)
			}
		})
	}
}

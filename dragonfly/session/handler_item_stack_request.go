package session

import (
	"fmt"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/gamemode"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"math"
	"sync/atomic"
)

// ItemStackRequestHandler handles the ItemStackRequest packet. It handles the actions done within the
// inventory.
type ItemStackRequestHandler struct {
	changes map[byte]map[byte]protocol.StackResponseSlotInfo
}

// Handle ...
func (h *ItemStackRequestHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.ItemStackRequest)

	atomic.StoreUint32(s.inTransaction, 1)
	defer atomic.StoreUint32(s.inTransaction, 0)

	for _, req := range pk.Requests {
		if err := h.handleRequest(req, s); err != nil {
			// Item stacks being out of sync isn't uncommon, so don't error. Just debug the error and let the
			// revert do its work.
			s.log.Debugf("error resolving item stack request: %v", err)
			return nil
		}
	}
	return nil
}

// handleRequest resolves a single item stack request from the client.
func (h *ItemStackRequestHandler) handleRequest(req protocol.ItemStackRequest, s *Session) (err error) {
	defer func() {
		if err != nil {
			h.reject(req.RequestID, s)
			return
		}
		h.resolve(req.RequestID, s)
	}()

	for _, action := range req.Actions {
		switch a := action.(type) {
		case *protocol.TakeStackRequestAction:
			err = h.handleTake(a, s)
		case *protocol.PlaceStackRequestAction:
			err = h.handlePlace(a, s)
		case *protocol.SwapStackRequestAction:
			err = h.handleSwap(a, s)
		case *protocol.DestroyStackRequestAction:
			err = h.handleDestroy(a, s)
		case *protocol.CraftCreativeStackRequestAction:
			err = h.handleCreativeCraft(a, s)
		case *protocol.CraftResultsDeprecatedStackRequestAction:
			// Don't do anything with this.
		default:
			return fmt.Errorf("unhandled stack request action %#v", action)
		}
		if err != nil {
			err = fmt.Errorf("%T: %w", action, err)
			return
		}
	}
	return
}

// handleTake handles a Take stack request action.
func (h *ItemStackRequestHandler) handleTake(a *protocol.TakeStackRequestAction, s *Session) error {
	return h.handleTransfer(a.Source, a.Destination, a.Count, s)
}

// handlePlace handles a Place stack request action.
func (h *ItemStackRequestHandler) handlePlace(a *protocol.PlaceStackRequestAction, s *Session) error {
	return h.handleTransfer(a.Source, a.Destination, a.Count, s)
}

// handleSwap handles a Swap stack request action.
func (h *ItemStackRequestHandler) handleSwap(a *protocol.SwapStackRequestAction, s *Session) error {
	if err := h.verifySlots(s, a.Source, a.Destination); err != nil {
		return fmt.Errorf("slot out of sync: %w", err)
	}
	i, _ := h.itemInSlot(a.Source, s)
	dest, _ := h.itemInSlot(a.Destination, s)

	h.setItemInSlot(a.Source, dest, s)
	h.setItemInSlot(a.Destination, i, s)

	return nil
}

// handleCreativeCraft handles the CreativeCraft request action.
func (h *ItemStackRequestHandler) handleCreativeCraft(a *protocol.CraftCreativeStackRequestAction, s *Session) error {
	if (s.c.GameMode() != gamemode.Creative{} && s.c.GameMode() != gamemode.Spectator{}) {
		return fmt.Errorf("can only craft creative items in gamemode creative/spectator")
	}
	index := a.CreativeItemNetworkID - 1
	if int(index) >= len(item.CreativeItems()) {
		return fmt.Errorf("creative item with network ID %v does not exist", index)
	}
	it := item.CreativeItems()[index]
	it = it.Grow(it.MaxCount())

	h.setItemInSlot(protocol.StackRequestSlotInfo{
		// This seems to be the creative inventory output slot.
		ContainerID:    59,
		Slot:           50,
		StackNetworkID: item_id(it),
	}, it, s)
	return nil
}

// handleDestroy handles the destroying of an item by moving it into the creative inventory.
func (h *ItemStackRequestHandler) handleDestroy(a *protocol.DestroyStackRequestAction, s *Session) error {
	if (s.c.GameMode() != gamemode.Creative{} && s.c.GameMode() != gamemode.Spectator{}) {
		return fmt.Errorf("can only destroy items in gamemode creative/spectator")
	}
	if err := h.verifySlot(a.Source, s); err != nil {
		return fmt.Errorf("source slot out of sync: %w", err)
	}
	i, _ := h.itemInSlot(a.Source, s)
	if i.Count() < int(a.Count) {
		return fmt.Errorf("client attempted to destroy %v items, but only %v present", a.Count, i.Count())
	}

	h.setItemInSlot(a.Source, i.Grow(-int(a.Count)), s)
	return nil
}

// handleTransfer handles the transferring of x count from a source slot to a destination slot.
func (h *ItemStackRequestHandler) handleTransfer(from, to protocol.StackRequestSlotInfo, count byte, s *Session) error {
	if err := h.verifySlots(s, from, to); err != nil {
		return fmt.Errorf("source slot out of sync: %w", err)
	}
	i, _ := h.itemInSlot(from, s)
	dest, _ := h.itemInSlot(to, s)
	if !i.Comparable(dest) {
		return fmt.Errorf("client tried transferring %v to %v, but the stacks are incomparable", i, dest)
	}
	if i.Count() < int(count) {
		return fmt.Errorf("client tried subtracting %v from item count, but there are only %v", count, i.Count())
	}
	if (dest.Count()+int(count) > dest.MaxCount()) && !dest.Empty() {
		return fmt.Errorf("client tried adding %v to item count %v, but max is %v", count, dest.Count(), dest.MaxCount())
	}
	if dest.Empty() {
		dest = i.Grow(-math.MaxInt32)
	}

	h.setItemInSlot(from, i.Grow(-int(count)), s)
	h.setItemInSlot(to, dest.Grow(int(count)), s)

	return nil
}

// verifySlots verifies a list of slots passed.
func (h *ItemStackRequestHandler) verifySlots(s *Session, slots ...protocol.StackRequestSlotInfo) error {
	for _, slot := range slots {
		if err := h.verifySlot(slot, s); err != nil {
			return err
		}
	}
	return nil
}

// verifySlot checks if the slot passed by the client is the same as that expected by the server.
func (h *ItemStackRequestHandler) verifySlot(slot protocol.StackRequestSlotInfo, s *Session) error {
	i, err := h.itemInSlot(slot, s)
	if err != nil {
		return err
	}
	// The client seems to send negative stack network IDs for predictions, which we can ignore. We'll simply
	// override this network ID later.
	if id := item_id(i); id != slot.StackNetworkID && slot.StackNetworkID > 0 {
		return fmt.Errorf("stack ID mismatch: client expected %v, but server had %v", slot.StackNetworkID, id)
	}
	return nil
}

// itemInSlot looks for the item in the slot as indicated by the slot info passed.
func (h *ItemStackRequestHandler) itemInSlot(slot protocol.StackRequestSlotInfo, s *Session) (item.Stack, error) {
	inventory, ok := s.invByID(int32(slot.ContainerID))
	if !ok {
		return item.Stack{}, fmt.Errorf("unable to find container with ID %v", slot.ContainerID)
	}
	i, err := inventory.Item(int(slot.Slot))
	if err != nil {
		return i, err
	}
	return i, nil
}

// setItemInSlot sets an item stack in the slot of a container present in the slot info.
func (h *ItemStackRequestHandler) setItemInSlot(slot protocol.StackRequestSlotInfo, i item.Stack, s *Session) {
	inventory, _ := s.invByID(int32(slot.ContainerID))
	_ = inventory.SetItem(int(slot.Slot), i)

	if h.changes[slot.ContainerID] == nil {
		h.changes[slot.ContainerID] = map[byte]protocol.StackResponseSlotInfo{}
	}
	h.changes[slot.ContainerID][slot.Slot] = protocol.StackResponseSlotInfo{
		Slot:           slot.Slot,
		HotbarSlot:     slot.Slot,
		Count:          byte(i.Count()),
		StackNetworkID: item_id(i),
	}
}

// resolve resolves the request with the ID passed.
func (h *ItemStackRequestHandler) resolve(id int32, s *Session) {
	info := make([]protocol.StackResponseContainerInfo, 0, len(h.changes))
	for container, slotInfo := range h.changes {
		slots := make([]protocol.StackResponseSlotInfo, 0, len(slotInfo))
		for _, slot := range slotInfo {
			slots = append(slots, slot)
		}
		info = append(info, protocol.StackResponseContainerInfo{
			ContainerID: container,
			SlotInfo:    slots,
		})
	}
	s.writePacket(&packet.ItemStackResponse{Responses: []protocol.ItemStackResponse{{
		Success:       true,
		RequestID:     id,
		ContainerInfo: info,
	}}})
	h.changes = map[byte]map[byte]protocol.StackResponseSlotInfo{}
}

// reject rejects the item stack request sent by the client so that it is reverted client-side.
func (h *ItemStackRequestHandler) reject(id int32, s *Session) {
	s.writePacket(&packet.ItemStackResponse{
		Responses: []protocol.ItemStackResponse{{
			Success:   false,
			RequestID: id,
		}},
	})
	h.changes = map[byte]map[byte]protocol.StackResponseSlotInfo{}
}

package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/creative"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"math"
	"time"
)

// ItemStackRequestHandler handles the ItemStackRequest packet. It handles the actions done within the
// inventory.
type ItemStackRequestHandler struct {
	currentRequest  int32
	changes         map[byte]map[byte]changeInfo
	responseChanges map[int32]map[*inventory.Inventory]map[byte]responseChange
	current         time.Time
	ignoreDestroy   bool
}

// responseChange represents a change in a specific item stack response. It holds the timestamp of the
// response which is used to get rid of changes that the client will have received.
type responseChange struct {
	id        int32
	timestamp time.Time
}

// changeInfo holds information on a slot change initiated by an item stack request. It holds both the new and the old
// item information and is used for reverting and verifying.
type changeInfo struct {
	after  protocol.StackResponseSlotInfo
	before item.Stack
}

// Handle ...
func (h *ItemStackRequestHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.ItemStackRequest)
	h.current = time.Now()

	s.inTransaction.Store(true)
	defer s.inTransaction.Store(false)

	for _, req := range pk.Requests {
		if err := h.handleRequest(req, s); err != nil {
			// Item stacks being out of sync isn't uncommon, so don't error. Just debug the error and let the
			// revert do its work.
			s.log.Debugf("failed processing packet from %v (%v): ItemStackRequest: error resolving item stack request: %v", s.conn.RemoteAddr(), s.c.Name(), err)
		}
	}
	return nil
}

// handleRequest resolves a single item stack request from the client.
func (h *ItemStackRequestHandler) handleRequest(req protocol.ItemStackRequest, s *Session) (err error) {
	h.currentRequest = req.RequestID
	defer func() {
		if err != nil {
			s.log.Debugf("%v", err)
			h.reject(req.RequestID, s)
			return
		}
		h.resolve(req.RequestID, s)
		h.ignoreDestroy = false
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
		case *protocol.DropStackRequestAction:
			err = h.handleDrop(a, s)
		case *protocol.BeaconPaymentStackRequestAction:
			err = h.handleBeaconPayment(a, s)
		case *protocol.CraftRecipeStackRequestAction:
			err = h.handleCraft(a, s)
		case *protocol.AutoCraftRecipeStackRequestAction:
			err = h.handleAutoCraft(a, s)
		case *protocol.CraftCreativeStackRequestAction:
			err = h.handleCreativeCraft(a, s)
		case *protocol.MineBlockStackRequestAction:
			err = h.handleMineBlock(a, s)
		case *protocol.ConsumeStackRequestAction, *protocol.CraftResultsDeprecatedStackRequestAction:
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

	invA, _ := s.invByID(int32(from.ContainerID))
	invB, _ := s.invByID(int32(to.ContainerID))

	ctx := event.C()
	_ = call(ctx, int(from.Slot), i.Grow(int(count)-i.Count()), invA.Handler().HandleTake)
	err := call(ctx, int(to.Slot), i.Grow(int(count)-i.Count()), invB.Handler().HandlePlace)
	if err != nil {
		return err
	}

	h.setItemInSlot(from, i.Grow(-int(count)), s)
	h.setItemInSlot(to, dest.Grow(int(count)), s)

	return nil
}

// handleSwap handles a Swap stack request action.
func (h *ItemStackRequestHandler) handleSwap(a *protocol.SwapStackRequestAction, s *Session) error {
	if err := h.verifySlots(s, a.Source, a.Destination); err != nil {
		return fmt.Errorf("slot out of sync: %w", err)
	}
	i, _ := h.itemInSlot(a.Source, s)
	dest, _ := h.itemInSlot(a.Destination, s)

	invA, _ := s.invByID(int32(a.Source.ContainerID))
	invB, _ := s.invByID(int32(a.Destination.ContainerID))

	ctx := event.C()
	_ = call(ctx, int(a.Source.Slot), i, invA.Handler().HandleTake)
	_ = call(ctx, int(a.Source.Slot), dest, invA.Handler().HandlePlace)
	_ = call(ctx, int(a.Destination.Slot), dest, invB.Handler().HandleTake)
	err := call(ctx, int(a.Destination.Slot), i, invB.Handler().HandlePlace)
	if err != nil {
		return err
	}

	h.setItemInSlot(a.Source, dest, s)
	h.setItemInSlot(a.Destination, i, s)

	return nil
}

// call uses an event.Context, slot and item.Stack to call the event handler function passed. An error is returned if
// the event.Context was cancelled either before or after the call.
func call(ctx *event.Context, slot int, it item.Stack, f func(ctx *event.Context, slot int, it item.Stack)) error {
	if ctx.Cancelled() {
		return fmt.Errorf("action was cancelled")
	}
	f(ctx, slot, it)
	if ctx.Cancelled() {
		return fmt.Errorf("action was cancelled")
	}
	return nil
}

// handleCraft handles the CraftRecipe request action.
func (h *ItemStackRequestHandler) handleCraft(a *protocol.CraftRecipeStackRequestAction, s *Session) error {
	if s.c.GameMode().CreativeInventory() {
		return fmt.Errorf("can only craft items in gamemode survival/adventure")
	}
	if int(a.RecipeNetworkID) > len(s.recipes) {
		return fmt.Errorf("recipe with network id %v does not exist (cap: %v)", a.RecipeNetworkID, len(s.recipes))
	}

	craft := s.recipes[a.RecipeNetworkID]
	_, shaped := craft.(recipe.ShapedRecipe)
	_, shapeless := craft.(recipe.ShapelessRecipe)
	if !shaped && !shapeless {
		return fmt.Errorf("recipe with network id %v is not a shaped or shapeless recipe", a.RecipeNetworkID)
	}

	input := craft.Input()
	size := s.craftingSize()
	offset := s.craftingOffset()
	consumptions := make(map[uint32]recipe.InputItem)
	for _, expected := range input {
		var consumed bool
		for slot := offset; slot < offset+size; slot++ {
			if _, ok := consumptions[slot]; ok {
				// Already consumed this slot, skip to the next.
				continue
			}
			has, _ := s.ui.Item(int(slot))
			if has.Empty() && !expected.Empty() {
				// Expected is not empty, but we don't have anything.
				continue
			}
			if !has.Empty() && expected.Empty() {
				// Expected is empty, but we have an item.
				continue
			}
			if has.Empty() && expected.Empty() {
				consumed = true
				break
			}
			if !expected.Variants && has.Comparable(expected.Stack) {
				consumed, consumptions[slot] = true, expected
				break
			}
			name, _ := has.Item().EncodeItem()
			otherName, _ := expected.Item().EncodeItem()
			if name == otherName && has.Count() >= expected.Count() {
				consumed, consumptions[slot] = true, expected
				break
			}
		}
		if !consumed {
			return fmt.Errorf("could not consume item %v from crafting grid", expected)
		}
	}
	for slot, expected := range consumptions {
		has, _ := s.ui.Item(int(slot))
		st := has.Grow(-expected.Count())
		h.setItemInSlot(protocol.StackRequestSlotInfo{
			ContainerID:    containerCraftingGrid,
			Slot:           byte(slot),
			StackNetworkID: item_id(st),
		}, st, s)
	}

	output := craft.Output()
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID:    containerCraftingResult,
		Slot:           craftingResultIndex,
		StackNetworkID: item_id(output[0]),
	}, output[0], s)
	return nil
}

// handleAutoCraft handles the AutoCraftRecipe request action.
func (h *ItemStackRequestHandler) handleAutoCraft(a *protocol.AutoCraftRecipeStackRequestAction, s *Session) error {
	if s.c.GameMode().CreativeInventory() {
		return fmt.Errorf("can only craft items in gamemode survival/adventure")
	}
	if int(a.RecipeNetworkID) > len(s.recipes) {
		return fmt.Errorf("recipe with network id %v does not exist (cap: %v)", a.RecipeNetworkID, len(s.recipes))
	}

	craft := s.recipes[a.RecipeNetworkID]
	_, shaped := craft.(recipe.ShapedRecipe)
	_, shapeless := craft.(recipe.ShapelessRecipe)
	if !shaped && !shapeless {
		return fmt.Errorf("recipe with network id %v is not a shaped or shapeless recipe", a.RecipeNetworkID)
	}

	input := make([]recipe.InputItem, 0, len(craft.Input()))
	output := make([]item.Stack, 0, len(craft.Output()))
	for _, i := range craft.Input() {
		i.Stack = i.Grow(i.Count() * (int(a.TimesCrafted) - 1))
		input = append(input, i)
	}
	for _, o := range craft.Output() {
		output = append(output, o.Grow(o.Count()*(int(a.TimesCrafted)-1)))
	}

	for _, expected := range input {
		var consumed bool
		for _, has := range append(s.inv.Items(), s.ui.Items()...) {
			_ = has
			// TODO
		}
		if !consumed {
			return fmt.Errorf("could not consume item %v from inventory", expected)
		}
	}

	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID:    containerCraftingResult,
		Slot:           craftingResultIndex,
		StackNetworkID: item_id(output[0]),
	}, output[0], s)
	return nil
}

// handleCreativeCraft handles the CreativeCraft request action.
func (h *ItemStackRequestHandler) handleCreativeCraft(a *protocol.CraftCreativeStackRequestAction, s *Session) error {
	if !s.c.GameMode().CreativeInventory() {
		return fmt.Errorf("can only craft creative items in gamemode creative/spectator")
	}
	index := a.CreativeItemNetworkID - 1
	if int(index) >= len(creative.Items()) {
		return fmt.Errorf("creative item with network ID %v does not exist", index)
	}
	it := creative.Items()[index]
	it = it.Grow(it.MaxCount() - 1)

	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID:    containerCreativeOutput,
		Slot:           50,
		StackNetworkID: item_id(it),
	}, it, s)
	return nil
}

// handleDestroy handles the destroying of an item by moving it into the creative inventory.
func (h *ItemStackRequestHandler) handleDestroy(a *protocol.DestroyStackRequestAction, s *Session) error {
	if h.ignoreDestroy {
		return nil
	}
	if !s.c.GameMode().CreativeInventory() {
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

// handleDrop handles the dropping of an item by moving it outside the inventory while having the
// inventory opened.
func (h *ItemStackRequestHandler) handleDrop(a *protocol.DropStackRequestAction, s *Session) error {
	if err := h.verifySlot(a.Source, s); err != nil {
		return fmt.Errorf("source slot out of sync: %w", err)
	}
	i, _ := h.itemInSlot(a.Source, s)
	if i.Count() < int(a.Count) {
		return fmt.Errorf("client attempted to drop %v items, but only %v present", a.Count, i.Count())
	}

	inv, _ := s.invByID(int32(a.Source.ContainerID))
	if err := call(event.C(), int(a.Source.Slot), i.Grow(int(a.Count)-i.Count()), inv.Handler().HandleDrop); err != nil {
		return err
	}

	n := s.c.Drop(i.Grow(int(a.Count) - i.Count()))
	h.setItemInSlot(a.Source, i.Grow(-n), s)
	return nil
}

// handleBeaconPayment handles the selection of effects in a beacon and the removal of the item used to pay
// for those effects.
func (h *ItemStackRequestHandler) handleBeaconPayment(a *protocol.BeaconPaymentStackRequestAction, s *Session) error {
	slot := protocol.StackRequestSlotInfo{
		ContainerID: containerBeacon,
		Slot:        0x1b,
	}
	// First check if there actually is a beacon opened.
	if !s.containerOpened.Load() {
		return fmt.Errorf("no beacon container opened")
	}
	pos := s.openedPos.Load()
	beacon, ok := s.c.World().Block(pos).(block.Beacon)
	if !ok {
		return fmt.Errorf("no beacon container opened")
	}

	// Check if the item present in the beacon slot is valid.
	payment, _ := h.itemInSlot(slot, s)
	if payable, ok := payment.Item().(item.BeaconPayment); !ok || !payable.PayableForBeacon() {
		return fmt.Errorf("item %#v in beacon slot cannot be used as payment", payment)
	}

	// Check if the effects are valid and allowed for the beacon's level.
	if !h.validBeaconEffect(a.PrimaryEffect, beacon) {
		return fmt.Errorf("primary effect selected is not allowed: %v for level %v", a.PrimaryEffect, beacon.Level())
	} else if !h.validBeaconEffect(a.SecondaryEffect, beacon) || (beacon.Level() < 4 && a.SecondaryEffect != 0) {
		return fmt.Errorf("secondary effect selected is not allowed: %v for level %v", a.SecondaryEffect, beacon.Level())
	}

	primary, pOk := effect.ByID(int(a.PrimaryEffect))
	secondary, sOk := effect.ByID(int(a.SecondaryEffect))
	if pOk {
		beacon.Primary = primary.(effect.LastingType)
	}
	if sOk {
		beacon.Secondary = secondary.(effect.LastingType)
	}
	s.c.World().SetBlock(pos, beacon, nil)

	// The client will send a Destroy action after this action, but we can't rely on that because the client
	// could just not send it.
	// We just ignore the next Destroy action and set the item to air here.
	h.setItemInSlot(slot, item.NewStack(block.Air{}, 0), s)
	h.ignoreDestroy = true
	return nil
}

// handleMineBlock handles the action associated with a block being mined by the player. This seems to be a workaround
// by Mojang to deal with the durability changes client-side.
func (h *ItemStackRequestHandler) handleMineBlock(a *protocol.MineBlockStackRequestAction, s *Session) error {
	slot := protocol.StackRequestSlotInfo{
		ContainerID:    containerInventory,
		Slot:           byte(a.HotbarSlot),
		StackNetworkID: a.StackNetworkID,
	}
	if err := h.verifySlot(slot, s); err != nil {
		return err
	}

	// Update the slots through ItemStackResponses, don't actually do anything special with this action.
	i, _ := h.itemInSlot(slot, s)
	h.setItemInSlot(slot, i, s)

	return nil
}

// validBeaconEffect checks if the ID passed is a valid beacon effect.
func (h *ItemStackRequestHandler) validBeaconEffect(id int32, beacon block.Beacon) bool {
	switch id {
	case 1, 3:
		return beacon.Level() >= 1
	case 8, 11:
		return beacon.Level() >= 2
	case 5:
		return beacon.Level() >= 3
	case 10:
		return beacon.Level() >= 4
	case 0:
		return true
	}
	return false
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
	if err := h.tryAcknowledgeChanges(s, slot); err != nil {
		return err
	}
	if len(h.responseChanges) > 256 {
		return fmt.Errorf("too many unacknowledged request slot changes")
	}
	inv, _ := s.invByID(int32(slot.ContainerID))

	i, err := h.itemInSlot(slot, s)
	if err != nil {
		return err
	}
	clientID, err := h.resolveID(inv, slot)
	if err != nil {
		return err
	}
	// The client seems to send negative stack network IDs for predictions, which we can ignore. We'll simply
	// override this network ID later.
	if id := item_id(i); id != clientID {
		return fmt.Errorf("stack ID mismatch: client expected %v, but server had %v", clientID, id)
	}
	return nil
}

// resolveID resolves the stack network ID in the slot passed. If it is negative, it points to an earlier
// request, in which case it will look it up in the changes of an earlier response to a request to find the
// actual stack network ID in the slot. If it is positive, the ID will be returned again.
func (h *ItemStackRequestHandler) resolveID(inv *inventory.Inventory, slot protocol.StackRequestSlotInfo) (int32, error) {
	if slot.StackNetworkID >= 0 {
		return slot.StackNetworkID, nil
	}
	containerChanges, ok := h.responseChanges[slot.StackNetworkID]
	if !ok {
		return 0, fmt.Errorf("slot pointed to stack request %v, but request could not be found", slot.StackNetworkID)
	}
	changes, ok := containerChanges[inv]
	if !ok {
		return 0, fmt.Errorf("slot pointed to stack request %v with container %v, but that container was not changed in the request", slot.StackNetworkID, slot.ContainerID)
	}
	actual, ok := changes[slot.Slot]
	if !ok {
		return 0, fmt.Errorf("slot pointed to stack request %v with container %v and slot %v, but that slot was not changed in the request", slot.StackNetworkID, slot.ContainerID, slot.Slot)
	}
	return actual.id, nil
}

// tryAcknowledgeChanges iterates through all cached response changes and checks if the stack request slot
// info passed from the client has the right stack network ID in any of the stored slots. If this is the case,
// that entry is removed, so that the maps are cleaned up eventually.
func (h *ItemStackRequestHandler) tryAcknowledgeChanges(s *Session, slot protocol.StackRequestSlotInfo) error {
	inv, ok := s.invByID(int32(slot.ContainerID))
	if !ok {
		return fmt.Errorf("could not find container with id %v", slot.ContainerID)
	}

	for requestID, containerChanges := range h.responseChanges {
		for newInv, changes := range containerChanges {
			for slotIndex, val := range changes {
				if (slot.Slot == slotIndex && slot.StackNetworkID >= 0 && newInv == inv) || h.current.Sub(val.timestamp) > time.Second*5 {
					delete(changes, slotIndex)
				}
			}
			if len(changes) == 0 {
				delete(containerChanges, newInv)
			}
		}
		if len(containerChanges) == 0 {
			delete(h.responseChanges, requestID)
		}
	}
	return nil
}

// itemInSlot looks for the item in the slot as indicated by the slot info passed.
func (h *ItemStackRequestHandler) itemInSlot(slot protocol.StackRequestSlotInfo, s *Session) (item.Stack, error) {
	inv, ok := s.invByID(int32(slot.ContainerID))
	if !ok {
		return item.Stack{}, fmt.Errorf("unable to find container with ID %v", slot.ContainerID)
	}

	sl := int(slot.Slot)
	if inv == s.offHand {
		sl = 0
	}

	i, err := inv.Item(sl)
	if err != nil {
		return i, err
	}
	return i, nil
}

// setItemInSlot sets an item stack in the slot of a container present in the slot info.
func (h *ItemStackRequestHandler) setItemInSlot(slot protocol.StackRequestSlotInfo, i item.Stack, s *Session) {
	inv, _ := s.invByID(int32(slot.ContainerID))

	sl := int(slot.Slot)
	if inv == s.offHand {
		sl = 0
	}

	before, _ := inv.Item(sl)
	_ = inv.SetItem(sl, i)

	respSlot := protocol.StackResponseSlotInfo{
		Slot:                 slot.Slot,
		HotbarSlot:           slot.Slot,
		Count:                byte(i.Count()),
		StackNetworkID:       item_id(i),
		DurabilityCorrection: int32(i.MaxDurability() - i.Durability()),
	}

	if h.changes[slot.ContainerID] == nil {
		h.changes[slot.ContainerID] = map[byte]changeInfo{}
	}
	h.changes[slot.ContainerID][slot.Slot] = changeInfo{
		after:  respSlot,
		before: before,
	}

	if h.responseChanges[h.currentRequest] == nil {
		h.responseChanges[h.currentRequest] = map[*inventory.Inventory]map[byte]responseChange{}
	}
	if h.responseChanges[h.currentRequest][inv] == nil {
		h.responseChanges[h.currentRequest][inv] = map[byte]responseChange{}
	}
	h.responseChanges[h.currentRequest][inv][slot.Slot] = responseChange{
		id:        respSlot.StackNetworkID,
		timestamp: h.current,
	}
}

// resolve resolves the request with the ID passed.
func (h *ItemStackRequestHandler) resolve(id int32, s *Session) {
	info := make([]protocol.StackResponseContainerInfo, 0, len(h.changes))
	for container, slotInfo := range h.changes {
		slots := make([]protocol.StackResponseSlotInfo, 0, len(slotInfo))
		for _, slot := range slotInfo {
			slots = append(slots, slot.after)
		}
		info = append(info, protocol.StackResponseContainerInfo{
			ContainerID: container,
			SlotInfo:    slots,
		})
	}
	s.writePacket(&packet.ItemStackResponse{Responses: []protocol.ItemStackResponse{{
		Status:        protocol.ItemStackResponseStatusOK,
		RequestID:     id,
		ContainerInfo: info,
	}}})
	h.changes = map[byte]map[byte]changeInfo{}
}

// reject rejects the item stack request sent by the client so that it is reverted client-side.
func (h *ItemStackRequestHandler) reject(id int32, s *Session) {
	s.writePacket(&packet.ItemStackResponse{
		Responses: []protocol.ItemStackResponse{{
			Status:    protocol.ItemStackResponseStatusError,
			RequestID: id,
		}},
	})
	// Revert changes that we already made for valid actions.
	for container, slots := range h.changes {
		for slot, info := range slots {
			inv, _ := s.invByID(int32(container))
			_ = inv.SetItem(int(slot), info.before)
		}
	}
	h.changes = map[byte]map[byte]changeInfo{}
}

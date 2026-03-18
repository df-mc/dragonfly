package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/creative"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// handleCraft handles the CraftRecipe request action.
func (h *ItemStackRequestHandler) handleCraft(a *protocol.CraftRecipeStackRequestAction, s *Session, tx *world.Tx, c Controllable) error {
	craft, ok := s.recipes[a.RecipeNetworkID]
	if !ok {
		// Try dynamic recipes if no static recipe matches
		plan, err := c.DynamicCraftItem(int(a.NumberOfCrafts))
		if err != nil {
			return err
		}
		return h.applyCraftingPlan(plan, s, tx)
	}
	plan, err := c.CraftItem(craft, int(a.NumberOfCrafts))
	if err != nil {
		return err
	}
	return h.applyCraftingPlan(plan, s, tx)
}

// handleAutoCraft handles the AutoCraftRecipe request action.
func (h *ItemStackRequestHandler) handleAutoCraft(a *protocol.AutoCraftRecipeStackRequestAction, s *Session, tx *world.Tx, c Controllable) error {
	craft, ok := s.recipes[a.RecipeNetworkID]
	if !ok {
		// Try dynamic recipes if no static recipe matches
		plan, err := c.DynamicCraftItem(int(a.TimesCrafted))
		if err != nil {
			return err
		}
		return h.applyCraftingPlan(plan, s, tx)
	}
	plan, err := c.AutoCraftItem(craft, int(a.TimesCrafted))
	if err != nil {
		return err
	}
	return h.applyCraftingPlan(plan, s, tx)
}

// handleCreativeCraft handles the CreativeCraft request action.
func (h *ItemStackRequestHandler) handleCreativeCraft(a *protocol.CraftCreativeStackRequestAction, s *Session, tx *world.Tx, c Controllable) error {
	if !c.GameMode().CreativeInventory() {
		return fmt.Errorf("can only craft creative items in gamemode creative/spectator")
	}
	index := a.CreativeItemNetworkID - 1
	if int(index) >= len(creative.Items()) {
		return fmt.Errorf("creative item with network ID %v does not exist", index)
	}
	it := creative.Items()[index].Stack
	it = it.Grow(it.MaxCount() - 1)
	return h.createResults(s, tx, it)
}

// applyCraftingPlan applies a crafting plan returned by a controllable and creates the resulting output items.
func (h *ItemStackRequestHandler) applyCraftingPlan(plan recipe.CraftingPlan, s *Session, tx *world.Tx) error {
	for _, change := range plan.Changes {
		if err := h.setItemInCraftingInventory(change.Inventory, change.Slot, change.Stack, s); err != nil {
			return err
		}
	}
	return h.createResults(s, tx, plan.Results...)
}

// setItemInCraftingInventory applies a crafting slot change using the correct client container metadata.
func (h *ItemStackRequestHandler) setItemInCraftingInventory(inv *inventory.Inventory, slot int, stack item.Stack, s *Session) error {
	info, err := s.craftingSlotInfo(inv, slot)
	if err != nil {
		return err
	}

	before, _ := inv.Item(slot)
	_ = inv.SetItem(slot, stack)

	respSlot := protocol.StackResponseSlotInfo{
		Slot:                 info.Slot,
		HotbarSlot:           info.Slot,
		Count:                byte(stack.Count()),
		StackNetworkID:       item_id(stack),
		DurabilityCorrection: int32(stack.MaxDurability() - stack.Durability()),
	}

	if h.changes[info.Container.ContainerID] == nil {
		h.changes[info.Container.ContainerID] = map[byte]changeInfo{}
	}
	h.changes[info.Container.ContainerID][info.Slot] = changeInfo{
		after:  respSlot,
		before: before,
	}

	if h.responseChanges[h.currentRequest] == nil {
		h.responseChanges[h.currentRequest] = map[*inventory.Inventory]map[byte]responseChange{}
	}
	if h.responseChanges[h.currentRequest][inv] == nil {
		h.responseChanges[h.currentRequest][inv] = map[byte]responseChange{}
	}
	h.responseChanges[h.currentRequest][inv][info.Slot] = responseChange{
		id:        respSlot.StackNetworkID,
		timestamp: h.current,
	}
	return nil
}

// craftingSlotInfo resolves the client-facing slot information for an inventory slot used by the crafting handlers.
func (s *Session) craftingSlotInfo(inv *inventory.Inventory, slot int) (protocol.StackRequestSlotInfo, error) {
	switch inv {
	case s.ui:
		return protocol.StackRequestSlotInfo{
			Container: protocol.FullContainerName{ContainerID: protocol.ContainerCraftingInput},
			Slot:      byte(slot),
		}, nil
	case s.inv:
		return protocol.StackRequestSlotInfo{
			Container: protocol.FullContainerName{ContainerID: protocol.ContainerCombinedHotBarAndInventory},
			Slot:      byte(slot),
		}, nil
	default:
		return protocol.StackRequestSlotInfo{}, fmt.Errorf("unsupported crafting inventory")
	}
}

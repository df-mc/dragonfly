package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/creative"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"golang.org/x/exp/slices"
)

// handleCraft handles the CraftRecipe request action.
func (h *ItemStackRequestHandler) handleCraft(a *protocol.CraftRecipeStackRequestAction, s *Session) error {
	craft, ok := s.recipes[a.RecipeNetworkID]
	if !ok {
		return fmt.Errorf("recipe with network id %v does not exist", a.RecipeNetworkID)
	}
	_, shaped := craft.(recipe.Shaped)
	_, shapeless := craft.(recipe.Shapeless)
	if !shaped && !shapeless {
		return fmt.Errorf("recipe with network id %v is not a shaped or shapeless recipe", a.RecipeNetworkID)
	}

	size := s.craftingSize()
	offset := s.craftingOffset()
	consumed := make([]bool, size)
	for _, expected := range craft.Input() {
		var processed bool
		for slot := offset; slot < offset+size; slot++ {
			if consumed[slot-offset] {
				// We've already consumed this slot, skip it.
				continue
			}
			has, _ := s.ui.Item(int(slot))
			_, variants := expected.Value("variants")
			if has.Empty() != expected.Empty() || has.Count() < expected.Count() {
				// We can't process this item, as it's not a part of the recipe.
				continue
			}
			if !variants && !has.Comparable(expected) {
				// Not the same item without accounting for variants.
				continue
			}
			if variants {
				nameOne, _ := has.Item().EncodeItem()
				nameTwo, _ := expected.Item().EncodeItem()
				if nameOne != nameTwo {
					// Not the same item even when accounting for variants.
					continue
				}
			}
			processed, consumed[slot-offset] = true, true
			st := has.Grow(-expected.Count())
			h.setItemInSlot(protocol.StackRequestSlotInfo{
				ContainerID:    containerCraftingGrid,
				Slot:           byte(slot),
				StackNetworkID: item_id(st),
			}, st, s)
			break
		}
		if !processed {
			return fmt.Errorf("recipe %v: could not consume expected item: %v", a.RecipeNetworkID, expected)
		}
	}

	output := craft.Output()
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID:    containerCraftingGrid,
		Slot:           craftingResult,
		StackNetworkID: item_id(output[0]),
	}, output[0], s)
	return nil
}

// handleAutoCraft handles the AutoCraftRecipe request action.
func (h *ItemStackRequestHandler) handleAutoCraft(a *protocol.AutoCraftRecipeStackRequestAction, s *Session) error {
	craft, ok := s.recipes[a.RecipeNetworkID]
	if !ok {
		return fmt.Errorf("recipe with network id %v does not exist", a.RecipeNetworkID)
	}
	_, shaped := craft.(recipe.Shaped)
	_, shapeless := craft.(recipe.Shapeless)
	if !shaped && !shapeless {
		return fmt.Errorf("recipe with network id %v is not a shaped or shapeless recipe", a.RecipeNetworkID)
	}

	input := make([]item.Stack, 0, len(craft.Input()))
	for _, i := range craft.Input() {
		input = append(input, i.Grow(i.Count()*(int(a.TimesCrafted)-1)))
	}

	expectancies := make([]item.Stack, 0, len(input))
	for _, i := range input {
		if i.Empty() {
			// We don't actually need this item - it's empty, so avoid putting it in our expectancies.
			continue
		}

		_, variants := i.Value("variants")
		if ind := slices.IndexFunc(expectancies, func(st item.Stack) bool {
			if variants {
				nameOne, _ := st.Item().EncodeItem()
				nameTwo, _ := i.Item().EncodeItem()
				return nameOne == nameTwo
			}
			return st.Comparable(i)
		}); ind >= 0 {
			i = i.Grow(expectancies[ind].Count())
			expectancies = slices.Delete(expectancies, ind, ind+1)
		}
		expectancies = append(expectancies, i)
	}

	for _, expected := range expectancies {
		_, variants := expected.Value("variants")
		for id, inv := range map[byte]*inventory.Inventory{containerCraftingGrid: s.ui, containerFullInventory: s.inv} {
			for slot, has := range inv.Slots() {
				if has.Empty() {
					// We don't have this item, skip it.
					continue
				}
				if !variants && !has.Comparable(expected) {
					// Not the same item without accounting for variants.
					continue
				}
				if variants {
					nameOne, _ := has.Item().EncodeItem()
					nameTwo, _ := expected.Item().EncodeItem()
					if nameOne != nameTwo {
						// Not the same item even when accounting for variants.
						continue
					}
				}

				remaining, removal := expected.Count(), has.Count()
				if remaining < removal {
					removal = remaining
				}

				expected, has = expected.Grow(-removal), has.Grow(-removal)
				h.setItemInSlot(protocol.StackRequestSlotInfo{
					ContainerID: id,
					Slot:        byte(slot),
				}, has, s)
				if expected.Empty() {
					// Consumed this item, so go to the next one.
					break
				}
			}
			if expected.Empty() {
				// Consumed this item, so go to the next one.
				break
			}
		}
		if !expected.Empty() {
			return fmt.Errorf("recipe %v: could not consume expected item: %v", a.RecipeNetworkID, expected)
		}
	}

	output := make([]item.Stack, 0, len(craft.Output()))
	for _, o := range craft.Output() {
		output = append(output, o.Grow(o.Count()*(int(a.TimesCrafted)-1)))
	}
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerCraftingGrid,
		Slot:        craftingResult,
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
		ContainerID: containerOutput,
		Slot:        craftingResult,
	}, it, s)
	return nil
}

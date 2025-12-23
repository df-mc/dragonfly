package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/creative"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
	"slices"
)

// handleCraft handles the CraftRecipe request action.
func (h *ItemStackRequestHandler) handleCraft(a *protocol.CraftRecipeStackRequestAction, s *Session, tx *world.Tx) error {
	craft, ok := s.recipes[a.RecipeNetworkID]
	if !ok {
		// Try dynamic recipes if no static recipe matches
		return h.tryDynamicCraft(s, tx, int(a.NumberOfCrafts))
	}
	_, shaped := craft.(recipe.Shaped)
	_, shapeless := craft.(recipe.Shapeless)
	if !shaped && !shapeless {
		return fmt.Errorf("recipe with network id %v is not a shaped or shapeless recipe", a.RecipeNetworkID)
	}
	if craft.Block() != "crafting_table" {
		return fmt.Errorf("recipe with network id %v is not a crafting table recipe", a.RecipeNetworkID)
	}

	timesCrafted := int(a.NumberOfCrafts)
	if timesCrafted < 1 {
		return fmt.Errorf("times crafted must be at least 1")
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
			if has.Empty() != expected.Empty() || has.Count() < expected.Count()*timesCrafted {
				// We can't process this item, as it's not a part of the recipe.
				continue
			}
			if !matchingStacks(has, expected) {
				// Not the same item.
				continue
			}
			processed, consumed[slot-offset] = true, true
			st := has.Grow(-expected.Count() * timesCrafted)
			h.setItemInSlot(protocol.StackRequestSlotInfo{
				Container: protocol.FullContainerName{ContainerID: protocol.ContainerCraftingInput},
				Slot:      byte(slot),
			}, st, s, tx)
			break
		}
		if !processed {
			return fmt.Errorf("recipe %v: could not consume expected item: %v", a.RecipeNetworkID, expected)
		}
	}
	return h.createResults(s, tx, repeatStacks(craft.Output(), timesCrafted)...)
}

// handleAutoCraft handles the AutoCraftRecipe request action.
func (h *ItemStackRequestHandler) handleAutoCraft(a *protocol.AutoCraftRecipeStackRequestAction, s *Session, tx *world.Tx) error {
	craft, ok := s.recipes[a.RecipeNetworkID]
	if !ok {
		// Try dynamic recipes if no static recipe matches
		return h.tryDynamicCraft(s, tx, int(a.TimesCrafted))
	}
	_, shaped := craft.(recipe.Shaped)
	_, shapeless := craft.(recipe.Shapeless)
	if !shaped && !shapeless {
		return fmt.Errorf("recipe with network id %v is not a shaped or shapeless recipe", a.RecipeNetworkID)
	}
	if craft.Block() != "crafting_table" {
		return fmt.Errorf("recipe with network id %v is not a crafting table recipe", a.RecipeNetworkID)
	}

	timesCrafted := int(a.TimesCrafted)
	if timesCrafted < 1 {
		return fmt.Errorf("times crafted must be at least 1")
	}

	flattenedInputs := make([]recipe.Item, 0, len(craft.Input()))
	for _, i := range craft.Input() {
		if i.Empty() {
			// We don't actually need this item - it's empty, so avoid putting it in our flattened inputs.
			continue
		}

		if ind := slices.IndexFunc(flattenedInputs, func(it recipe.Item) bool {
			return matchingStacks(it, i)
		}); ind >= 0 {
			flattenedInputs[ind] = grow(i, flattenedInputs[ind].Count())
			continue
		}
		flattenedInputs = append(flattenedInputs, i)
	}

	for _, expected := range flattenedInputs {
		remaining := expected.Count() * timesCrafted

		for id, inv := range map[byte]*inventory.Inventory{
			protocol.ContainerCraftingInput:              s.ui,
			protocol.ContainerCombinedHotBarAndInventory: s.inv,
		} {
			for slot, has := range inv.Slots() {
				if has.Empty() {
					// We don't have this item, skip it.
					continue
				}
				if !matchingStacks(has, expected) {
					// Not the same item.
					continue
				}

				removal := has.Count()
				if remaining < removal {
					removal = remaining
				}
				remaining -= removal

				has = has.Grow(-removal)
				h.setItemInSlot(protocol.StackRequestSlotInfo{
					Container: protocol.FullContainerName{ContainerID: id},
					Slot:      byte(slot),
				}, has, s, tx)
				if remaining == 0 {
					// Consumed this item, so go to the next one.
					break
				}
			}
			if remaining == 0 {
				// Consumed this item, so go to the next one.
				break
			}
		}
		if remaining != 0 {
			return fmt.Errorf("recipe %v: could not consume expected item: %v", a.RecipeNetworkID, expected)
		}
	}

	return h.createResults(s, tx, repeatStacks(craft.Output(), timesCrafted)...)
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

// craftingSize gets the crafting size based on the opened container ID.
func (s *Session) craftingSize() uint32 {
	if s.openedContainerID.Load() == 1 {
		return craftingGridSizeLarge
	}
	return craftingGridSizeSmall
}

// craftingOffset gets the crafting offset based on the opened container ID.
func (s *Session) craftingOffset() uint32 {
	if s.openedContainerID.Load() == 1 {
		return craftingGridLargeOffset
	}
	return craftingGridSmallOffset
}

// matchingStacks returns true if the two stacks are the same in a crafting scenario.
func matchingStacks(has, expected recipe.Item) bool {
	switch expected := expected.(type) {
	case item.Stack:
		switch has := has.(type) {
		case recipe.ItemTag:
			name, _ := expected.Item().EncodeItem()
			return has.Contains(name)
		case item.Stack:
			_, variants := expected.Value("variants")
			if !variants {
				return has.Comparable(expected)
			}
			nameOne, _ := has.Item().EncodeItem()
			nameTwo, _ := expected.Item().EncodeItem()
			return nameOne == nameTwo
		}
		panic(fmt.Errorf("client has unexpected recipe item %T", has))
	case recipe.ItemTag:
		switch has := has.(type) {
		case item.Stack:
			name, _ := has.Item().EncodeItem()
			return expected.Contains(name)
		case recipe.ItemTag:
			return has.Tag() == expected.Tag()
		}
		panic(fmt.Errorf("client has unexpected recipe item %T", has))
	}
	panic(fmt.Errorf("tried to match with unexpected recipe item %T", expected))
}

// repeatStacks multiplies the count of all item stacks provided by the number of repetitions provided. Item
// stacks where the new count would exceed the item's max count are split into multiple item stacks.
func repeatStacks(items []item.Stack, repetitions int) []item.Stack {
	output := make([]item.Stack, 0, len(items))
	for _, o := range items {
		count, maxCount := o.Count(), o.MaxCount()
		total := count * repetitions

		stacks := int(math.Ceil(float64(total) / float64(maxCount)))
		for i := 0; i < stacks; i++ {
			inc := min(total, maxCount)
			total -= inc

			output = append(output, o.Grow(inc-count))
		}
	}
	return output
}

func grow(i recipe.Item, count int) recipe.Item {
	switch i := i.(type) {
	case item.Stack:
		return i.Grow(count)
	case recipe.ItemTag:
		return recipe.NewItemTag(i.Tag(), i.Count()+count)
	}
	panic(fmt.Errorf("unexpected recipe item %T", i))
}

// tryDynamicCraft attempts to match the items in the crafting grid with any registered dynamic recipes.
func (h *ItemStackRequestHandler) tryDynamicCraft(s *Session, tx *world.Tx, timesCrafted int) error {
	if timesCrafted < 1 {
		return fmt.Errorf("times crafted must be at least 1")
	}

	size := s.craftingSize()
	offset := s.craftingOffset()

	// Collect all items from the crafting grid
	input := make([]recipe.Item, size)
	for i := uint32(0); i < size; i++ {
		slot := offset + i
		it, _ := s.ui.Item(int(slot))
		if it.Empty() {
			input[i] = item.Stack{}
		} else {
			input[i] = it
		}
	}

	// Try to match with any dynamic recipe
	for _, dynamicRecipe := range recipe.DynamicRecipes() {
		if dynamicRecipe.Block() != "crafting_table" {
			continue
		}

		output, ok := dynamicRecipe.Match(input)
		if !ok {
			continue
		}

		// Found a matching dynamic recipe! Now validate ingredient counts and consume the items
		// For dynamic recipes, we consume all non-empty slots, but we need to ensure each slot
		// has enough items to craft timesCrafted times.
		minStackCount := math.MaxInt
		for i := uint32(0); i < size; i++ {
			slot := offset + i
			it, _ := s.ui.Item(int(slot))
			if !it.Empty() {
				if it.Count() < minStackCount {
					minStackCount = it.Count()
				}
			}
		}

		// Cap timesCrafted to the minimum available stack count to prevent item duplication
		if minStackCount < timesCrafted {
			timesCrafted = minStackCount
		}

		// Now consume the validated amount from each non-empty slot
		for i := uint32(0); i < size; i++ {
			slot := offset + i
			it, _ := s.ui.Item(int(slot))
			if !it.Empty() {
				// Consume one item from this slot per craft
				st := it.Grow(-1 * timesCrafted)
				h.setItemInSlot(protocol.StackRequestSlotInfo{
					Container: protocol.FullContainerName{ContainerID: protocol.ContainerCraftingInput},
					Slot:      byte(slot),
				}, st, s, tx)
			}
		}

		return h.createResults(s, tx, repeatStacks(output, timesCrafted)...)
	}

	return fmt.Errorf("no matching recipe found for crafting grid")
}

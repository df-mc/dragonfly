package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// stonecutterInputSlot is the slot index of the input item in the stonecutter.
const stonecutterInputSlot = 0x03

// handleStonecutting handles a CraftRecipe stack request action made using a stonecutter.
func (h *ItemStackRequestHandler) handleStonecutting(a *protocol.CraftRecipeStackRequestAction, s *Session) error {
	craft, ok := s.recipes[a.RecipeNetworkID]
	if !ok {
		return fmt.Errorf("recipe with network id %v does not exist", a.RecipeNetworkID)
	}
	if _, shapeless := craft.(recipe.Shapeless); !shapeless {
		return fmt.Errorf("recipe with network id %v is not a shapeless recipe", a.RecipeNetworkID)
	}
	if craft.Block() != "stonecutter" {
		return fmt.Errorf("recipe with network id %v is not a stonecutter recipe", a.RecipeNetworkID)
	}

	expectedInputs := craft.Input()
	input, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerStonecutterInput,
		Slot:        stonecutterInputSlot,
	}, s)
	if !matchingStacks(input, expectedInputs[0]) {
		return fmt.Errorf("input item is not the same as expected input")
	}

	output := craft.Output()
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerStonecutterInput,
		Slot:        stonecutterInputSlot,
	}, input.Grow(-1), s)
	return h.createResults(s, output...)
}

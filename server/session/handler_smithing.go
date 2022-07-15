package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	// smithingInputSlot is the slot index of the input item in the smithing table.
	smithingInputSlot = 0x33
	// smithingMaterialSlot is the slot index of the material in the smithing table.
	smithingMaterialSlot = 0x34
)

// handleSmithing handles a CraftRecipe stack request action made using a smithing table.
func (h *ItemStackRequestHandler) handleSmithing(a *protocol.CraftRecipeStackRequestAction, s *Session) error {
	craft, ok := s.recipes[a.RecipeNetworkID]
	if !ok {
		return fmt.Errorf("recipe with network id %v does not exist", a.RecipeNetworkID)
	}
	if _, shapeless := craft.(recipe.Shapeless); !shapeless {
		return fmt.Errorf("recipe with network id %v is not a shaped or shapeless recipe", a.RecipeNetworkID)
	}
	if _, ok := craft.Block().(block.SmithingTable); !ok {
		return fmt.Errorf("recipe with network id %v is not a smithing table recipe", a.RecipeNetworkID)
	}

	expectedInputs := craft.Input()
	input, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerSmithingInput,
		Slot:        smithingInputSlot,
	}, s)
	if !matchingStacks(input, expectedInputs[0]) {
		return fmt.Errorf("input item is not the same as expected input")
	}
	material, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerSmithingMaterial,
		Slot:        smithingMaterialSlot,
	}, s)
	if !matchingStacks(material, expectedInputs[1]) {
		return fmt.Errorf("material item is not the same as expected material")
	}

	output := craft.Output()
	outputStack := item.NewStack(output[0].Item(), input.Count()).
		WithDurability(input.Durability()).
		WithCustomName(input.CustomName()).
		WithLore(input.Lore()...).
		WithEnchantments(input.Enchantments()...).
		WithAnvilCost(input.AnvilCost())
	for k, v := range input.Values() {
		outputStack = outputStack.WithValue(k, v)
	}
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerSmithingInput,
		Slot:        smithingInputSlot,
	}, input.Grow(-1), s)
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerSmithingMaterial,
		Slot:        smithingMaterialSlot,
	}, material.Grow(-1), s)
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerOutput,
		Slot:        craftingResult,
	}, outputStack, s)
	return nil
}

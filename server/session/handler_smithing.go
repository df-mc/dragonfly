package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	// smithingInputSlot is the slot index of the input item in the smithing table.
	smithingInputSlot = 0x33
	// smithingMaterialSlot is the slot index of the material in the smithing table.
	smithingMaterialSlot = 0x34
	// smithingTemplateSlot is the slot index of the template item in the smithing table.
	smithingTemplateSlot = 0x35
)

// handleSmithing handles a CraftRecipe stack request action made using a smithing table.
func (h *ItemStackRequestHandler) handleSmithing(a *protocol.CraftRecipeStackRequestAction, s *Session) error {
	// First, check the recipe and ensure it is valid for the smithing table.
	craft, ok := s.recipes[a.RecipeNetworkID]
	if !ok {
		return fmt.Errorf("recipe with network id %v does not exist", a.RecipeNetworkID)
	}
	if craft.Block() != "smithing_table" {
		return fmt.Errorf("recipe with network id %v is not a smithing table recipe", a.RecipeNetworkID)
	}
	switch craft.(type) {
	case recipe.SmithingTransform, recipe.SmithingTrim:
	default:
		return fmt.Errorf("recipe with network id %v is not a smithing recipe", a.RecipeNetworkID)
	}

	// Check if the input item and material item match what the recipe requires.
	expectedInputs := craft.Input()
	input, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerSmithingTableInput},
		Slot:      smithingInputSlot,
	}, s)
	if !matchingStacks(input, expectedInputs[0]) {
		return fmt.Errorf("input item is not the same as expected input")
	}
	material, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerSmithingTableMaterial},
		Slot:      smithingMaterialSlot,
	}, s)
	if !matchingStacks(material, expectedInputs[1]) {
		return fmt.Errorf("material item is not the same as expected material")
	}
	template, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerSmithingTableTemplate},
		Slot:      smithingTemplateSlot,
	}, s)
	if !matchingStacks(template, expectedInputs[2]) {
		return fmt.Errorf("template item is not the same as expected template")
	}

	// Create the output using the input stack as reference and the recipe's output item type.
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerSmithingTableInput},
		Slot:      smithingInputSlot,
	}, input.Grow(-1), s)
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerSmithingTableMaterial},
		Slot:      smithingMaterialSlot,
	}, material.Grow(-1), s)
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerSmithingTableTemplate},
		Slot:      smithingTemplateSlot,
	}, template.Grow(-1), s)

	if _, ok = craft.(recipe.SmithingTrim); ok {
		var trim item.ArmourTrim
		if t, ok := template.Item().(item.SmithingTemplate); ok {
			trim.Template = t.Template
		} else {
			return fmt.Errorf("template item is not a smithing template")
		}
		if trim.Material, ok = material.Item().(item.ArmourTrimMaterial); !ok {
			return fmt.Errorf("material item is not an armour trim material")
		}
		return h.createResults(s, duplicateStack(input.WithArmourTrim(trim), input.Item()))
	}
	return h.createResults(s, duplicateStack(input, craft.Output()[0].Item()))
}

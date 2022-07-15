package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math/rand"
)

// handleCraftRecipeOptional handles the CraftRecipeOptional request action, sent when taking a result from an anvil
// menu. It also contains information such as the new name of the item and the multi-recipe network ID.
func (h *ItemStackRequestHandler) handleCraftRecipeOptional(a *protocol.CraftRecipeOptionalStackRequestAction, s *Session, filterStrings []string) (err error) {
	// First check if there actually is an anvil opened.
	if !s.containerOpened.Load() {
		return fmt.Errorf("no anvil container opened")
	}

	w := s.c.World()
	pos := s.openedPos.Load()
	anvil, ok := w.Block(pos).(block.Anvil)
	if !ok {
		return fmt.Errorf("no anvil container opened")
	}
	if len(filterStrings) < int(a.FilterStringIndex) {
		return fmt.Errorf("filter string index %v is out of bounds", a.FilterStringIndex)
	}

	input, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerAnvilInput,
		Slot:        0x1,
	}, s)
	if input.Empty() {
		return fmt.Errorf("no item in input input slot")
	}
	material, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerAnvilMaterial,
		Slot:        0x2,
	}, s)
	result := input

	// The sum of the input's anvil cost as well as the material's anvil cost.
	anvilCost := input.AnvilCost()
	if !material.Empty() {
		anvilCost += material.AnvilCost()
	}

	// The material input may be empty (if the player is only renaming, for example).
	var actionCost, renameCost, repairCount int
	if !material.Empty() {
		// First check if we are trying to repair the item with a material.
		if repairable, ok := input.Item().(item.Repairable); ok && repairable.RepairableBy(material) {
			result, actionCost, repairCount, err = repairWithMaterial(input, material, result)
			if err != nil {
				return err
			}
		} else {
			_, book := material.Item().(item.EnchantedBook)
			_, durable := input.Item().(item.Durable)

			// Ensure that the input item is repairable, or the material item is an enchanted book. If not, this is an
			// invalid scenario, and we should return an error.
			enchantedBook := book && len(material.Enchantments()) > 0
			if !enchantedBook && (input.Item() != material.Item() || !durable) {
				return fmt.Errorf("input item is not repairable/same type or material item is not an enchanted book")
			}

			// If the material is another durable item, we just need to increase the durability of the result by the
			// material's durability at 12%.
			if durable && !enchantedBook {
				result, actionCost = repairWithDurable(input, material, result)
			}

			// Merge enchantments on the material item onto the result item.
			var hasCompatible, hasIncompatible bool
			result, hasCompatible, hasIncompatible, actionCost = mergeEnchantments(input, material, result, actionCost, enchantedBook)

			// If we don't have any compatible enchantments and the input item isn't durable, then this is an invalid
			// scenario, and we should return an error.
			if !durable && hasIncompatible && !hasCompatible {
				return fmt.Errorf("no compatible enchantments but have incompatible ones")
			}
		}
	}

	// First get the new name and the existing name. The existing name is either the custom name if it exists, or the
	// item's display name in-game, which is locale dependent.
	newName := filterStrings[int(a.FilterStringIndex)]
	existingName := item.DisplayName(input.Item(), s.c.Locale())
	if customName := input.CustomName(); len(customName) > 0 {
		existingName = customName
	}

	// If our existing name isn't the same as the new name, then something changed, and we should update the custom
	// name of the item.
	if existingName != newName {
		renameCost = 1
		actionCost += renameCost
		result = result.WithCustomName(newName)
	}

	// Calculate the total cost. (action cost + anvil cost)
	cost := actionCost + anvilCost
	if cost <= 0 {
		return fmt.Errorf("no action was taken")
	}

	// If our only action was renaming, the cost should never exceed 40.
	if renameCost == actionCost && renameCost > 0 && cost >= 40 {
		cost = 39
	}

	// We can bypass the "impossible cost" limit if we're in creative mode.
	c := s.c.GameMode().CreativeInventory()
	if cost >= 40 && !c {
		return fmt.Errorf("impossible cost")
	}

	// Ensure we have enough levels (or if we're in creative mode, ignore the cost) to perform the action.
	level := s.c.ExperienceLevel()
	if level < cost && !c {
		return fmt.Errorf("not enough experience")
	} else if !c {
		s.c.SetExperienceLevel(level - cost)
	}

	// If we had a result item, we need to calculate the new anvil cost and update it on the item.
	if !result.Empty() {
		updatedAnvilCost := result.AnvilCost()
		if !material.Empty() && updatedAnvilCost < material.AnvilCost() {
			updatedAnvilCost = material.AnvilCost()
		}
		if renameCost != actionCost || renameCost == 0 {
			updatedAnvilCost = updatedAnvilCost*2 + 1
		}
		result = result.WithAnvilCost(updatedAnvilCost)
	}

	// If we're not in creative mode, we have a 12% chance of the anvil degrading down one state. If that is the case, we
	// need to play the related sound and update the block state. Otherwise, we play a regular anvil use sound.
	if !c && rand.Float64() < 0.12 {
		damaged := anvil.Break()
		if _, ok := damaged.(block.Air); ok {
			w.PlaySound(pos.Vec3Centre(), sound.AnvilBreak{})
		} else {
			w.PlaySound(pos.Vec3Centre(), sound.AnvilUse{})
		}
		defer w.SetBlock(pos, damaged, nil)
	} else {
		w.PlaySound(pos.Vec3Centre(), sound.AnvilUse{})
	}

	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerAnvilInput,
		Slot:        0x1,
	}, item.Stack{}, s)
	if repairCount > 0 {
		h.setItemInSlot(protocol.StackRequestSlotInfo{
			ContainerID: containerAnvilMaterial,
			Slot:        0x2,
		}, material.Grow(-repairCount), s)
	} else {
		h.setItemInSlot(protocol.StackRequestSlotInfo{
			ContainerID: containerAnvilMaterial,
			Slot:        0x2,
		}, item.Stack{}, s)
	}
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerOutput,
		Slot:        craftingResult,
	}, result, s)
	return nil
}

// repairWithMaterial is a helper function that repairs an item stack with a given material stack. It returns the new item
// stack, the cost, and the repaired items count.
func repairWithMaterial(input item.Stack, material item.Stack, result item.Stack) (item.Stack, int, int, error) {
	// Calculate the durability delta using the maximum durability and the current durability.
	delta := min(input.MaxDurability()-input.Durability(), input.MaxDurability()/4)
	if delta <= 0 {
		return item.Stack{}, 0, 0, fmt.Errorf("input item is already fully repaired")
	}

	// While the durability delta is more than zero and the repaired count is under the material count, increase
	// the durability of the result by the durability delta.
	var cost, count int
	for ; delta > 0 && count < material.Count(); count, delta = count+1, min(result.MaxDurability()-result.Durability(), result.MaxDurability()/4) {
		result = result.WithDurability(result.Durability() + delta)
		cost++
	}
	return result, cost, count, nil
}

// repairWithDurable is a helper function that repairs an item with another durable item stack.
func repairWithDurable(input item.Stack, durable item.Stack, result item.Stack) (item.Stack, int) {
	durability := input.Durability() + durable.Durability() + input.MaxDurability()*12/100
	if durability > input.MaxDurability() {
		durability = input.MaxDurability()
	}

	// Ensure the durability is higher than the input's current durability.
	var cost int
	if durability > input.Durability() {
		result = result.WithDurability(durability)
		cost += 2
	}
	return result, cost
}

// mergeEnchantments merges the enchantments of the material item stack onto the result item stack and returns the result
// item stack, booleans indicating whether the enchantments had any compatible or incompatible enchantments, and the cost.
func mergeEnchantments(input item.Stack, material item.Stack, result item.Stack, cost int, enchantedBook bool) (item.Stack, bool, bool, int) {
	var hasCompatible, hasIncompatible bool
	for _, enchant := range material.Enchantments() {
		// First ensure that the enchantment type is compatible with the input item.
		enchantType := enchant.Type()
		compatible := enchantType.CompatibleWithItem(input.Item())
		if _, ok := input.Item().(item.EnchantedBook); ok {
			compatible = true
		}

		// Then ensure that each input enchantment is compatible with this material enchantment. If one is not compatible,
		// increase the cost by one.
		for _, otherEnchant := range input.Enchantments() {
			if otherType := otherEnchant.Type(); enchantType != otherType && !enchantType.CompatibleWithEnchantment(otherType) {
				compatible = false
				cost++
			}
		}

		// Skip the enchantment if it isn't compatible with enchantments on the input item.
		if !compatible {
			hasIncompatible = true
			continue
		}
		hasCompatible = true

		resultLevel := enchant.Level()
		levelCost := resultLevel

		// Check if we have an enchantment of the same type on the input item.
		if existingEnchant, ok := input.Enchantment(enchantType); ok {
			if existingEnchant.Level() > resultLevel || (existingEnchant.Level() == resultLevel && resultLevel == enchantType.MaxLevel()) {
				// The result level is either lower than the existing enchantment's level or is higher than the maximum
				// level, so skip this enchantment.
				hasIncompatible = true
				continue
			} else if existingEnchant.Level() == resultLevel {
				// If the input level is equal to the material level, increase the result level by one.
				resultLevel++
			}
			// Update the level cost. (result level - existing level)
			levelCost = resultLevel - existingEnchant.Level()
		}

		// Now calculate the rarity cost. This is just the application cost of the rarity, however if the
		// material is an enchanted book, then the rarity cost gets halved. If the new rarity cost is under one,
		// it is set to one.
		rarityCost := enchantType.Rarity().Cost()
		if enchantedBook {
			rarityCost = max(1, rarityCost/2)
		}

		// Update the result item with the new enchantment.
		result = result.WithEnchantments(item.NewEnchantment(enchantType, resultLevel))

		// Update the cost appropriately.
		cost += rarityCost * levelCost
		if input.Count() > 1 {
			cost = 40
		}
	}
	return result, hasCompatible, hasIncompatible, cost
}

// max returns the max of two integers.
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// min returns the min of two integers.
func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

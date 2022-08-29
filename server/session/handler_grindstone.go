package session

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	// grindstoneFirstInputSlot is the slot index of the first input item in the grindstone.
	grindstoneFirstInputSlot = 0x10
	// grindstoneSecondInputSlot is the slot index of the second input item in the grindstone.
	grindstoneSecondInputSlot = 0x11
)

// handleGrindstoneCraft handles a CraftGrindstoneRecipe stack request action made using a grindstone.
func (h *ItemStackRequestHandler) handleGrindstoneCraft(s *Session) error {
	// First check if there actually is a grindstone opened.
	if !s.containerOpened.Load() {
		return fmt.Errorf("no grindstone container opened")
	}
	if _, ok := s.c.World().Block(s.openedPos.Load()).(block.Grindstone); !ok {
		return fmt.Errorf("no grindstone container opened")
	}

	// Next, get both input items and ensure they are comparable.
	firstInput, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerGrindstoneFirstInput,
		Slot:        grindstoneFirstInputSlot,
	}, s)
	secondInput, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerGrindstoneSecondInput,
		Slot:        grindstoneSecondInputSlot,
	}, s)
	if firstInput.Empty() && secondInput.Empty() {
		return fmt.Errorf("input item(s) are empty")
	}
	if firstInput.Count() > 1 || secondInput.Count() > 1 {
		return fmt.Errorf("input item(s) are not single items")
	}

	resultStack := nonZeroItem(firstInput, secondInput)
	if !firstInput.Empty() && !secondInput.Empty() {
		// We add the enchantments to the result stack in order to calculate the gained experience. These enchantments
		// are stripped when creating the result.
		resultStack = firstInput.WithEnchantments(secondInput.Enchantments()...)

		// Merge the durability of the two input items at 5%.
		maxDurability := firstInput.MaxDurability()
		firstDurability, secondDurability := firstInput.Durability(), secondInput.Durability()

		resultStack = resultStack.WithDurability(firstDurability + secondDurability + maxDurability*5/100)
	}

	w := s.c.World()
	for _, o := range entity.NewExperienceOrbs(entity.EyePosition(s.c), experienceFromEnchantments(resultStack)) {
		o.SetVelocity(mgl64.Vec3{(rand.Float64()*0.2 - 0.1) * 2, rand.Float64() * 0.4, (rand.Float64()*0.2 - 0.1) * 2})
		w.AddEntity(o)
	}

	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerGrindstoneFirstInput,
		Slot:        grindstoneFirstInputSlot,
	}, item.Stack{}, s)
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		ContainerID: containerGrindstoneSecondInput,
		Slot:        grindstoneSecondInputSlot,
	}, item.Stack{}, s)
	return h.createResults(s, stripPossibleEnchantments(resultStack))
}

// curseEnchantment represents an enchantment that may be a curse enchantment.
type curseEnchantment interface {
	Curse() bool
}

// experienceFromEnchantments returns the amount of experience that is gained from the enchantments on the given stack.
func experienceFromEnchantments(stack item.Stack) int {
	var totalCost int
	for _, enchant := range stack.Enchantments() {
		if _, ok := enchant.Type().(curseEnchantment); ok {
			continue
		}
		cost, _ := enchant.Type().Cost(enchant.Level())
		totalCost += cost
	}
	if totalCost == 0 {
		// No cost, no experience.
		return 0
	}

	minExperience := int(math.Ceil(float64(totalCost) / 2))
	return minExperience + rand.Intn(minExperience)
}

// stripPossibleEnchantments strips all enchantments possible, excluding curses.
func stripPossibleEnchantments(stack item.Stack) item.Stack {
	for _, enchant := range stack.Enchantments() {
		if _, ok := enchant.Type().(curseEnchantment); ok {
			continue
		}
		stack = stack.WithoutEnchantments(enchant.Type())
	}
	return stack
}

// nonZeroItem returns the item.Stack that exists out of two input items. The function expects at least one of the
// items to be non-empty.
func nonZeroItem(first, second item.Stack) item.Stack {
	if first.Empty() {
		return second
	}
	return first
}

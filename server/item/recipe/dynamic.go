package recipe

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// DecoratedPotRecipe is a dynamic recipe for crafting decorated pots. The output depends on which
// pottery sherds or bricks are used in the crafting grid.
type DecoratedPotRecipe struct {
	block string
}

// NewDecoratedPotRecipe creates a new decorated pot recipe.
func NewDecoratedPotRecipe() DecoratedPotRecipe {
	return DecoratedPotRecipe{block: "crafting_table"}
}

// potDecoration is a local interface to check if an item can be used as a pot decoration
// without importing the block package (which would create an import cycle).
type potDecoration interface {
	world.Item
	PotDecoration() bool
}

// Match checks if the given input items match the decorated pot recipe pattern.
// The pattern requires exactly 4 PotDecoration items (bricks or pottery sherds) in a diamond/plus shape:
// - Slot 1 (top centre)
// - Slot 3 (middle left)
// - Slot 5 (middle right)
// - Slot 7 (bottom centre)
// All other slots must be empty.
func (r DecoratedPotRecipe) Match(input []Item) (output []item.Stack, ok bool) {
	// For a 3x3 crafting grid, we need exactly 9 slots
	if len(input) != 9 {
		return nil, false
	}

	// Define the slots for the diamond pattern (0-indexed)
	// Layout:  0 1 2
	//          3 4 5
	//          6 7 8
	// We need items at: 1 (top), 3 (left), 5 (right), 7 (bottom)
	// Odd indices should have items, even indices should be empty

	decorations := [4]world.Item{}
	decorationIndex := 0
	for i := range input {
		it := input[i]
		if i%2 == 0 {
			// Even slots (0, 2, 4, 6, 8) should be empty
			if !it.Empty() {
				return nil, false
			}
		} else {
			// Odd slots (1, 3, 5, 7) should have items
			if it.Empty() {
				return nil, false
			}

			// Extract the actual item from the Item interface
			var actualItem item.Stack
			if v, ok := it.(item.Stack); ok {
				actualItem = v
			} else {
				// ItemTag or other types are not valid for decorated pots
				return nil, false
			}

			// Check if the item implements PotDecoration
			decoration, ok := actualItem.Item().(potDecoration)
			if !ok {
				return nil, false
			}
			decorations[decorationIndex] = decoration
			decorationIndex++
		}
	}

	// Create the decorated pot by encoding the decorations into NBT
	// We'll use world.BlockByName to get the DecoratedPot block and set its decorations
	// The decorations are ordered: [top, left, right, bottom] in the crafting grid
	// For the pot NBT: [back, left, front, right] based on facing direction

	// Get a decorated pot block instance
	pot, ok := world.BlockByName("minecraft:decorated_pot", map[string]any{"direction": int32(2)})
	if !ok {
		return nil, false
	}

	// The pot will be decoded with the decorations through NBT when placed
	// For now, we'll create a pot with the decorations in the correct order
	// DecoratedPot.DecodeNBT expects sherds in order: [back, left, front, right]
	sherds := []any{}
	// Order: top -> back, left -> left, bottom -> front, right -> right
	for _, idx := range []int{0, 1, 3, 2} { // top, left, bottom, right
		name, _ := decorations[idx].EncodeItem()
		sherds = append(sherds, name)
	}

	// Decode the pot with the sherds NBT data using type assertion
	if nbtDecoder, ok := pot.(interface {
		DecodeNBT(map[string]any) any
	}); ok {
		decodedPot := nbtDecoder.DecodeNBT(map[string]any{
			"id":     "DecoratedPot",
			"sherds": sherds,
		})
		if potItem, ok := decodedPot.(world.Item); ok {
			return []item.Stack{item.NewStack(potItem, 1)}, true
		}
	}

	return nil, false
}

// Block returns the block used to craft this recipe.
func (r DecoratedPotRecipe) Block() string {
	return r.block
}

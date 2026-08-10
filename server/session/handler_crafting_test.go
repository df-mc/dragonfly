package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/recipe"
)

// TestMatchingStacksEmpty verifies that an empty slot does not match what a recipe asks for. An empty stack compares
// equal to anything and has no item to read a name from, so matching one either lets a recipe be crafted without it
// or dereferences nil, depending on how the recipe states its input.
func TestMatchingStacksEmpty(t *testing.T) {
	tests := []struct {
		name     string
		expected recipe.Item
	}{
		{name: "an item the recipe names", expected: item.NewStack(item.Diamond{}, 1)},
		{name: "a tag the recipe names", expected: recipe.NewItemTag("minecraft:planks", 1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if matchingStacks(item.Stack{}, tt.expected) {
				t.Error("an empty slot matched what the recipe asks for")
			}
		})
	}
}

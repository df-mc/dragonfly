package recipe

import "github.com/df-mc/dragonfly/server/item"

// Recipe is implemented by all recipe types.
type Recipe interface {
	// Input returns the items required to craft the recipe.
	Input() []InputItem
	// Output returns the items that are produced when the recipe is crafted.
	Output() []item.Stack
	// Block returns the block that is used to craft the recipe.
	Block() string
}

// ShapelessRecipe is a recipe that has no particular shape.
type ShapelessRecipe struct {
	recipe
}

// ShapedRecipe is a recipe that has a specific shape that must be used to craft the output of the recipe.
type ShapedRecipe struct {
	recipe
	// Dimensions are the dimensions for the shaped recipe.
	Dimensions Dimensions
}

// recipe implements the Recipe interface. Structs in this package may embed it to gets its functionality
// out of the box.
type recipe struct {
	// input is a list of items that serve as the input of the shaped recipe. These items are the items
	// required to craft the output. The amount of input items must be exactly equal to Width * Height.
	input []InputItem
	// output contains items that are created as a result of crafting the recipe.
	output []item.Stack
	// block is the block that is used to craft the recipe.
	block string
}

// Input returns the items required to craft the recipe.
func (r recipe) Input() []InputItem {
	return r.input
}

// Output returns the item that is produced when the recipe is crafted.
func (r recipe) Output() []item.Stack {
	return r.output
}

// Block returns the block that is used to craft the recipe.
func (r recipe) Block() string {
	return r.block
}

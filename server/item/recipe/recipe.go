package recipe

import "github.com/df-mc/dragonfly/server/item"

// Recipe is implemented by all recipe types.
type Recipe interface {
	// Inputs returns the items required to craft the recipe.
	Inputs() []InputItem
	// UpdateInputs updates the inputs of the recipe.
	UpdateInputs(inputs []InputItem)
	// Output returns the item that is produced when the recipe is crafted.
	Output() item.Stack
	// UpdateOutput updates the output of the recipe.
	UpdateOutput(output item.Stack)
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

// recipe implements the Recipe interface. Structures in this package may embed it to gets its functionality
// out of the box.
type recipe struct {
	// inputs is a list of items that serve as the input of the shaped recipe. These items are the items
	// required to craft the output. The amount of input items must be exactly equal to Width * Height.
	inputs []InputItem
	// output is an item that is created as a result of crafting the recipe.
	output item.Stack
}

// Inputs ...
func (r *recipe) Inputs() []InputItem {
	return r.inputs
}

// UpdateInputs ...
func (r *recipe) UpdateInputs(inputs []InputItem) {
	r.inputs = inputs
}

// Output ...
func (r *recipe) Output() item.Stack {
	return r.output
}

// UpdateOutput ...
func (r *recipe) UpdateOutput(output item.Stack) {
	r.output = output
}

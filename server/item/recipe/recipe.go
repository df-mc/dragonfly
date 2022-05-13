package recipe

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Recipe is implemented by all recipe types.
type Recipe interface {
	// Input returns the items required to craft the recipe.
	Input() []item.Stack
	// Output returns the items that are produced when the recipe is crafted.
	Output() []item.Stack
	// Block returns the block that is used to craft the recipe.
	Block() world.Block
	// Priority returns the priority of the recipe. Recipes with lower priority are preferred compared to recipes with
	// higher priority.
	Priority() uint32
}

// ShapelessRecipe is a recipe that has no particular shape.
type ShapelessRecipe struct {
	recipe
}

// NewShapelessRecipe creates a new shapeless recipe and returns it.
func NewShapelessRecipe(input []item.Stack, output []item.Stack, block world.Block, priority uint32) ShapelessRecipe {
	return ShapelessRecipe{recipe: recipe{
		input:    input,
		output:   output,
		priority: priority,
		block:    block,
	}}
}

// ShapedRecipe is a recipe that has a specific shape that must be used to craft the output of the recipe.
type ShapedRecipe struct {
	recipe
	// shape contains the width and height of the shaped recipe.
	shape Shape
}

// NewShapedRecipe creates a new shaped recipe and returns it.
func NewShapedRecipe(input []item.Stack, output []item.Stack, block world.Block, priority uint32, shape Shape) ShapedRecipe {
	return ShapedRecipe{
		shape: shape,
		recipe: recipe{
			input:    input,
			output:   output,
			priority: priority,
			block:    block,
		},
	}
}

// Shape returns the shape of the recipe.
func (r ShapedRecipe) Shape() Shape {
	return r.shape
}

// recipe implements the Recipe interface. Structs in this package may embed it to gets its functionality
// out of the box.
type recipe struct {
	// input is a list of items that serve as the input of the shaped recipe. These items are the items
	// required to craft the output. The amount of input items must be exactly equal to Width * Height.
	input []item.Stack
	// output contains items that are created as a result of crafting the recipe.
	output []item.Stack
	// block is the block that is used to craft the recipe.
	block world.Block
	// priority is the priority of the recipe versus others.
	priority uint32
}

// Input ...
func (r recipe) Input() []item.Stack {
	return r.input
}

// Output ...
func (r recipe) Output() []item.Stack {
	return r.output
}

// Block ...
func (r recipe) Block() world.Block {
	return r.block
}

// Priority ...
func (r recipe) Priority() uint32 {
	return r.priority
}

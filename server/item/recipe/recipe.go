package recipe

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Recipe is implemented by all recipe types.
type Recipe interface {
	// Input returns the items required to craft the recipe.
	Input() []Item
	// Output returns the items that are produced when the recipe is crafted.
	Output() []item.Stack
	// Block returns the block that is used to craft the recipe.
	Block() string
	// Priority returns the priority of the recipe. Recipes with lower priority are preferred compared to recipes with
	// higher priority.
	Priority() uint32
}

// DynamicRecipe represents a recipe whose output depends on the specific items used in crafting.
// These recipes are not sent to the client and are validated server-side.
type DynamicRecipe interface {
	// Match checks if the given input items match this dynamic recipe pattern.
	// It returns true if the pattern matches, along with the computed output items.
	Match(input []Item) (output []item.Stack, ok bool)
	// Block returns the block that is used to craft the recipe.
	Block() string
}

// Shapeless is a recipe that has no particular shape.
type Shapeless struct {
	recipe
}

// NewShapeless creates a new shapeless recipe and returns it. The recipe can only be crafted on the block passed in the
// parameters. If the block given a crafting table, the recipe can also be crafted in the 2x2 crafting grid in the
// player's inventory.
func NewShapeless(input []Item, output item.Stack, block string) Shapeless {
	return Shapeless{recipe: recipe{
		input:  input,
		output: []item.Stack{output},
		block:  block,
	}}
}

// SmithingTransform represents a recipe only craftable on a smithing table.
type SmithingTransform struct {
	recipe
}

// NewSmithingTransform creates a new smithing recipe and returns it.
func NewSmithingTransform(base, addition, template Item, output item.Stack, block string) SmithingTransform {
	return SmithingTransform{recipe: recipe{
		input:  []Item{base, addition, template},
		output: []item.Stack{output},
		block:  block,
	}}
}

// SmithingTrim represents a recipe only craftable on a smithing table using an armour trim.
type SmithingTrim struct {
	recipe
}

// NewSmithingTrim creates a new smithing trim recipe and returns it. This is
// almost identical to SmithingTransform except there is no output item.
func NewSmithingTrim(base, addition, template Item, block string) SmithingTrim {
	return SmithingTrim{recipe: recipe{
		input: []Item{base, addition, template},
		block: block,
	}}
}

// Furnace represents a recipe only craftable in a furnace.
type Furnace struct {
	recipe
}

// NewFurnace creates a new furnace recipe and returns it.
func NewFurnace(input Item, output item.Stack, block string) Furnace {
	return Furnace{recipe: recipe{
		input:  []Item{input},
		output: []item.Stack{output},
		block:  block,
	}}
}

// PotionContainerChange is a recipe to convert a potion from one type to another, such as from a drinkable potion to a
// splash potion, or from a splash potion to a lingering potion.
type PotionContainerChange struct {
	recipe
}

// NewPotionContainerChange creates a new potion container change recipe and returns it.
func NewPotionContainerChange(input, output world.Item, reagent item.Stack) PotionContainerChange {
	return PotionContainerChange{recipe: recipe{
		input:  []Item{item.NewStack(input, 1), reagent},
		output: []item.Stack{item.NewStack(output, 1)},
		block:  "brewing_stand",
	}}
}

// Potion is a potion mixing recipe which may be used in the brewing stand.
type Potion struct {
	recipe
}

// NewPotion creates a new potion recipe and returns it.
func NewPotion(input, reagent Item, output item.Stack) Potion {
	return Potion{recipe: recipe{
		input:  []Item{input, reagent},
		output: []item.Stack{output},
		block:  "brewing_stand",
	}}
}

// Shaped is a recipe that has a specific shape that must be used to craft the output of the recipe.
type Shaped struct {
	recipe
	// shape contains the width and height of the shaped recipe.
	shape Shape
}

// NewShaped creates a new shaped recipe and returns it. The recipe can only be crafted on the block passed in the
// parameters. If the block given a crafting table, the recipe can also be crafted in the 2x2 crafting grid in the
// player's inventory. If nil is passed, the block will be autofilled as a crafting table. The inputs must always match
// the width*height of the shape.
func NewShaped(input []Item, output item.Stack, shape Shape, block string) Shaped {
	return Shaped{
		shape: shape,
		recipe: recipe{
			input:  input,
			output: []item.Stack{output},
			block:  block,
		},
	}
}

// Shape returns the shape of the recipe.
func (r Shaped) Shape() Shape {
	return r.shape
}

// recipe implements the Recipe interface. Structs in this package may embed it to gets its functionality
// out of the box.
type recipe struct {
	// input is a list of items that serve as the input of the shaped recipe. These items are the items
	// required to craft the output. The amount of input items must be exactly equal to Width * Height.
	input []Item
	// output contains items that are created as a result of crafting the recipe.
	output []item.Stack
	// block is the block that is used to craft the recipe.
	block string
	// priority is the priority of the recipe versus others.
	priority uint32
}

// Input ...
func (r recipe) Input() []Item {
	return r.input
}

// Output ...
func (r recipe) Output() []item.Stack {
	return r.output
}

// Block ...
func (r recipe) Block() string {
	return r.block
}

// Priority ...
func (r recipe) Priority() uint32 {
	return r.priority
}

package recipe

import (
	_ "embed"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

var (
	//go:embed crafting_data.nbt
	vanillaCraftingData []byte
	//go:embed brewing_data.nbt
	vanillaBrewingData []byte
	//go:embed smithing_data.nbt
	vanillaSmithingData []byte
	//go:embed stonecutter_data.nbt
	vanillaStonecutterData []byte
)

// shapedRecipe is a recipe that must be crafted in a specific shape.
type shapedRecipe struct {
	Input    inputItems  `nbt:"input"`
	Output   outputItems `nbt:"output"`
	Block    string      `nbt:"block"`
	Width    int32       `nbt:"width"`
	Height   int32       `nbt:"height"`
	Priority int32       `nbt:"priority"`
}

// shapelessRecipe is a recipe that may be crafted without a strict shape.
type shapelessRecipe struct {
	Input    inputItems  `nbt:"input"`
	Output   outputItems `nbt:"output"`
	Block    string      `nbt:"block"`
	Priority int32       `nbt:"priority"`
}

// potionContainerChangeRecipe ...
type potionContainerChangeRecipe struct {
	InputItem      string `nbt:"input_item"`
	IngredientItem string `nbt:"ingredient_item"`
	OutputItem     string `nbt:"output_item"`
}

// potionRecipe ...
type potionRecipe struct {
	InputPotion         string `nbt:"input_potion"`
	InputPotionMetadata int16  `nbt:"input_potion_metadata"`

	IngredientItem         string `nbt:"ingredient_item"`
	IngredientItemMetadata int16  `nbt:"ingredient_item_metadata"`

	OutputPotion         string `nbt:"output_potion"`
	OutputPotionMetadata int16  `nbt:"output_potion_metadata"`
}

// registerVanillaRecipes registers all vanilla recipes.
//lint:ignore U1000 This method is explicitly present to be used using compiler directives.
func registerVanillaRecipes() {
	var craftingRecipes struct {
		Shaped    []shapedRecipe    `nbt:"shaped"`
		Shapeless []shapelessRecipe `nbt:"shapeless"`
	}
	if err := nbt.Unmarshal(vanillaCraftingData, &craftingRecipes); err != nil {
		panic(err)
	}
	var brewingRecipes struct {
		ContainerChange []potionContainerChangeRecipe `nbt:"container_change"`
		Regular         []potionRecipe                `nbt:"regular"`
	}
	if err := nbt.Unmarshal(vanillaBrewingData, &brewingRecipes); err != nil {
		panic(err)
	}

	var smithingRecipes []shapelessRecipe
	if err := nbt.Unmarshal(vanillaSmithingData, &smithingRecipes); err != nil {
		panic(err)
	}

	var stonecutterRecipes []shapelessRecipe
	if err := nbt.Unmarshal(vanillaStonecutterData, &stonecutterRecipes); err != nil {
		panic(err)
	}

	for _, s := range append(craftingRecipes.Shapeless, append(smithingRecipes, stonecutterRecipes...)...) {
		input, ok := s.Input.Stacks()
		output, okTwo := s.Output.Stacks()
		if !ok || !okTwo {
			// This can be expected to happen, as some recipes contain blocks or items that aren't currently implemented.
			continue
		}
		Register(Shapeless{recipe{
			input:    input,
			output:   output,
			block:    s.Block,
			priority: uint32(s.Priority),
		}})
	}

	for _, s := range craftingRecipes.Shaped {
		input, ok := s.Input.Stacks()
		output, okTwo := s.Output.Stacks()
		if !ok || !okTwo {
			// This can be expected to happen; refer to the comment above.
			continue
		}
		Register(Shaped{
			shape: Shape{int(s.Width), int(s.Height)},
			recipe: recipe{
				input:    input,
				output:   output,
				block:    s.Block,
				priority: uint32(s.Priority),
			},
		})
	}

	for _, c := range brewingRecipes.ContainerChange {
		input, ok := world.ItemByName(c.InputItem, 0)
		ingredient, okTwo := world.ItemByName(c.IngredientItem, 0)
		output, okThree := world.ItemByName(c.OutputItem, 0)
		if !ok || !okTwo || !okThree {
			// This can be expected to happen; refer to the comment above.
			continue
		}
		Register(NewPotionContainerChange(input, ingredient, output))
	}

	for _, r := range brewingRecipes.Regular {
		input, ok := world.ItemByName(r.InputPotion, r.InputPotionMetadata)
		ingredient, okTwo := world.ItemByName(r.IngredientItem, r.IngredientItemMetadata)
		output, okThree := world.ItemByName(r.OutputPotion, r.OutputPotionMetadata)
		if !ok || !okTwo || !okThree {
			// This can be expected to happen; refer to the comment above.
			continue
		}
		Register(NewPotion(input, ingredient, output))
	}
}

package recipe

import (
	_ "embed"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

var (
	//go:embed crafting_data.nbt
	vanillaCraftingData []byte
	//go:embed smithing_data.nbt
	vanillaSmithingData []byte
	//go:embed smithing_trim_data.nbt
	vanillaSmithingTrimData []byte
	//go:embed furnace_data.nbt
	furnaceData []byte
	//go:embed potion_data.nbt
	vanillaPotionData []byte
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

// furnaceRecipe is a recipe that may be crafted in a furnace.
type furnaceRecipe struct {
	Input  inputItem  `nbt:"input"`
	Output outputItem `nbt:"output"`
	Block  string     `nbt:"block"`
}

// potionRecipe is a recipe that may be crafted in a brewing stand.
type potionRecipe struct {
	Input   inputItem  `nbt:"input"`
	Reagent inputItem  `nbt:"reagent"`
	Output  outputItem `nbt:"output"`
}

// potionContainerChangeRecipe is a recipe that may be crafted in a brewing stand.
type potionContainerChangeRecipe struct {
	Input   string    `nbt:"input"`
	Reagent inputItem `nbt:"reagent"`
	Output  string    `nbt:"output"`
}

// registerVanilla can be called to register all vanilla recipes from the generated data files.
// noinspection GoUnusedFunction
//
//lint:ignore U1000 Function is used through compiler directives.
func registerVanilla() {
	var craftingRecipes struct {
		Shaped    []shapedRecipe    `nbt:"shaped"`
		Shapeless []shapelessRecipe `nbt:"shapeless"`
	}
	if err := nbt.Unmarshal(vanillaCraftingData, &craftingRecipes); err != nil {
		panic(err)
	}

	for _, s := range craftingRecipes.Shapeless {
		input, ok := s.Input.Items()
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
		input, ok := s.Input.Items()
		output, okTwo := s.Output.Stacks()
		if !ok || !okTwo {
			// This can be expected to happen - refer to the comment above.
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

	var smithingRecipes []shapelessRecipe
	if err := nbt.Unmarshal(vanillaSmithingData, &smithingRecipes); err != nil {
		panic(err)
	}

	for _, s := range smithingRecipes {
		input, ok := s.Input.Items()
		output, okTwo := s.Output.Stacks()
		if !ok || !okTwo {
			// This can be expected to happen - refer to the comment above.
			continue
		}
		Register(SmithingTransform{recipe{
			input:    input,
			output:   output,
			block:    s.Block,
			priority: uint32(s.Priority),
		}})
	}

	var smithingTrimRecipes []shapelessRecipe
	if err := nbt.Unmarshal(vanillaSmithingTrimData, &smithingTrimRecipes); err != nil {
		panic(err)
	}

	for _, s := range smithingTrimRecipes {
		input, ok := s.Input.Items()
		if !ok {
			// This can be expected to happen - refer to the comment above.
			continue
		}
		Register(SmithingTrim{recipe{
			input:    input,
			block:    s.Block,
			priority: uint32(s.Priority),
		}})
	}

	var furnaceRecipes []furnaceRecipe
	if err := nbt.Unmarshal(furnaceData, &furnaceRecipes); err != nil {
		panic(err)
	}

	for _, s := range furnaceRecipes {
		input, ok := s.Input.Item()
		output, okTwo := s.Output.Stack()
		if !ok || !okTwo {
			// This can be expected to happen - refer to the comment above.
			continue
		}

		Register(Furnace{recipe{
			input:  []Item{input},
			output: []item.Stack{output},
			block:  s.Block,
		}})
	}

	var potionRecipes struct {
		Potions          []potionRecipe                `nbt:"potions"`
		ContainerChanges []potionContainerChangeRecipe `nbt:"container_changes"`
	}

	if err := nbt.Unmarshal(vanillaPotionData, &potionRecipes); err != nil {
		panic(err)
	}

	for _, r := range potionRecipes.Potions {
		input, ok := r.Input.Item()
		reagent, okTwo := r.Reagent.Item()
		output, okThree := r.Output.Stack()
		if !ok || !okTwo || !okThree {
			// This can be expected to happen - refer to the comment above.
			continue
		}

		Register(Potion{recipe{
			input:  []Item{input, reagent},
			output: []item.Stack{output},
			block:  "brewing_stand",
		}})
	}

	for _, c := range potionRecipes.ContainerChanges {
		input, ok := world.ItemByName(c.Input, 0)
		reagent, okTwo := c.Reagent.Item()
		output, okThree := world.ItemByName(c.Output, 0)
		if !ok || !okTwo || !okThree {
			// This can be expected to happen - refer to the comment above.
			continue
		}

		Register(PotionContainerChange{recipe{
			input:  []Item{item.NewStack(input, 1), reagent},
			output: []item.Stack{item.NewStack(output, 1)},
			block:  "brewing_stand",
		}})
	}

	// Register dynamic recipes
	RegisterDynamic(NewDecoratedPotRecipe())
}

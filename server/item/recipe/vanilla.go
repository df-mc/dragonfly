package recipe

import (
	// Ensure all blocks and items are registered before trying to load vanilla recipes.
	_ "github.com/df-mc/dragonfly/server/block"
	_ "github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"

	_ "embed"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

//go:embed crafting_data.nbt
var vanillaCraftingData []byte

func init() {
	var vanillaRecipes struct {
		Shaped []struct {
			Input    inputItems  `nbt:"input"`
			Output   outputItems `nbt:"output"`
			Block    string      `nbt:"block"`
			Width    int32       `nbt:"width"`
			Height   int32       `nbt:"height"`
			Priority int32       `nbt:"priority"`
		} `nbt:"shaped"`
		Shapeless []struct {
			Input    inputItems  `nbt:"input"`
			Output   outputItems `nbt:"output"`
			Block    string      `nbt:"block"`
			Priority int32       `nbt:"priority"`
		} `nbt:"shapeless"`
	}

	if err := nbt.Unmarshal(vanillaCraftingData, &vanillaRecipes); err != nil {
		panic(err)
	}

	for _, s := range vanillaRecipes.Shapeless {
		input, ok := s.Input.Stacks()
		output, okTwo := s.Output.Stacks()
		if !ok || !okTwo {
			// This can be expected to happen, as some recipes contain blocks or items that aren't currently implemented.
			continue
		}
		b, _ := world.BlockByName("minecraft:"+s.Block, nil)
		Register(Shapeless{recipe{
			input:    input,
			output:   output,
			block:    b,
			priority: uint32(s.Priority),
		}})
	}

	for _, s := range vanillaRecipes.Shaped {
		input, ok := s.Input.Stacks()
		output, okTwo := s.Output.Stacks()
		if !ok || !okTwo {
			// This can be expected to happen - refer to the comment above.
			continue
		}
		b, _ := world.BlockByName("minecraft:"+s.Block, nil)
		Register(Shaped{
			shape: Shape{int(s.Width), int(s.Height)},
			recipe: recipe{
				input:    input,
				output:   output,
				block:    b,
				priority: uint32(s.Priority),
			},
		})
	}
}

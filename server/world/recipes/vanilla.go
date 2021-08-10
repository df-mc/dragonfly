package recipes

import (
	_ "embed"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

//go:embed crafting_data.nbt
var vanillaCraftingData []byte

func init() {
	var vanillaRecipes struct {
		Shaped []struct {
			Input    InputItems `nbt:"input"`
			Output   OutputItem `nbt:"output"`
			Priority int32      `nbt:"priority"`
			Width    int32      `nbt:"width"`
			Height   int32      `nbt:"height"`
		} `nbt:"shaped"`
		Shapeless []struct {
			Input    InputItems `nbt:"input"`
			Output   OutputItem `nbt:"output"`
			Priority int32      `nbt:"priority"`
		} `nbt:"shapeless"`
	}

	if err := nbt.Unmarshal(vanillaCraftingData, &vanillaRecipes); err != nil {
		panic(err)
	}

	for _, s := range vanillaRecipes.Shapeless {
		input, ok := s.Input.ToStacks()
		if !ok {
			continue
		}
		output, ok := s.Output.ToStack()
		if !ok {
			continue
		}
		Register(ShapelessRecipe{
			Inputs:   input,
			Output:   output,
			Priority: s.Priority,
		})
	}

	for _, s := range vanillaRecipes.Shaped {
		input, ok := s.Input.ToStacks()
		if !ok {
			continue
		}
		output, ok := s.Output.ToStack()
		if !ok {
			continue
		}
		Register(ShapedRecipe{
			Inputs:   input,
			Output:   output,
			Priority: s.Priority,
			Dimensions: Dimensions{
				Width:  s.Width,
				Height: s.Height,
			},
		})
	}
}

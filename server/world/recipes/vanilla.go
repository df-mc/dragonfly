package recipes

import (
	"fmt"
	// Insure all blocks and items are registered before trying to load vanilla recipes.
	_ "github.com/df-mc/dragonfly/server/block"
	_ "github.com/df-mc/dragonfly/server/item"

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

	for i, s := range vanillaRecipes.Shapeless {
		fmt.Println("-----------------------------")
		fmt.Println("Crafting Recipe:", i)
		for _, i := range s.Input {
			fmt.Println(i.Name, i.MetadataValue)
		}
		fmt.Println("Results in:")
		fmt.Println(s.Output.Name, s.Output.MetadataValue)
		fmt.Println("-----------------------------")
		input, ok := s.Input.ToStacks()
		if !ok {
			continue
		}
		output, ok := s.Output.ToStack()
		if !ok {
			continue
		}
		fmt.Println("Registered recipe:", i)
		Register(ShapelessRecipe{
			Inputs:   input,
			Output:   output,
			Priority: s.Priority,
		})
	}

	for i, s := range vanillaRecipes.Shaped {
		fmt.Println("-----------------------------")
		fmt.Println("Crafting Recipe:", i)
		for _, i := range s.Input {
			fmt.Println(i.Name, i.MetadataValue)
		}
		fmt.Println("Results in:")
		fmt.Println(s.Output.Name, s.Output.MetadataValue)
		fmt.Println("Dimensions:")
		fmt.Println(s.Width, s.Height)
		fmt.Println("-----------------------------")
		input, ok := s.Input.ToStacks()
		if !ok {
			continue
		}
		output, ok := s.Output.ToStack()
		if !ok {
			continue
		}
		fmt.Println("Registered recipe:", i)
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

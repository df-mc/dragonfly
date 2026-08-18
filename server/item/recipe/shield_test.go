package recipe_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/recipe"
)

func TestShieldDecorationRecipe(t *testing.T) {
	shield := item.NewStack(item.Shield{}, 1).Damage(7)
	banner := block.Banner{
		Colour: item.ColourRed(),
		Patterns: []block.BannerPatternLayer{{
			Type:   block.CreeperBannerPattern(),
			Colour: item.ColourBlack(),
		}},
	}

	output, ok := recipe.NewShieldDecorationRecipe().Match([]recipe.Item{
		item.Stack{}, item.NewStack(banner, 1), shield, item.Stack{},
	})
	if !ok || len(output) != 1 {
		t.Fatalf("shield decoration match = (%#v, %v), want one output", output, ok)
	}
	decorated, ok := output[0].Item().(item.Shield)
	if !ok || decorated.Banner == nil {
		t.Fatalf("output item = %#v, want decorated shield", output[0].Item())
	}
	if decorated.Banner.BaseColour != item.ColourRed() || len(decorated.Banner.Patterns) != 1 ||
		decorated.Banner.Patterns[0].Pattern != "cre" || decorated.Banner.Patterns[0].Colour != item.ColourBlack() {
		t.Fatalf("decorated banner = %#v, want copied red creeper design", decorated.Banner)
	}
	if output[0].Durability() != shield.Durability() {
		t.Fatalf("decorated shield durability = %v, want %v", output[0].Durability(), shield.Durability())
	}
}

func TestShieldDecorationRecipeRejectsInvalidInputs(t *testing.T) {
	r := recipe.NewShieldDecorationRecipe()
	banner := item.NewStack(block.Banner{}, 1)
	decorated := item.NewStack(item.Shield{Banner: &item.ShieldBanner{}}, 1)
	tests := []struct {
		name  string
		input []recipe.Item
	}{
		{name: "missing banner", input: []recipe.Item{item.NewStack(item.Shield{}, 1)}},
		{name: "missing shield", input: []recipe.Item{banner}},
		{name: "already decorated", input: []recipe.Item{banner, decorated}},
		{name: "extra ingredient", input: []recipe.Item{banner, item.NewStack(item.Shield{}, 1), item.NewStack(item.Stick{}, 1)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if output, ok := r.Match(tt.input); ok || output != nil {
				t.Fatalf("Match() = (%#v, %v), want no match", output, ok)
			}
		})
	}
}

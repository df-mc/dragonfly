package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestMultiRecipeCraftsDecoratedShield(t *testing.T) {
	recipe.RegisterDynamic(recipe.NewShieldDecorationRecipe())
	ui := inventory.New(craftingResult+1, nil)
	shield := item.NewStack(item.Shield{}, 1).Damage(7)
	banner := block.Banner{Colour: item.ColourRed()}
	_ = ui.SetItem(craftingGridSmallOffset, shield)
	_ = ui.SetItem(craftingGridSmallOffset+1, item.NewStack(banner, 1))

	shieldRecipe := recipe.NewShieldDecorationRecipe()
	s := &Session{
		ui:      ui,
		recipes: map[uint32]recipe.Recipe{7: recipe.NewMulti(shieldRecipe.UUID())},
	}
	h := &ItemStackRequestHandler{
		changes:         map[byte]map[byte]changeInfo{},
		responseChanges: map[int32]map[*inventory.Inventory]map[byte]responseChange{},
	}
	if err := h.handleCraft(&protocol.CraftRecipeStackRequestAction{
		RecipeNetworkID: 7,
		NumberOfCrafts:  1,
	}, s, nil); err != nil {
		t.Fatalf("handleCraft() error: %v", err)
	}

	for _, slot := range []int{craftingGridSmallOffset, craftingGridSmallOffset + 1} {
		if stack, _ := ui.Item(slot); !stack.Empty() {
			t.Fatalf("crafting input slot %v = %v, want empty", slot, stack)
		}
	}
	result, _ := ui.Item(craftingResult)
	decorated, ok := result.Item().(item.Shield)
	if !ok || decorated.Banner == nil || decorated.Banner.BaseColour != item.ColourRed() {
		t.Fatalf("crafting result = %#v, want red decorated shield", result.Item())
	}
	if result.Durability() != shield.Durability() {
		t.Fatalf("crafted shield durability = %v, want %v", result.Durability(), shield.Durability())
	}
}

func TestMultiRecipeRejectsMismatchedUUID(t *testing.T) {
	recipe.RegisterDynamic(recipe.NewShieldDecorationRecipe())
	ui := inventory.New(craftingResult+1, nil)
	_ = ui.SetItem(craftingGridSmallOffset, item.NewStack(item.Shield{}, 1))
	_ = ui.SetItem(craftingGridSmallOffset+1, item.NewStack(block.Banner{}, 1))

	s := &Session{ui: ui, recipes: map[uint32]recipe.Recipe{7: recipe.NewMulti(uuid.New())}}
	h := &ItemStackRequestHandler{
		changes:         map[byte]map[byte]changeInfo{},
		responseChanges: map[int32]map[*inventory.Inventory]map[byte]responseChange{},
	}
	err := h.handleCraft(&protocol.CraftRecipeStackRequestAction{RecipeNetworkID: 7, NumberOfCrafts: 1}, s, nil)
	if err == nil {
		t.Fatal("handleCraft() succeeded for unrelated multi-recipe UUID")
	}
	for _, slot := range []int{craftingGridSmallOffset, craftingGridSmallOffset + 1} {
		if stack, _ := ui.Item(slot); stack.Empty() {
			t.Fatalf("crafting input slot %v was consumed for unrelated multi-recipe UUID", slot)
		}
	}
}

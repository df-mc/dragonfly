package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

func TestShulkerBoxItemsRegistered(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		wantItemName string
		wantBlock    string
	}{
		{name: "minecraft:shulker_box", wantItemName: "minecraft:shulker_box", wantBlock: "minecraft:undyed_shulker_box"},
		{name: "minecraft:undyed_shulker_box", wantItemName: "minecraft:undyed_shulker_box", wantBlock: "minecraft:undyed_shulker_box"},
		{name: "minecraft:white_shulker_box", wantItemName: "minecraft:white_shulker_box", wantBlock: "minecraft:white_shulker_box"},
		{name: "minecraft:light_gray_shulker_box", wantItemName: "minecraft:light_gray_shulker_box", wantBlock: "minecraft:light_gray_shulker_box"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			it, ok := world.ItemByName(tt.name, 0)
			if !ok {
				t.Fatalf("item %q not registered", tt.name)
			}

			var box ShulkerBox
			switch v := it.(type) {
			case ShulkerBox:
				box = v
			case baseShulkerBoxItem:
				box = v.ShulkerBox
			default:
				t.Fatalf("item %q resolved to %T, want shulker box item", tt.name, it)
			}

			itemName, _ := it.EncodeItem()
			if itemName != tt.wantItemName {
				t.Fatalf("EncodeItem() = %q, want %q", itemName, tt.wantItemName)
			}
			blockName, _ := box.EncodeBlock()
			if blockName != tt.wantBlock {
				t.Fatalf("EncodeBlock() = %q, want %q", blockName, tt.wantBlock)
			}
		})
	}
}

func TestShulkerBoxNBTRoundTrip(t *testing.T) {
	t.Parallel()

	box := NewShulkerBox()
	box.Dyed = true
	box.Colour = item.ColourBlue()
	box.CustomName = "stash"
	box.Facing = cube.FaceNorth
	if err := box.inventory.SetItem(0, item.NewStack(item.Bread{}, 3)); err != nil {
		t.Fatalf("set inventory item: %v", err)
	}

	decoded := box.DecodeNBT(box.EncodeNBT()).(ShulkerBox)
	if !decoded.Dyed {
		t.Fatal("decoded box unexpectedly undyed")
	}
	if decoded.Colour != item.ColourBlue() {
		t.Fatalf("decoded colour = %v, want %v", decoded.Colour, item.ColourBlue())
	}
	if decoded.CustomName != "stash" {
		t.Fatalf("decoded custom name = %q, want %q", decoded.CustomName, "stash")
	}
	if decoded.Facing != cube.FaceNorth {
		t.Fatalf("decoded facing = %v, want %v", decoded.Facing, cube.FaceNorth)
	}

	stack, err := decoded.inventory.Item(0)
	if err != nil {
		t.Fatalf("read inventory item: %v", err)
	}
	if stack.Count() != 3 {
		t.Fatalf("decoded count = %d, want 3", stack.Count())
	}
	if stack.Item() == nil {
		t.Fatal("decoded item was nil")
	}
	name, _ := stack.Item().EncodeItem()
	if name != "minecraft:bread" {
		t.Fatalf("decoded item = %q, want %q", name, "minecraft:bread")
	}
}

func TestShulkerBoxItemAliasDecodeNBTPreservesAlias(t *testing.T) {
	t.Parallel()

	box := baseShulkerBoxItem{}
	decoded := box.DecodeNBT(map[string]any{}).(baseShulkerBoxItem)
	name, _ := decoded.EncodeItem()
	if name != "minecraft:shulker_box" {
		t.Fatalf("decoded alias item name = %q, want %q", name, "minecraft:shulker_box")
	}
}

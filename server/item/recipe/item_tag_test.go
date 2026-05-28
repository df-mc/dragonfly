package recipe

import (
	"slices"
	"testing"
)

func TestItemsInTag(t *testing.T) {
	items := ItemsInTag("minecraft:is_sword")
	if !slices.Contains(items, "minecraft:mace") {
		t.Fatal("expected minecraft:is_sword to contain minecraft:mace")
	}

	items[0] = "minecraft:mutated"
	if NewItemTag("minecraft:is_sword", 1).Contains("minecraft:mutated") {
		t.Fatal("expected returned tag item slice mutation not to affect package state")
	}
}

func TestItemTagsReturnsCopy(t *testing.T) {
	tags := ItemTags()
	items, ok := tags["minecraft:is_sword"]
	if !ok {
		t.Fatal("expected minecraft:is_sword tag")
	}
	if !slices.Contains(items, "minecraft:mace") {
		t.Fatal("expected minecraft:is_sword to contain minecraft:mace")
	}

	tags["minecraft:is_sword"] = nil
	items[0] = "minecraft:mutated"
	if !NewItemTag("minecraft:is_sword", 1).Contains("minecraft:mace") {
		t.Fatal("expected returned tag map and slice mutation not to affect package state")
	}
}

package loot_test

import (
	"math/rand/v2"
	"testing"

	_ "github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	_ "github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/loot"
)

func TestFishingTables(t *testing.T) {
	for _, path := range []string{loot.Fishing, loot.JungleFishing} {
		table, ok := loot.Lookup(path)
		if !ok {
			t.Fatalf("loot table %v not found", path)
		}
		for luck := 0; luck <= 3; luck++ {
			rnd := rand.New(rand.NewPCG(1, uint64(luck)))
			for range 5000 {
				stacks := table.Generate(loot.Context{Luck: float64(luck), Rand: rnd})
				if len(stacks) != 1 || stacks[0].Empty() {
					t.Fatalf("%v with luck %v: expected exactly one item, got %v", path, luck, stacks)
				}
				s := stacks[0]
				if s.MaxDurability() > 0 && (s.Durability() < 1 || s.Durability() > s.MaxDurability()) {
					t.Fatalf("%v: invalid durability %v/%v", s, s.Durability(), s.MaxDurability())
				}
				if _, ok := s.Item().(item.Book); ok {
					t.Fatalf("%v: book was not enchanted", path)
				}
			}
		}
	}
}

func TestFishingTableItems(t *testing.T) {
	table, _ := loot.Lookup(loot.JungleFishing)
	rnd := rand.New(rand.NewPCG(2, 2))
	seen := map[string]bool{}
	for range 50000 {
		for _, s := range table.Generate(loot.Context{Rand: rnd}) {
			name, _ := s.Item().EncodeItem()
			seen[name] = true
			if name == "minecraft:ink_sac" && s.Count() != 10 {
				t.Fatalf("expected 10 ink sacs, got %v", s.Count())
			}
		}
	}
	// Legacy names and data values in the vanilla tables must be upgraded.
	for _, name := range []string{"minecraft:cod", "minecraft:tropical_fish", "minecraft:cocoa_beans", "minecraft:ink_sac", "minecraft:bamboo", "minecraft:enchanted_book"} {
		if !seen[name] {
			t.Errorf("expected %v to be caught at least once", name)
		}
	}
}

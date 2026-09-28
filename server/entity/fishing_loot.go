package entity

import (
	"math"
	"math/rand/v2"
	"slices"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
)

// fishingLootEntry is an entry in one of the fishing loot tables. It has a
// weight that determines how likely it is to be selected.
type fishingLootEntry struct {
	weight int
	stack  func() item.Stack
}

// fishingLootStack returns a function that always returns a stack of the item
// passed with the count passed.
func fishingLootStack(it world.Item, count int) func() item.Stack {
	return func() item.Stack { return item.NewStack(it, count) }
}

var (
	// fishingFish is the loot table for fish, the most common catch.
	fishingFish = []fishingLootEntry{
		{weight: 60, stack: fishingLootStack(item.Cod{}, 1)},
		{weight: 25, stack: fishingLootStack(item.Salmon{}, 1)},
		{weight: 2, stack: fishingLootStack(item.TropicalFish{}, 1)},
		{weight: 13, stack: fishingLootStack(item.Pufferfish{}, 1)},
	}
	// fishingJunk is the loot table for junk.
	// TODO: Add tripwire hooks, and bamboo in jungle biomes.
	fishingJunk = []fishingLootEntry{
		{weight: 17, stack: fishingLootStack(block.LilyPad{}, 1)},
		{weight: 10, stack: func() item.Stack {
			return damagedFishingLoot(item.NewStack(item.Boots{Tier: item.ArmourTierLeather{}}, 1), 0.9)
		}},
		{weight: 10, stack: fishingLootStack(item.Leather{}, 1)},
		{weight: 10, stack: fishingLootStack(item.Bone{}, 1)},
		{weight: 10, stack: fishingLootStack(item.Potion{Type: potion.Water()}, 1)},
		{weight: 5, stack: fishingLootStack(block.String{}, 1)},
		{weight: 2, stack: func() item.Stack {
			return damagedFishingLoot(item.NewStack(item.FishingRod{}, 1), 0.9)
		}},
		{weight: 10, stack: fishingLootStack(item.Bowl{}, 1)},
		{weight: 5, stack: fishingLootStack(item.Stick{}, 1)},
		{weight: 1, stack: fishingLootStack(item.InkSac{}, 10)},
		{weight: 10, stack: fishingLootStack(item.RottenFlesh{}, 1)},
	}
	// fishingTreasure is the loot table for treasure, the rarest catch.
	// TODO: Add name tags and saddles.
	fishingTreasure = []fishingLootEntry{
		{weight: 1, stack: func() item.Stack {
			return enchantFishingLoot(damagedFishingLoot(item.NewStack(item.Bow{}, 1), 0.25), 30)
		}},
		{weight: 1, stack: func() item.Stack {
			return enchantFishingLoot(damagedFishingLoot(item.NewStack(item.FishingRod{}, 1), 0.25), 30)
		}},
		{weight: 1, stack: func() item.Stack {
			return enchantFishingLoot(item.NewStack(item.Book{}, 1), 30)
		}},
		{weight: 1, stack: fishingLootStack(item.NautilusShell{}, 1)},
	}
)

// fishingLoot returns a random item.Stack caught with a fishing rod with the
// luck passed. Every level of luck makes fish and junk less likely to be
// caught, and treasure more likely.
func fishingLoot(luck int) item.Stack {
	tables := []struct {
		weight  int
		entries []fishingLootEntry
	}{
		{weight: max(85-luck, 0), entries: fishingFish},
		{weight: max(10-luck*2, 0), entries: fishingJunk},
		{weight: max(5+luck*2, 0), entries: fishingTreasure},
	}
	total := 0
	for _, t := range tables {
		total += t.weight
	}
	r := rand.IntN(total)
	for _, t := range tables {
		if r -= t.weight; r < 0 {
			return pickFishingLoot(t.entries)
		}
	}
	panic("should never happen")
}

// pickFishingLoot picks a random entry from the entries passed, taking into
// account their weights.
func pickFishingLoot(entries []fishingLootEntry) item.Stack {
	total := 0
	for _, e := range entries {
		total += e.weight
	}
	r := rand.IntN(total)
	for _, e := range entries {
		if r -= e.weight; r < 0 {
			return e.stack()
		}
	}
	panic("should never happen")
}

// damagedFishingLoot returns the item.Stack passed with between 0 and
// maxDamage (a fraction of its maximum durability) of damage.
func damagedFishingLoot(s item.Stack, maxDamage float64) item.Stack {
	maxDurability := s.MaxDurability()
	return s.WithDurability(maxDurability - int(rand.Float64()*maxDamage*float64(maxDurability)))
}

// enchantFishingLoot enchants the item.Stack passed as if it was enchanted in
// an enchanting table with the level passed. Unlike enchanting in an
// enchanting table, treasure enchantments may be selected.
func enchantFishingLoot(s item.Stack, level int) item.Stack {
	it := s.Item()
	enchantable, ok := it.(item.Enchantable)
	if !ok || enchantable.EnchantmentValue() <= 0 {
		return s
	}
	value := enchantable.EnchantmentValue()
	_, book := it.(item.Book)

	cost := level + 1 + rand.IntN(value/4+1) + rand.IntN(value/4+1)
	bonus := (rand.Float64() + rand.Float64() - 1) * 0.15
	cost = max(int(math.Round(float64(cost)+float64(cost)*bonus)), 1)

	var available []item.Enchantment
	for _, t := range item.Enchantments() {
		if !book && !t.CompatibleWithItem(it) {
			continue
		}
		for lvl := t.MaxLevel(); lvl > 0; lvl-- {
			if minCost, maxCost := t.Cost(lvl); cost >= minCost && cost <= maxCost {
				available = append(available, item.NewEnchantment(t, lvl))
				break
			}
		}
	}
	if len(available) == 0 {
		return s
	}

	selected := []item.Enchantment{pickEnchantment(&available)}
	for rand.IntN(50) <= cost {
		last := selected[len(selected)-1].Type()
		available = slices.DeleteFunc(available, func(e item.Enchantment) bool {
			return !last.CompatibleWithEnchantment(e.Type()) || !e.Type().CompatibleWithEnchantment(last)
		})
		if len(available) == 0 {
			break
		}
		selected = append(selected, pickEnchantment(&available))
		cost /= 2
	}
	return s.WithEnchantments(selected...)
}

// pickEnchantment picks a random enchantment from the enchantments passed,
// weighted by rarity, and removes it from the slice.
func pickEnchantment(enchants *[]item.Enchantment) item.Enchantment {
	total := 0
	for _, e := range *enchants {
		total += e.Type().Rarity().Weight()
	}
	r := rand.IntN(total)
	for i, e := range *enchants {
		if r -= e.Type().Rarity().Weight(); r < 0 {
			*enchants = slices.Delete(*enchants, i, i+1)
			return e
		}
	}
	panic("should never happen")
}

package loot

import (
	"math"
	"math/rand/v2"
	"slices"

	"github.com/df-mc/dragonfly/server/item"
)

// treasureEnchantment is an enchantment that may be a treasure enchantment,
// which cannot be obtained through enchanting tables.
type treasureEnchantment interface {
	Treasure() bool
}

// EnchantWithLevels enchants the item.Stack passed as if it was enchanted in
// an enchanting table with the level passed. If treasure is true, treasure
// enchantments such as Mending may be selected as well. Books are turned into
// enchanted books.
func EnchantWithLevels(s item.Stack, level int, treasure bool, rnd *rand.Rand) item.Stack {
	it := s.Item()
	enchantable, ok := it.(item.Enchantable)
	if !ok || enchantable.EnchantmentValue() <= 0 {
		return s
	}
	value := enchantable.EnchantmentValue()
	_, book := it.(item.Book)

	cost := level + 1 + rnd.IntN(value/4+1) + rnd.IntN(value/4+1)
	bonus := (rnd.Float64() + rnd.Float64() - 1) * 0.15
	cost = max(int(math.Round(float64(cost)+float64(cost)*bonus)), 1)

	var available []item.Enchantment
	for _, t := range item.Enchantments() {
		if tr, ok := t.(treasureEnchantment); ok && tr.Treasure() && !treasure {
			continue
		}
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

	selected := []item.Enchantment{pickEnchantment(&available, rnd)}
	for rnd.IntN(50) <= cost {
		last := selected[len(selected)-1].Type()
		available = slices.DeleteFunc(available, func(e item.Enchantment) bool {
			return !last.CompatibleWithEnchantment(e.Type()) || !e.Type().CompatibleWithEnchantment(last)
		})
		if len(available) == 0 {
			break
		}
		selected = append(selected, pickEnchantment(&available, rnd))
		cost /= 2
	}
	return s.WithEnchantments(selected...)
}

// pickEnchantment picks a random enchantment from the enchantments passed,
// weighted by rarity, and removes it from the slice.
func pickEnchantment(enchants *[]item.Enchantment, rnd *rand.Rand) item.Enchantment {
	total := 0
	for _, e := range *enchants {
		total += e.Type().Rarity().Weight()
	}
	r := rnd.IntN(total)
	for i, e := range *enchants {
		if r -= e.Type().Rarity().Weight(); r < 0 {
			*enchants = slices.Delete(*enchants, i, i+1)
			return e
		}
	}
	panic("should never happen")
}

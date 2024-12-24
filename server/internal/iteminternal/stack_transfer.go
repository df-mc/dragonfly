package iteminternal

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// DuplicateStack duplicates an item.Stack with the new item type given.
func DuplicateStack(input item.Stack, i world.Item) item.Stack {
	outputStack := item.NewStack(i, input.Count()).
		Damage(input.MaxDurability() - input.Durability()).
		WithCustomName(input.CustomName()).
		WithLore(input.Lore()...).
		WithEnchantments(input.Enchantments()...).
		WithAnvilCost(input.AnvilCost())
	for k, v := range input.Values() {
		outputStack = outputStack.WithValue(k, v)
	}
	return outputStack
}

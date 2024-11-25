package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// Flame turns your arrows into flaming arrows allowing you to set your targets
// on fire.
var Flame flame

type flame struct{}

// Name ...
func (flame) Name() string {
	return "Flame"
}

// MaxLevel ...
func (flame) MaxLevel() int {
	return 1
}

// Cost ...
func (flame) Cost(int) (int, int) {
	return 20, 50
}

// Rarity ...
func (flame) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// BurnDuration always returns five seconds, no matter the level.
func (flame) BurnDuration() time.Duration {
	return time.Second * 5
}

// CompatibleWithEnchantment ...
func (flame) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (flame) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Bow)
	return ok
}

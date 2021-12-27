package item

import (
	"reflect"
)

// Enchantment represents an enchantment that can be applied to an item. It has methods to get the name,
// get the current level, get the maximum level, get a new instance of the enchantment with a certain level
// and to check if the enchantment is compatible with an item stack.
type Enchantment interface {
	// Name returns the name of the enchantment.
	Name() string
	// Level returns the current level of the enchantment. The best way to use this is
	// by having a struct similar to the enchantment.enchantment struct to store the level.
	Level() int
	// MaxLevel returns the maximum level the enchantment should be able to have.
	MaxLevel() int
	// WithLevel is called to create a new instance of the enchantment using the level passed.
	// This method could also be used to limit the level of the enchantment using MaxLevel.
	WithLevel(level int) Enchantment
	// CompatibleWith is called when an enchantment is added to an item. It can be used to check if
	// the enchantment is compatible with the item stack based on the item type, current enchantments etc.
	CompatibleWith(s Stack) bool
}

// RegisterEnchantment registers an enchantment with the ID passed. Once registered, enchantments may be received
// by instantiating an Enchantment struct (e.g. enchantment.Protection{})
func RegisterEnchantment(id int, enchantment Enchantment) {
	enchantments[id] = enchantment
	enchantmentIds[reflect.TypeOf(enchantment)] = id
}

var (
	enchantments   = map[int]Enchantment{}
	enchantmentIds = map[reflect.Type]int{}
)

// EnchantmentByID attempts to return an enchantment by the ID it was registered with. If found, the enchantment found
// is returned and the bool true.
func EnchantmentByID(id int) (Enchantment, bool) {
	e, ok := enchantments[id]
	return e, ok
}

// EnchantmentID attempts to return the ID the enchantment was registered with. If found, the id is returned and
// the bool true.
func EnchantmentID(e Enchantment) (int, bool) {
	id, ok := enchantmentIds[reflect.TypeOf(e)]
	return id, ok
}

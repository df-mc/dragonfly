package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"maps"
	"slices"
	"sort"
)

// Enchantment is an enchantment that can be applied to a Stack. It holds an EnchantmentType and level that influences
// the power of the enchantment.
type Enchantment struct {
	t   EnchantmentType
	lvl int
}

// NewEnchantment creates and returns an Enchantment with a specific EnchantmentType and level. If the level passed
// exceeds EnchantmentType.MaxLevel, NewEnchantment panics.
func NewEnchantment(t EnchantmentType, lvl int) Enchantment {
	if lvl < 1 {
		panic("enchantment level must never be below 1")
	}
	return Enchantment{t: t, lvl: lvl}
}

// Level returns the current level of the Enchantment as passed to NewEnchantment upon construction.
func (e Enchantment) Level() int {
	return e.lvl
}

// Type returns the EnchantmentType of the Enchantment as passed to NewEnchantment upon construction.
func (e Enchantment) Type() EnchantmentType {
	return e.t
}

// EnchantmentType represents an enchantment type that can be applied to a Stack, with specific behaviour that modifies
// the Stack's behaviour.
// An instance of an EnchantmentType may be created using NewEnchantment.
type EnchantmentType interface {
	// Name returns the name of the enchantment.
	Name() string
	// MaxLevel returns the maximum level the enchantment should be able to have.
	MaxLevel() int
	// Cost returns the minimum and maximum cost the enchantment may inhibit. The higher this range is, the more likely
	// better enchantments are to be selected.
	Cost(level int) (int, int)
	// Rarity returns the enchantment's rarity.
	Rarity() EnchantmentRarity
	// CompatibleWithEnchantment is called when an enchantment is added to an item. It can be used to check if
	// the enchantment is compatible with other enchantments.
	CompatibleWithEnchantment(t EnchantmentType) bool
	// CompatibleWithItem is also called when an enchantment is added to an item. It can be used to check if
	// the enchantment is compatible with the item type.
	CompatibleWithItem(i world.Item) bool
}

// Enchantable is an interface that can be implemented by items that can be enchanted through an enchanting table.
type Enchantable interface {
	// EnchantmentValue returns the value the item may inhibit on possible enchantments.
	EnchantmentValue() int
}

// RegisterEnchantment registers an enchantment with the ID passed. Once registered, enchantments may be received
// by instantiating an EnchantmentType struct (e.g. enchantment.Protection{})
func RegisterEnchantment(id int, enchantment EnchantmentType) {
	enchantmentsMap[id] = enchantment
	enchantmentIDs[enchantment] = id
}

var (
	enchantmentsMap = map[int]EnchantmentType{}
	enchantmentIDs  = map[EnchantmentType]int{}
)

// EnchantmentByID attempts to return an enchantment by the ID it was registered with. If found, the enchantment found
// is returned and the bool true.
func EnchantmentByID(id int) (EnchantmentType, bool) {
	e, ok := enchantmentsMap[id]
	return e, ok
}

// EnchantmentID attempts to return the ID the enchantment was registered with. If found, the id is returned and
// the bool true.
func EnchantmentID(e EnchantmentType) (int, bool) {
	id, ok := enchantmentIDs[e]
	return id, ok
}

// Enchantments returns a slice of all registered enchantments.
func Enchantments() []EnchantmentType {
	e := slices.Collect(maps.Values(enchantmentsMap))
	sort.Slice(e, func(i, j int) bool {
		id1, _ := EnchantmentID(e[i])
		id2, _ := EnchantmentID(e[j])
		return id1 < id2
	})
	return e
}

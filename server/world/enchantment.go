package world

// TODO: Move world/enchantment.go, world/item.go, world.inventory.go to their own package in world

type Enchantment interface {
	Level() int
	Type() EnchantmentType
}

// EnchantmentRarity represents an enchantment rarity for enchantments. These rarities may inhibit certain properties,
// such as anvil costs or enchanting table weights.
type EnchantmentRarity interface {
	// Name returns the name of the enchantment rarity.
	Name() string
	// Cost returns the cost of the enchantment rarity.
	Cost() int
	// Weight returns the weight of the enchantment rarity.
	Weight() int
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
	CompatibleWithItem(i Item) bool
}

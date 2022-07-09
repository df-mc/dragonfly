package item

var (
	// EnchantmentRarityCommon is the common enchantment rarity.
	EnchantmentRarityCommon = EnchantmentRarity{Name: "Common", Cost: 1, Weight: 10}
	// EnchantmentRarityUncommon is the uncommon enchantment rarity.
	EnchantmentRarityUncommon = EnchantmentRarity{Name: "Uncommon", Cost: 2, Weight: 5}
	// EnchantmentRarityRare is the rare enchantment rarity.
	EnchantmentRarityRare = EnchantmentRarity{Name: "Rare", Cost: 4, Weight: 2}
	// EnchantmentRarityVeryRare is the very rare enchantment rarity.
	EnchantmentRarityVeryRare = EnchantmentRarity{Name: "Very Rare", Cost: 8, Weight: 1}
)

// EnchantmentRarity is an enchantment rarity type containing a cost and weight for the enchantment rarity.
type EnchantmentRarity struct {
	// Name is the name of the enchantment rarity.
	Name string
	// Cost is the cost of the enchantment rarity.
	Cost int
	// Weight is the weight of the enchantment rarity.
	Weight int
}

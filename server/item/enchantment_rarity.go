package item

var (
	// EnchantmentRarityCommon is the common enchantment rarity.
	EnchantmentRarityCommon = EnchantmentRarity{Name: "Common", Cost: 1}
	// EnchantmentRarityUncommon is the uncommon enchantment rarity.
	EnchantmentRarityUncommon = EnchantmentRarity{Name: "Uncommon", Cost: 2}
	// EnchantmentRarityRare is the rare enchantment rarity.
	EnchantmentRarityRare = EnchantmentRarity{Name: "Rare", Cost: 4}
	// EnchantmentRarityVeryRare is the very rare enchantment rarity.
	EnchantmentRarityVeryRare = EnchantmentRarity{Name: "Very Rare", Cost: 8}
)

// EnchantmentRarity ...
type EnchantmentRarity struct {
	Name string
	Cost int
}

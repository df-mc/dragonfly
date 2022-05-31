package item

var (
	// EnchantmentRarityCommon is the common enchantment rarity.
	EnchantmentRarityCommon = EnchantmentRarity{Name: "Common", ApplyCost: 1}
	// EnchantmentRarityUncommon is the uncommon enchantment rarity.
	EnchantmentRarityUncommon = EnchantmentRarity{Name: "Uncommon", ApplyCost: 2}
	// EnchantmentRarityRare is the rare enchantment rarity.
	EnchantmentRarityRare = EnchantmentRarity{Name: "Rare", ApplyCost: 4}
	// EnchantmentRarityVeryRare is the very rare enchantment rarity.
	EnchantmentRarityVeryRare = EnchantmentRarity{Name: "Very Rare", ApplyCost: 8}
)

// EnchantmentRarity ...
type EnchantmentRarity struct {
	Name      string
	ApplyCost int
}

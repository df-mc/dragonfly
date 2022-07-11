package item

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

var (
	// EnchantmentRarityCommon represents the common enchantment rarity.
	EnchantmentRarityCommon enchantmentRarityCommon
	// EnchantmentRarityUncommon represents the uncommon enchantment rarity.
	EnchantmentRarityUncommon enchantmentRarityUncommon
	// EnchantmentRarityRare represents the rare enchantment rarity.
	EnchantmentRarityRare enchantmentRarityRare
	// EnchantmentRarityVeryRare represents the very rare enchantment rarity.
	EnchantmentRarityVeryRare enchantmentRarityVeryRare
)

// enchantmentRarityCommon represents the common enchantment rarity.
type enchantmentRarityCommon struct{}

func (enchantmentRarityCommon) Name() string { return "Common" }
func (enchantmentRarityCommon) Cost() int    { return 1 }
func (enchantmentRarityCommon) Weight() int  { return 10 }

// enchantmentRarityUncommon represents the uncommon enchantment rarity.
type enchantmentRarityUncommon struct{}

func (enchantmentRarityUncommon) Name() string { return "Uncommon" }
func (enchantmentRarityUncommon) Cost() int    { return 2 }
func (enchantmentRarityUncommon) Weight() int  { return 5 }

// enchantmentRarityRare represents the rare enchantment rarity.
type enchantmentRarityRare struct{}

func (enchantmentRarityRare) Name() string { return "Rare" }
func (enchantmentRarityRare) Cost() int    { return 4 }
func (enchantmentRarityRare) Weight() int  { return 2 }

// enchantmentRarityVeryRare represents the very rare enchantment rarity.
type enchantmentRarityVeryRare struct{}

func (enchantmentRarityVeryRare) Name() string { return "Very Rare" }
func (enchantmentRarityVeryRare) Cost() int    { return 8 }
func (enchantmentRarityVeryRare) Weight() int  { return 1 }

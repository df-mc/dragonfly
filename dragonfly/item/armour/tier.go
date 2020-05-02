package armour

// Tier represents the tier, or material, that a piece of armour is made of.
type Tier struct {
	// BaseDurability is the base durability of armour with this tier. This is otherwise the durability of
	// the helmet with this tier.
	BaseDurability float64
}

// TierLeather is the tier of leather armour.
var TierLeather = Tier{BaseDurability: 55}

// TierGold is the tier of gold armour.
var TierGold = Tier{BaseDurability: 77}

// TierChain is the tier of chain armour.
var TierChain = Tier{BaseDurability: 166}

// TierIron is the tier of iron armour.
var TierIron = Tier{BaseDurability: 165}

// TierDiamond is the tier of diamond armour.
var TierDiamond = Tier{BaseDurability: 363}

// TierNetherite is the tier of netherite armour.
// TODO: Implement netherite armour once 1.16 lands.
var TierNetherite = Tier{BaseDurability: 408}

package item

var (
	// ArmourTierLeather is the ArmourTier of leather armour.
	ArmourTierLeather = ArmourTier{BaseDurability: 55, Name: "leather"}
	// ArmourTierGold is the ArmourTier of gold armour.
	ArmourTierGold = ArmourTier{BaseDurability: 77, Name: "golden"}
	// ArmourTierChain is the ArmourTier of chain armour.
	ArmourTierChain = ArmourTier{BaseDurability: 166, Name: "chainmail"}
	// ArmourTierIron is the ArmourTier of iron armour.
	ArmourTierIron = ArmourTier{BaseDurability: 165, Name: "iron"}
	// ArmourTierDiamond is the ArmourTier of diamond armour.
	ArmourTierDiamond = ArmourTier{BaseDurability: 363, Name: "diamond"}
	// ArmourTierNetherite is the ArmourTier of netherite armour.
	ArmourTierNetherite = ArmourTier{BaseDurability: 408, KnockBackResistance: 0.1, Name: "netherite"}
)

type (
	// Armour represents an item that may be worn as armour. Generally, these items provide armour points, which
	// reduce damage taken. Some pieces of armour also provide toughness, which negates damage proportional to
	// the total damage dealt.
	Armour interface {
		// DefencePoints returns the defence points that the armour provides when worn.
		DefencePoints() float64
		// KnockBackResistance returns a number from 0-1 that decides the amount of knock back force that is
		// resisted upon being attacked. 1 knock back resistance point client-side translates to 10% knock back
		// reduction.
		KnockBackResistance() float64
	}
	// ArmourTier represents the tier, or material, that a piece of armour is made of.
	ArmourTier struct {
		// BaseDurability is the base durability of armour with this tier. This is otherwise the durability of
		// the helmet with this tier.
		BaseDurability float64
		// KnockBackResistance is a number from 0-1 that decides the amount of knock back force that is resisted
		// upon being attacked. 1 knock back resistance point client-side translates to 10% knock back reduction.
		KnockBackResistance float64
		// Name is the name of the tier.
		Name string
	}
	// HelmetType is an Armour item that can be worn in the helmet slot.
	HelmetType interface {
		Armour
		Helmet() bool
	}
	// ChestplateType is an Armour item that can be worn in the chestplate slot.
	ChestplateType interface {
		Armour
		Chestplate() bool
	}
	// LeggingsType are an Armour item that can be worn in the leggings slot.
	LeggingsType interface {
		Armour
		Leggings() bool
	}
	// BootsType are an Armour item that can be worn in the boots slot.
	BootsType interface {
		Armour
		Boots() bool
	}
)

// ArmourTiers returns a list of all armour tiers.
func ArmourTiers() []ArmourTier {
	return []ArmourTier{ArmourTierLeather, ArmourTierGold, ArmourTierChain, ArmourTierIron, ArmourTierDiamond, ArmourTierNetherite}
}

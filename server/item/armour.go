package item

import "image/color"

type (
	// Armour represents an item that may be worn as armour. Generally, these items provide armour points, which
	// reduce damage taken. Some pieces of armour also provide toughness, which negates damage proportional to
	// the total damage dealt.
	Armour interface {
		// DefencePoints returns the defence points that the armour provides when worn.
		DefencePoints() float64
		// Toughness returns the toughness that the armor provides when worn. The toughness reduces defense reduction
		// caused by increased damage.
		Toughness() float64
		// KnockBackResistance returns a number from 0-1 that decides the amount of knock back force that is
		// resisted upon being attacked. 1 knock back resistance point client-side translates to 10% knock back
		// reduction.
		KnockBackResistance() float64
	}
	// ArmourTier represents the tier, or material, that a piece of armour is made of.
	ArmourTier interface {
		// BaseDurability is the base durability of armour with this tier. This is otherwise the durability of
		// the helmet with this tier.
		BaseDurability() float64
		// Toughness reduces the defense reduction caused by damage increases.
		Toughness() float64
		// KnockBackResistance is a number from 0-1 that decides the amount of knock back force that is resisted
		// upon being attacked. 1 knock back resistance point client-side translates to 10% knock back reduction.
		KnockBackResistance() float64
		// Name is the name of the tier.
		Name() string
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

// ArmourTierLeather is the ArmourTier of leather armour
type ArmourTierLeather struct {
	// Colour is the dyed colour of the armour.
	Colour color.RGBA
}

func (ArmourTierLeather) BaseDurability() float64      { return 55 }
func (ArmourTierLeather) Toughness() float64           { return 0 }
func (ArmourTierLeather) KnockBackResistance() float64 { return 0 }
func (ArmourTierLeather) Name() string                 { return "leather" }

// ArmourTierGold is the ArmourTier of gold armour.
type ArmourTierGold struct{}

func (ArmourTierGold) BaseDurability() float64      { return 77 }
func (ArmourTierGold) Toughness() float64           { return 0 }
func (ArmourTierGold) KnockBackResistance() float64 { return 0 }
func (ArmourTierGold) Name() string                 { return "golden" }

// ArmourTierChain is the ArmourTier of chain armour.
type ArmourTierChain struct{}

func (ArmourTierChain) BaseDurability() float64      { return 166 }
func (ArmourTierChain) Toughness() float64           { return 0 }
func (ArmourTierChain) KnockBackResistance() float64 { return 0 }
func (ArmourTierChain) Name() string                 { return "chainmail" }

// ArmourTierIron is the ArmourTier of iron armour.
type ArmourTierIron struct{}

func (ArmourTierIron) BaseDurability() float64      { return 165 }
func (ArmourTierIron) Toughness() float64           { return 0 }
func (ArmourTierIron) KnockBackResistance() float64 { return 0 }
func (ArmourTierIron) Name() string                 { return "iron" }

// ArmourTierDiamond is the ArmourTier of diamond armour.
type ArmourTierDiamond struct{}

func (ArmourTierDiamond) BaseDurability() float64      { return 363 }
func (ArmourTierDiamond) Toughness() float64           { return 2 }
func (ArmourTierDiamond) KnockBackResistance() float64 { return 0 }
func (ArmourTierDiamond) Name() string                 { return "diamond" }

// ArmourTierNetherite is the ArmourTier of netherite armour.
type ArmourTierNetherite struct{}

func (ArmourTierNetherite) BaseDurability() float64      { return 408 }
func (ArmourTierNetherite) Toughness() float64           { return 3 }
func (ArmourTierNetherite) KnockBackResistance() float64 { return 0.1 }
func (ArmourTierNetherite) Name() string                 { return "netherite" }

// ArmourTiers returns a list of all armour tiers.
func ArmourTiers() []ArmourTier {
	return []ArmourTier{ArmourTierLeather{}, ArmourTierGold{}, ArmourTierChain{}, ArmourTierIron{}, ArmourTierDiamond{}, ArmourTierNetherite{}}
}

// armourTierRepairable returns true if the ArmourTier passed is repairable.
func armourTierRepairable(tier ArmourTier) func(Stack) bool {
	return func(stack Stack) bool {
		var ok bool
		switch tier.(type) {
		case ArmourTierLeather:
			_, ok = stack.Item().(Leather)
		case ArmourTierGold:
			_, ok = stack.Item().(GoldIngot)
		case ArmourTierChain, ArmourTierIron:
			_, ok = stack.Item().(IronIngot)
		case ArmourTierDiamond:
			_, ok = stack.Item().(Diamond)
		case ArmourTierNetherite:
			_, ok = stack.Item().(NetheriteIngot)
		}
		return ok
	}
}

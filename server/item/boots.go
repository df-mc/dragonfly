package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

// Boots are a defensive item that may be equipped in the boots armour slot. They come in several tiers, like
// leather, gold, chain, iron and diamond.
type Boots struct {
	// Tier is the tier of the boots.
	Tier ArmourTier
}

// Use handles the auto-equipping of boots in the armour slot when using it.
func (b Boots) Use(_ *world.World, _ User, ctx *UseContext) bool {
	ctx.SwapHeldWithArmour(3)
	return false
}

// MaxCount always returns 1.
func (b Boots) MaxCount() int {
	return 1
}

// DurabilityInfo ...
func (b Boots) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: int(b.Tier.BaseDurability + b.Tier.BaseDurability/5.5),
		BrokenItem:    simpleItem(Stack{}),
	}
}

// SmeltInfo ...
func (b Boots) SmeltInfo() SmeltInfo {
	switch b.Tier {
	case ArmourTierChain, ArmourTierIron:
		return SmeltInfo{
			Product:    NewStack(IronNugget{}, 1),
			Experience: 0.1,
			Regular:    true,
		}
	case ArmourTierGold:
		return SmeltInfo{
			Product:    NewStack(GoldNugget{}, 1),
			Experience: 0.1,
			Regular:    true,
		}
	}
	return SmeltInfo{}
}

// DefencePoints ...
func (b Boots) DefencePoints() float64 {
	switch b.Tier {
	case ArmourTierLeather, ArmourTierGold, ArmourTierChain:
		return 1
	case ArmourTierIron:
		return 2
	case ArmourTierDiamond, ArmourTierNetherite:
		return 3
	}
	panic("invalid boots tier")
}

// Toughness ...
func (b Boots) Toughness() float64 {
	return b.Tier.Toughness
}

// KnockBackResistance ...
func (b Boots) KnockBackResistance() float64 {
	return b.Tier.KnockBackResistance
}

// Boots ...
func (b Boots) Boots() bool {
	return true
}

// EncodeItem ...
func (b Boots) EncodeItem() (name string, meta int16) {
	return "minecraft:" + b.Tier.Name + "_boots", 0
}

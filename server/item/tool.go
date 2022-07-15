package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

var (
	// TypeNone is the ToolType of items that are not tools.
	TypeNone = ToolType{-1}
	// TypePickaxe is the ToolType for pickaxes.
	TypePickaxe = ToolType{0}
	// TypeAxe is the ToolType for axes.
	TypeAxe = ToolType{1}
	// TypeHoe is the ToolType for hoes.
	TypeHoe = ToolType{2}
	// TypeShovel is the ToolType for shovels.
	TypeShovel = ToolType{3}
	// TypeShears is the ToolType for shears.
	TypeShears = ToolType{4}
	// TypeSword is the ToolType for swords.
	TypeSword = ToolType{5}

	// ToolTierWood is the ToolTier of wood tools. This is the lowest possible tier.
	ToolTierWood = ToolTier{HarvestLevel: 1, Durability: 59, BaseMiningEfficiency: 2, BaseAttackDamage: 1, Name: "wooden"}
	// ToolTierGold is the ToolTier of gold tools.
	ToolTierGold = ToolTier{HarvestLevel: 1, Durability: 32, BaseMiningEfficiency: 12, BaseAttackDamage: 1, Name: "golden"}
	// ToolTierStone is the ToolTier of stone tools.
	ToolTierStone = ToolTier{HarvestLevel: 2, Durability: 131, BaseMiningEfficiency: 4, BaseAttackDamage: 2, Name: "stone"}
	// ToolTierIron is the ToolTier of iron tools.
	ToolTierIron = ToolTier{HarvestLevel: 3, Durability: 250, BaseMiningEfficiency: 6, BaseAttackDamage: 3, Name: "iron"}
	// ToolTierDiamond is the ToolTier of diamond tools.
	ToolTierDiamond = ToolTier{HarvestLevel: 4, Durability: 1561, BaseMiningEfficiency: 8, BaseAttackDamage: 4, Name: "diamond"}
	// ToolTierNetherite is the ToolTier of netherite tools. This is the highest possible tier.
	ToolTierNetherite = ToolTier{HarvestLevel: 4, Durability: 2031, BaseMiningEfficiency: 9, BaseAttackDamage: 5, Name: "netherite"}
)

type (
	// Tool represents an item that may be used as a tool.
	Tool interface {
		// ToolType returns the type of the tool. The blocks that can be mined with this tool depend on this
		// tool type.
		ToolType() ToolType
		// HarvestLevel returns the level that this tool is able to harvest. If a block has a harvest level above
		// this one, this tool won't be able to harvest it.
		HarvestLevel() int
		// BaseMiningEfficiency is the base efficiency of the tool, when it comes to mining blocks. This decides
		// the speed with which blocks can be mined.
		// Some tools have a mining efficiency that depends on the block (swords, shears). The block mined is
		// passed for this behaviour.
		BaseMiningEfficiency(b world.Block) float64
	}
	// ToolTier represents the tier, or material, that a Tool is made of.
	ToolTier struct {
		// HarvestLevel is the level that this tier of tools is able to harvest. If a block has a harvest level
		// above this one, a tool with this tier won't be able to harvest it.
		HarvestLevel int
		// BaseMiningEfficiency is the base efficiency of the tier, when it comes to mining blocks. This is
		// specifically used for tools such as pickaxes.
		BaseMiningEfficiency float64
		// BaseAttackDamage is the base attack damage to tools with this tier. All tools have a constant value
		// that is added on top of this.
		BaseAttackDamage float64
		// BaseDurability returns the maximum durability that a tool with this tier has.
		Durability int
		// Name is the name of the tier.
		Name string
	}
	// ToolType represents the type of tool. This decides the type of blocks that the tool is used for.
	ToolType struct{ t }
	t        int

	// ToolNone is a ToolType typically used in functions for items that do not function as tools.
	ToolNone struct{}
)

// ToolTiers returns a ToolTier slice containing all available tiers.
func ToolTiers() []ToolTier {
	return []ToolTier{ToolTierWood, ToolTierGold, ToolTierStone, ToolTierIron, ToolTierDiamond, ToolTierNetherite}
}

// ToolType ...
func (n ToolNone) ToolType() ToolType { return TypeNone }

// HarvestLevel ...
func (n ToolNone) HarvestLevel() int { return 0 }

// BaseMiningEfficiency ...
func (n ToolNone) BaseMiningEfficiency(world.Block) float64 { return 1 }

// toolTierRepairable returns true if the ToolTier passed is repairable.
func toolTierRepairable(tier ToolTier) func(Stack) bool {
	return func(stack Stack) bool {
		switch tier {
		case ToolTierWood:
			if planks, ok := stack.Item().(interface{ RepairsWoodTools() bool }); ok {
				return planks.RepairsWoodTools()
			}
		case ToolTierStone:
			if cobblestone, ok := stack.Item().(interface{ RepairsStoneTools() bool }); ok {
				return cobblestone.RepairsStoneTools()
			}
		case ToolTierGold:
			_, ok := stack.Item().(GoldIngot)
			return ok
		case ToolTierIron:
			_, ok := stack.Item().(IronIngot)
			return ok
		case ToolTierDiamond:
			_, ok := stack.Item().(Diamond)
			return ok
		case ToolTierNetherite:
			_, ok := stack.Item().(NetheriteIngot)
			return ok
		}
		return false
	}
}

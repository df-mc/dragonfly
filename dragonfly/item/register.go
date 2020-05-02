package item

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/armour"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/tool"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
)

func init() {
	world.RegisterItem("minecraft:wooden_pickaxe", Pickaxe{Tier: tool.TierWood})
	world.RegisterItem("minecraft:golden_pickaxe", Pickaxe{Tier: tool.TierGold})
	world.RegisterItem("minecraft:stone_pickaxe", Pickaxe{Tier: tool.TierStone})
	world.RegisterItem("minecraft:iron_pickaxe", Pickaxe{Tier: tool.TierIron})
	world.RegisterItem("minecraft:diamond_pickaxe", Pickaxe{Tier: tool.TierDiamond})

	world.RegisterItem("minecraft:wooden_axe", Axe{Tier: tool.TierWood})
	world.RegisterItem("minecraft:golden_axe", Axe{Tier: tool.TierGold})
	world.RegisterItem("minecraft:stone_axe", Axe{Tier: tool.TierStone})
	world.RegisterItem("minecraft:iron_axe", Axe{Tier: tool.TierIron})
	world.RegisterItem("minecraft:diamond_axe", Axe{Tier: tool.TierDiamond})

	world.RegisterItem("minecraft:wooden_shovel", Shovel{Tier: tool.TierWood})
	world.RegisterItem("minecraft:golden_shovel", Shovel{Tier: tool.TierGold})
	world.RegisterItem("minecraft:stone_shovel", Shovel{Tier: tool.TierStone})
	world.RegisterItem("minecraft:iron_shovel", Shovel{Tier: tool.TierIron})
	world.RegisterItem("minecraft:diamond_shovel", Shovel{Tier: tool.TierDiamond})

	world.RegisterItem("minecraft:wooden_sword", Sword{Tier: tool.TierWood})
	world.RegisterItem("minecraft:golden_sword", Sword{Tier: tool.TierGold})
	world.RegisterItem("minecraft:stone_sword", Sword{Tier: tool.TierStone})
	world.RegisterItem("minecraft:iron_sword", Sword{Tier: tool.TierIron})
	world.RegisterItem("minecraft:diamond_sword", Sword{Tier: tool.TierDiamond})

	world.RegisterItem("minecraft:leather_helmet", Helmet{Tier: armour.TierLeather})
	world.RegisterItem("minecraft:golden_helmet", Helmet{Tier: armour.TierGold})
	world.RegisterItem("minecraft:chain_helmet", Helmet{Tier: armour.TierChain})
	world.RegisterItem("minecraft:iron_helmet", Helmet{Tier: armour.TierIron})
	world.RegisterItem("minecraft:diamond_helmet", Helmet{Tier: armour.TierDiamond})
}

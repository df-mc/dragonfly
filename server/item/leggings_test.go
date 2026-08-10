package item

import "testing"

// TestArmourDurability verifies the durability of every armour piece against the values Minecraft uses. Durability is
// derived from the base of a tier, which is the durability of its helmet: a chestplate is 16/11 of it, leggings 15/11
// and boots 13/11, so every value is a whole multiple of the base divided by 11.
func TestArmourDurability(t *testing.T) {
	tests := []struct {
		name                                string
		tier                                ArmourTier
		helmet, chestplate, leggings, boots int
	}{
		{name: "leather", tier: ArmourTierLeather{}, helmet: 55, chestplate: 80, leggings: 75, boots: 65},
		{name: "golden", tier: ArmourTierGold{}, helmet: 77, chestplate: 112, leggings: 105, boots: 91},
		{name: "copper", tier: ArmourTierCopper{}, helmet: 121, chestplate: 176, leggings: 165, boots: 143},
		{name: "chainmail", tier: ArmourTierChain{}, helmet: 165, chestplate: 240, leggings: 225, boots: 195},
		{name: "iron", tier: ArmourTierIron{}, helmet: 165, chestplate: 240, leggings: 225, boots: 195},
		{name: "diamond", tier: ArmourTierDiamond{}, helmet: 363, chestplate: 528, leggings: 495, boots: 429},
		{name: "netherite", tier: ArmourTierNetherite{}, helmet: 407, chestplate: 592, leggings: 555, boots: 481},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pieces := []struct {
				name string
				got  int
				want int
			}{
				{name: "helmet", got: Helmet{Tier: tt.tier}.DurabilityInfo().MaxDurability, want: tt.helmet},
				{name: "chestplate", got: Chestplate{Tier: tt.tier}.DurabilityInfo().MaxDurability, want: tt.chestplate},
				{name: "leggings", got: Leggings{Tier: tt.tier}.DurabilityInfo().MaxDurability, want: tt.leggings},
				{name: "boots", got: Boots{Tier: tt.tier}.DurabilityInfo().MaxDurability, want: tt.boots},
			}
			for _, p := range pieces {
				if p.got != p.want {
					t.Errorf("%v durability = %v, want %v", p.name, p.got, p.want)
				}
			}
		})
	}
}

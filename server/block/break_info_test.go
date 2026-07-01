package block_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/stretchr/testify/require"
)

// TestBreakDuration verifies the Bedrock Edition breaking calculation against the reference values
// documented at https://minecraft.wiki/w/Breaking#Calculation.
func TestBreakDuration(t *testing.T) {
	diamondPick := item.NewStack(item.Pickaxe{Tier: item.ToolTierDiamond}, 1)
	efficiencyDiamondPick := diamondPick.WithEnchantments(item.NewEnchantment(enchantment.Efficiency, 3))
	woodPick := item.NewStack(item.Pickaxe{Tier: item.ToolTierWood}, 1)

	tests := []struct {
		name      string
		block     world.Block
		stack     item.Stack
		ctx       block.BreakContext
		wantTicks int64
	}{
		{
			name:      "efficiency adds to best harvestable tool speed",
			block:     block.Stone{},
			stack:     efficiencyDiamondPick,
			wantTicks: 3,
		},
		{
			name:      "haste applies speed and damage multipliers",
			block:     block.Stone{},
			stack:     diamondPick,
			ctx:       block.BreakContext{HasteLevel: 1},
			wantTicks: 4,
		},
		{
			name:      "grounded best tool",
			block:     block.Stone{},
			stack:     diamondPick,
			wantTicks: 6,
		},
		{
			name:      "aqua affinity removes water penalty",
			block:     block.Stone{},
			stack:     diamondPick,
			ctx:       block.BreakContext{Underwater: true, AquaAffinity: true},
			wantTicks: 6,
		},
		{
			name:      "mining fatigue applies speed and damage multipliers",
			block:     block.Stone{},
			stack:     diamondPick,
			ctx:       block.BreakContext{MiningFatigueLevel: 1},
			wantTicks: 27,
		},
		{
			name:      "airborne penalty before rounding",
			block:     block.Stone{},
			stack:     diamondPick,
			ctx:       block.BreakContext{AirBorne: true},
			wantTicks: 29,
		},
		{
			name:      "water without aqua affinity slows mining",
			block:     block.Stone{},
			stack:     diamondPick,
			ctx:       block.BreakContext{Underwater: true},
			wantTicks: 29,
		},
		{
			name:      "wrong tier tool cannot harvest",
			block:     block.DiamondOre{Type: block.StoneOre()},
			stack:     woodPick,
			wantTicks: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := block.BreakDuration(tt.block, tt.stack, tt.ctx).Milliseconds() / 50
			require.Equal(t, tt.wantTicks, got)
		})
	}
}

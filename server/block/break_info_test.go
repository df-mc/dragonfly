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
			ctx:       block.BreakContext{Airborne: true},
			wantTicks: 29,
		},
		{
			name:      "efficiency instant-mines a soft block on land",
			block:     block.Netherrack{},
			stack:     efficiencyDiamondPick,
			wantTicks: 0,
		},
		{
			name:      "airborne prevents instant-mining the same soft block",
			block:     block.Netherrack{},
			stack:     efficiencyDiamondPick,
			ctx:       block.BreakContext{Airborne: true},
			wantTicks: 4,
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

// TestBreaksInstantly verifies that BreaksInstantly reports an instant break both for zero-hardness blocks
// and for positive-hardness blocks that the item breaks within a single tick.
func TestBreaksInstantly(t *testing.T) {
	diamondPick := item.NewStack(item.Pickaxe{Tier: item.ToolTierDiamond}, 1)
	efficiencyDiamondPick := diamondPick.WithEnchantments(item.NewEnchantment(enchantment.Efficiency, 3))

	tests := []struct {
		name  string
		block world.Block
		stack item.Stack
		want  bool
	}{
		{
			name:  "zero-hardness block breaks instantly with anything",
			block: block.ShortGrass{},
			stack: item.Stack{},
			want:  true,
		},
		{
			name:  "soft block breaks instantly with an efficiency tool",
			block: block.Netherrack{},
			stack: efficiencyDiamondPick,
			want:  true,
		},
		{
			name:  "soft block does not break instantly by hand",
			block: block.Netherrack{},
			stack: item.Stack{},
			want:  false,
		},
		{
			name:  "stone does not break instantly with a diamond pickaxe",
			block: block.Stone{},
			stack: diamondPick,
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, block.BreaksInstantly(tt.block, tt.stack))
		})
	}
}

package block_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
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
			if got := block.BreakDuration(tt.block, tt.stack, tt.ctx).Milliseconds() / 50; got != tt.wantTicks {
				t.Errorf("got %d ticks, want %d", got, tt.wantTicks)
			}
		})
	}
}

// TestBreaksInstantly verifies that BreaksInstantly reports an instant break only for zero-hardness blocks,
// not for positive-hardness blocks that merely break within one tick due to a fast tool.
func TestBreaksInstantly(t *testing.T) {
	tests := []struct {
		name  string
		block world.Block
		want  bool
	}{
		{
			name:  "zero-hardness block breaks instantly",
			block: block.ShortGrass{},
			want:  true,
		},
		{
			name:  "positive-hardness block does not break instantly",
			block: block.Netherrack{},
			want:  false,
		},
		{
			name:  "stone does not break instantly",
			block: block.Stone{},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := block.BreaksInstantly(tt.block); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestBreakInfoZeroValueSmelters verifies BreakInfo works on zero-value
// smelter blocks, as produced when decoding blocks from a chunk palette.
func TestBreakInfoZeroValueSmelters(t *testing.T) {
	for _, b := range []world.Block{block.Furnace{}, block.BlastFurnace{}, block.Smoker{}} {
		name, _ := b.EncodeBlock()
		if xp := b.(block.Breakable).BreakInfo().XPDrops; xp[0] != 0 || xp[1] != 0 {
			t.Errorf("%v: expected no XP drops for zero-value block, got %+v", name, xp)
		}
	}
}

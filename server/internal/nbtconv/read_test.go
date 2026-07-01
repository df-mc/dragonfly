package nbtconv

import (
	"testing"

	"github.com/df-mc/dragonfly/server/item"
	_ "github.com/df-mc/dragonfly/server/item/enchantment"
)

func TestReadEnchantments_ClampsLevelBelowOne(t *testing.T) {
	stack := item.NewStack(item.Sword{Tier: item.ToolTierDiamond}, 1)

	readEnchantments(map[string]any{
		"ench": []any{
			map[string]any{
				"id":  int16(9),
				"lvl": int16(0),
			},
		},
	}, &stack)

	enchants := stack.Enchantments()
	if len(enchants) != 1 {
		t.Fatalf("expected 1 enchantment, got %d", len(enchants))
	}
	if enchants[0].Level() != 1 {
		t.Fatalf("expected clamped enchantment level 1, got %d", enchants[0].Level())
	}
}

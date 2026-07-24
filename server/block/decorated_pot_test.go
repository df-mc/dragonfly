package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/item"
)

func TestDecoratedPotDecodeNBTIgnoresExtraSherds(t *testing.T) {
	pot := (DecoratedPot{}).DecodeNBT(map[string]any{
		"sherds": []any{
			"minecraft:brick",
			"minecraft:brick",
			"minecraft:brick",
			"minecraft:brick",
			"minecraft:brick",
		},
	}).(DecoratedPot)

	for i, decoration := range pot.Decorations {
		if _, ok := decoration.(item.Brick); !ok {
			t.Fatalf("decoration %v was %T, expected item.Brick", i, decoration)
		}
	}
}

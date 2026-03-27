package gen

import "testing"

func TestOverworldSurfaceRuntime(t *testing.T) {
	const seed = int64(12345)

	biomeSource, err := NewBiomeSource(seed, NewWorldgenRegistry(), "overworld")
	if err != nil {
		t.Fatalf("create biome source: %v", err)
	}
	runtime := NewOverworldSurfaceRuntime(seed, NewNoiseRegistry(seed), biomeSource)
	ids := map[string]uint32{
		"minecraft:air":         1,
		"minecraft:dirt":        2,
		"minecraft:grass_block": 3,
		"minecraft:sand":        4,
	}
	lookup := func(name string, _ map[string]string) uint32 {
		if rid, ok := ids[name]; ok {
			return rid
		}
		return 0
	}

	t.Run("desert surface uses sand", func(t *testing.T) {
		ctx := SurfaceContext{
			BlockX:          0,
			BlockY:          64,
			BlockZ:          0,
			SurfaceDepth:    3,
			WaterHeight:     surfaceNoWaterHeight,
			StoneDepthAbove: 0,
			StoneDepthBelow: 5,
			Biome:           BiomeDesert,
			MinSurfaceLevel: 60,
			MinY:            -64,
			MaxY:            320,
		}

		rid, ok := runtime.TryApply(ctx, lookup)
		if !ok {
			t.Fatal("expected desert surface rule to match")
		}
		if rid != ids["minecraft:sand"] {
			t.Fatalf("expected sand runtime ID %d, got %d", ids["minecraft:sand"], rid)
		}
	})

	t.Run("plains surface uses grass", func(t *testing.T) {
		ctx := SurfaceContext{
			BlockX:          0,
			BlockY:          64,
			BlockZ:          0,
			SurfaceDepth:    3,
			WaterHeight:     surfaceNoWaterHeight,
			StoneDepthAbove: 0,
			StoneDepthBelow: 5,
			Biome:           BiomePlains,
			MinSurfaceLevel: 60,
			MinY:            -64,
			MaxY:            320,
		}

		rid, ok := runtime.TryApply(ctx, lookup)
		if !ok {
			t.Fatal("expected plains surface rule to match")
		}
		if rid != ids["minecraft:grass_block"] {
			t.Fatalf("expected grass runtime ID %d, got %d", ids["minecraft:grass_block"], rid)
		}
	})

	t.Run("cave floor stays untouched below preliminary surface", func(t *testing.T) {
		ctx := SurfaceContext{
			BlockX:          0,
			BlockY:          30,
			BlockZ:          0,
			SurfaceDepth:    3,
			WaterHeight:     surfaceNoWaterHeight,
			StoneDepthAbove: 0,
			Biome:           BiomePlains,
			MinSurfaceLevel: 60,
			MinY:            -64,
			MaxY:            320,
		}

		if _, ok := runtime.TryApply(ctx, lookup); ok {
			t.Fatal("expected cave floor to remain unchanged")
		}
	})
}

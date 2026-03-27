package gen

import "testing"

func TestLookupBiomeMushroomFields(t *testing.T) {
	got := lookupBiome([6]int64{0, 0, -11000, 0, 0, 0})
	if got != BiomeMushroomFields {
		t.Fatalf("expected mushroom fields, got %v", got)
	}
}

func TestLookupBiomeRiver(t *testing.T) {
	got := lookupBiome([6]int64{0, 0, 0, 1000, 0, 0})
	if got != BiomeRiver {
		t.Fatalf("expected river, got %v", got)
	}
}

func TestLookupBiomePeakVariant(t *testing.T) {
	got := lookupBiome([6]int64{-6000, 0, 4000, -9000, 0, 6000})
	if got != BiomeFrozenPeaks {
		t.Fatalf("expected frozen peaks, got %v", got)
	}
}

func TestLookupPresetBiomeOverworldPointsMushroomFields(t *testing.T) {
	got := lookupPresetBiome([6]int64{0, 0, -11000, 0, 0, 0}, overworldPresetPoints)
	if got != BiomeMushroomFields {
		t.Fatalf("expected mushroom fields, got %v", got)
	}
}

func TestLookupPresetBiomeOverworldPointsRiver(t *testing.T) {
	got := lookupPresetBiome([6]int64{0, 0, 0, 1000, 0, 0}, overworldPresetPoints)
	if got != BiomeRiver {
		t.Fatalf("expected river, got %v", got)
	}
}

func TestLookupPresetBiomeOverworldPointsPeakVariant(t *testing.T) {
	got := lookupPresetBiome([6]int64{-6000, 0, 4000, -9000, 0, 6000}, overworldPresetPoints)
	if got != BiomeFrozenPeaks {
		t.Fatalf("expected frozen peaks, got %v", got)
	}
}

func TestLookupBiomeMatchesOverworldPresetPointsForSampledCoordinates(t *testing.T) {
	noise := NewBiomeNoise(0)
	for x := -512; x <= 512; x += 64 {
		for z := -512; z <= 512; z += 64 {
			for y := -64; y <= 256; y += 64 {
				climate := noise.SampleClimate(x, y, z)
				got := lookupOverworldPresetBiome(climate)
				want := lookupPresetBiome(climate, overworldPresetPoints)
				if got != want {
					t.Fatalf("sample (%d,%d,%d) climate=%v: expected %v from preset points, got %v", x, y, z, climate, want, got)
				}
			}
		}
	}
}

func TestLookupPresetBiomeNetherPoints(t *testing.T) {
	tests := []struct {
		name    string
		climate [6]int64
		want    Biome
	}{
		{
			name:    "nether_wastes",
			climate: [6]int64{0, 0, 0, 0, 0, 0},
			want:    BiomeNetherWastes,
		},
		{
			name:    "soul_sand_valley",
			climate: [6]int64{0, -5000, 0, 0, 0, 0},
			want:    BiomeSoulSandValley,
		},
		{
			name:    "crimson_forest",
			climate: [6]int64{4000, 0, 0, 0, 0, 0},
			want:    BiomeCrimsonForest,
		},
		{
			name:    "warped_forest",
			climate: [6]int64{0, 5000, 0, 0, 0, 3750},
			want:    BiomeWarpedForest,
		},
		{
			name:    "basalt_deltas",
			climate: [6]int64{-5000, 0, 0, 0, 0, 1750},
			want:    BiomeBasaltDeltas,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lookupPresetBiome(tt.climate, netherPresetPoints)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestNewBiomeSourceSupportsNetherPreset(t *testing.T) {
	source, err := NewBiomeSource(0, NewWorldgenRegistry(), "nether")
	if err != nil {
		t.Fatalf("expected nether preset to resolve: %v", err)
	}
	biome := source.GetBiome(0, 64, 0)
	switch biome {
	case BiomeNetherWastes, BiomeSoulSandValley, BiomeCrimsonForest, BiomeWarpedForest, BiomeBasaltDeltas:
	default:
		t.Fatalf("expected a nether biome, got %v", biome)
	}
}

func TestNewBiomeSourceSupportsOverworldPreset(t *testing.T) {
	source, err := NewBiomeSource(0, NewWorldgenRegistry(), "overworld")
	if err != nil {
		t.Fatalf("expected overworld preset to resolve: %v", err)
	}
	biome := source.GetBiome(0, 64, 0)
	switch biome {
	case BiomeOcean, BiomePlains, BiomeForest, BiomeColdOcean, BiomeDeepColdOcean, BiomeBeach, BiomeRiver:
	default:
		t.Fatalf("expected an overworld biome, got %v", biome)
	}
}

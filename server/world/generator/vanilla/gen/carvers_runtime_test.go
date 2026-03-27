package gen

import "testing"

func TestCarverRegistryLoadsOverworldCarvers(t *testing.T) {
	registry := NewCarverRegistry()

	carvers := registry.BiomeCarvers("plains")
	if len(carvers) == 0 {
		t.Fatal("expected plains biome to expose configured carvers")
	}

	caveDef, err := registry.Configured("cave")
	if err != nil {
		t.Fatalf("failed to load cave carver: %v", err)
	}
	cave, err := caveDef.Cave()
	if err != nil {
		t.Fatalf("failed to decode cave carver: %v", err)
	}
	if cave.Probability <= 0 {
		t.Fatalf("expected cave probability to be positive, got %f", cave.Probability)
	}
	if cave.Y.Kind != "uniform" {
		t.Fatalf("expected cave height provider to be uniform, got %q", cave.Y.Kind)
	}

	canyonDef, err := registry.Configured("canyon")
	if err != nil {
		t.Fatalf("failed to load canyon carver: %v", err)
	}
	canyon, err := canyonDef.Canyon()
	if err != nil {
		t.Fatalf("failed to decode canyon carver: %v", err)
	}
	if canyon.Probability <= 0 {
		t.Fatalf("expected canyon probability to be positive, got %f", canyon.Probability)
	}
	if canyon.Shape.WidthSmoothness <= 0 {
		t.Fatalf("expected canyon width smoothness to be positive, got %d", canyon.Shape.WidthSmoothness)
	}
}

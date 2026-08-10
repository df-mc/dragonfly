package world

import (
	"io"
	"log/slog"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// TestNewWorldWeather verifies that a World does not start in a thunderstorm. The weather counters are decremented and
// toggle the weather they control when they reach zero, so leaving them at zero in the default Settings turned both
// rain and thunder on during the first tick, which Config.New runs itself.
func TestNewWorldWeather(t *testing.T) {
	w := Config{Log: slog.New(slog.NewTextHandler(io.Discard, nil))}.New()
	defer w.Close()

	w.set.Lock()
	raining, thundering := w.set.Raining, w.set.Thundering
	w.set.Unlock()

	if raining {
		t.Errorf("new world Raining = %v, want false", raining)
	}
	if thundering {
		t.Errorf("new world Thundering = %v, want false", thundering)
	}
}

// TestBiomeAlwaysResolves verifies that reading a biome returns one that can be used even when none is registered.
// Biomes register themselves from server/world/biome, which is only reached by importing the server package, so a
// World built straight from a Config has an empty registry, and every caller of BiomeByID assumed a non-nil Biome.
func TestBiomeAlwaysResolves(t *testing.T) {
	w := Config{Log: slog.New(slog.NewTextHandler(io.Discard, nil))}.New()
	defer w.Close()

	tests := []struct {
		name string
		pos  cube.Pos
	}{
		{name: "position in the world", pos: cube.Pos{0, 64, 0}},
		{name: "position out of bounds", pos: cube.Pos{0, -5000, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runWorld(w, func(tx *Tx) {
				if b := tx.biome(tt.pos); b == nil {
					t.Errorf("biome(%v) = nil, want a biome", tt.pos)
				}
				// rainingAt reads the rainfall of the biome, which is where a nil biome was dereferenced.
				tx.rainingAt(tt.pos)
			})
		})
	}
}

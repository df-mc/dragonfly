package model

import (
	"math"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// TestBambooBBox checks the size of bamboo bounding boxes and that their
// random offset is quantized within [-0.25, 0.25].
func TestBambooBBox(t *testing.T) {
	const step = 0.5 / 15
	for _, b := range []Bamboo{{}, {Thick: true}} {
		width := 2.0 / 16.0
		if b.Thick {
			width = 3.0 / 16.0
		}
		for x := -60; x <= 60; x += 7 {
			for z := -60; z <= 60; z += 7 {
				pos := cube.Pos{x, 0, z}
				box := b.BBox(pos, nil)[0]

				if w := box.Width(); math.Abs(w-width) > 1e-6 {
					t.Fatalf("bamboo at %v: expected width %v, got %v", pos, width, w)
				}
				if box.Min().Y() != 0 || box.Max().Y() != 1 {
					t.Fatalf("bamboo at %v: expected height 0-1, got %v-%v", pos, box.Min().Y(), box.Max().Y())
				}
				for _, off := range []float64{box.Min().X() - 0.5, box.Min().Z() - 0.5} {
					if off < -0.25-1e-6 || off > 0.25+1e-6 {
						t.Fatalf("bamboo at %v: offset %v outside [-0.25, 0.25]", pos, off)
					}
					if k := (off + 0.25) / step; math.Abs(k-math.Round(k)) > 1e-4 {
						t.Fatalf("bamboo at %v: offset %v not quantized to steps of %v", pos, off, step)
					}
				}
			}
		}
	}
}

package item

import (
	"math"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// TestThrowableOffsetVertical checks that throws aimed straight up or down,
// where the pitch reaches the clamp, keep a finite pitch instead of NaN.
func TestThrowableOffsetVertical(t *testing.T) {
	for _, pitch := range []float64{90, -90, 89.9, -89.9, 89.95, 0, 45} {
		r := throwableOffset(cube.Rotation{-51.43, pitch})
		if math.IsNaN(r[1]) || math.IsNaN(r.Vec3()[0]) {
			t.Errorf("pitch %v: got non-finite rotation %v", pitch, r)
		}
	}
}

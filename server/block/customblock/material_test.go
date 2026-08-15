package customblock

import (
	"math"
	"testing"
)

func TestMaterial_WithAmbientOcclusionIntensity(t *testing.T) {
	for _, intensity := range []float32{0, 2.5, 10} {
		enc := NewMaterial("texture", OpaqueRenderMethod()).WithAmbientOcclusionIntensity(intensity).Encode()
		if got, ok := enc["ambient_occlusion"].(float32); !ok || got != intensity {
			t.Fatalf("ambient_occlusion = %v (%T), want %v (float32)", enc["ambient_occlusion"], enc["ambient_occlusion"], intensity)
		}
	}
}

func TestMaterial_WithAmbientOcclusionIntensityPanicsOutsideRange(t *testing.T) {
	for _, tc := range []struct {
		name      string
		intensity float32
	}{
		{"below minimum", -0.1},
		{"above maximum", 10.1},
		{"not a number", float32(math.NaN())},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("WithAmbientOcclusionIntensity did not panic")
				}
			}()
			NewMaterial("texture", OpaqueRenderMethod()).WithAmbientOcclusionIntensity(tc.intensity)
		})
	}
}

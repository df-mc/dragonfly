package customblock

import "testing"

func TestMaterialEncodeAmbientOcclusion(t *testing.T) {
	tests := []struct {
		name     string
		material Material
		want     float32
	}{
		{name: "enabled", material: NewMaterial("stone", OpaqueRenderMethod()), want: 1},
		{name: "disabled", material: NewMaterial("stone", OpaqueRenderMethod()).WithoutAmbientOcclusion(), want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := test.material.Encode()["ambient_occlusion"].(float32)
			if !ok {
				t.Fatalf("ambient_occlusion has type %T, want float32", test.material.Encode()["ambient_occlusion"])
			}
			if got != test.want {
				t.Fatalf("ambient_occlusion = %v, want %v", got, test.want)
			}
		})
	}
}

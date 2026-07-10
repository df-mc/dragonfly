package model

import "testing"

func TestOffsetSeedUsesSignedCoordinates(t *testing.T) {
	tests := []struct {
		x, z int32
		want uint64
	}{
		{x: -1, want: 0x6a09e667a2e38a49},
		{x: -5, z: 100, want: 0x6a09e6679c2fc6dd},
		{x: -1000, z: -1000, want: 0x6a09e667b197cf06},
		{x: 1371, want: 0x95f619986c715225},
		{x: 2000, want: 0x95f619982eaa1307},
		{x: 5_000_000, z: -5_000_000, want: 0x6a09e66785e88ae0},
		{x: -5_000_000, z: 5_000_000, want: 0x6a09e66785e88ae0},
	}
	for _, test := range tests {
		if got := offsetSeed(test.x, test.z); got != test.want {
			t.Errorf("offsetSeed(%d, %d) = %#x, want %#x", test.x, test.z, got, test.want)
		}
	}
}

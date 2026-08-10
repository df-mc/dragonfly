package enchantment

import "testing"

// TestRespirationChance verifies that a higher level of Respiration blocks the air supply from ticking more often
// rather than less. The chance is level/(level+1), which coincides with 1/(level+1) at level one only.
func TestRespirationChance(t *testing.T) {
	tests := []struct {
		level int
		want  float64
	}{
		{level: 1, want: 1.0 / 2.0},
		{level: 2, want: 2.0 / 3.0},
		{level: 3, want: 3.0 / 4.0},
	}
	var last float64
	for _, tt := range tests {
		got := Respiration.Chance(tt.level)
		if got != tt.want {
			t.Errorf("Chance(%v) = %v, want %v", tt.level, got, tt.want)
		}
		if got <= last {
			t.Errorf("Chance(%v) = %v, want more than the %v of the level below it", tt.level, got, last)
		}
		last = got
	}
}

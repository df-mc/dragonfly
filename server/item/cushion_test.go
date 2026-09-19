package item

import "testing"

// TestCushionYaw checks the direction a cushion is placed in, as recorded on a vanilla 1.26.50 dedicated server.
func TestCushionYaw(t *testing.T) {
	cases := []struct{ user, cushion float64 }{
		{0, -180}, {45, -90}, {26.57, -180}, {-26.57, -180}, {-45, -180}, {-56.31, 90}, {90, -90}, {180, 0}, {-90, 90},
	}
	for _, c := range cases {
		if got := cushionYaw(c.user); got != c.cushion {
			t.Errorf("user yaw %v: got cushion yaw %v, vanilla %v", c.user, got, c.cushion)
		}
	}
}

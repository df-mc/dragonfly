package cube

import "math"

// Orientation represents the orientation of a sign
type Orientation int

// OrientationFromYaw returns an Orientation value that (roughly) matches the yaw passed.
func OrientationFromYaw(yaw float64) Orientation {
	yaw = math.Mod(yaw, 360)
	return Orientation(math.Round(yaw / 360 * 16))
}

// Yaw returns the yaw value that matches the orientation.
func (o Orientation) Yaw() float64 {
	return float64(o) / 16 * 360
}

// Opposite returns the opposite orientation value of the Orientation.
func (o Orientation) Opposite() Orientation {
	return OrientationFromYaw(o.Yaw() + 180)
}

// RotateLeft rotates the orientation left by 90 degrees and returns it.
func (o Orientation) RotateLeft() Orientation {
	return OrientationFromYaw(o.Yaw() - 90)
}

// RotateRight rotates the orientation right by 90 degrees and returns it.
func (o Orientation) RotateRight() Orientation {
	return OrientationFromYaw(o.Yaw() + 90)
}

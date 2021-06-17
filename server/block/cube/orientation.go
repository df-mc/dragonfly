package cube

import "math"

// Orientation represents the orientation of a sign
type Orientation int

const (
	south Orientation = iota
	southSouthWest
	southWest
	westSouthwest
	west
	westNorthwest
	northwest
	northNorthwest
	north
	northNortheast
	northeast
	eastNortheast
	east
	eastSoutheast
	southeast
	southSoutheast
)

func YawToOrientation(yaw float64) Orientation {
	return Orientation(math.Floor((yaw * 16 / 360) + 0.5))
}

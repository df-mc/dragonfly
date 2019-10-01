package block

// Axis represents the axis that a block may be directed in. Most blocks do not have an axis, but blocks such
// as logs or pillars do.
type Axis int

// XAxis returns an Axis value that represents the X axis. Blocks that have this axis face towards the X axis,
// which is east/west in-game.
func XAxis() Axis {
	return Axis(0x01)
}

// YAxis returns an Axis value that represents the Y axis. Blocks that have this axis face towards the Y axis,
// which is up/down in-game.
func YAxis() Axis {
	return Axis(0x00)
}

// ZAxis returns an Axis value that represents the Z axis. Blocks that have this axis face towards the Z axis,
// which is north/south in-game.
func ZAxis() Axis {
	return Axis(0x02)
}

// axisFromInt16 decodes an int16 to an Axis instance.
func axisFromInt16(x int16) Axis {
	return Axis(x & 0x03)
}

// axisToInt16 encodes an Axis to an int16.
func axisToInt16(x Axis) int16 {
	return int16(x & 0x03)
}

package nbtconv

import (
	"encoding/binary"
	"image/color"
)

// Int32FromRGBA converts a color.RGBA into an int32. These int32s are present in, for example, signs.
func Int32FromRGBA(x color.RGBA) int32 {
	if x.R == 0 && x.G == 0 && x.B == 0 {
		// Default to black colour. The default (0x000000) is a transparent colour. Text with this colour will not show
		// up on the sign.
		return int32(-0x1000000)
	}
	return int32(binary.BigEndian.Uint32([]byte{x.A, x.R, x.G, x.B}))
}

// RGBAFromInt32 converts an int32 into a color.RGBA. These int32s are present in, for example, signs.
func RGBAFromInt32(x int32) color.RGBA {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(x))

	return color.RGBA{A: b[0], R: b[1], G: b[2], B: b[3]}
}

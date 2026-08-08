package colour

import (
	"encoding/binary"
	"image/color"
)

// Int32FromRGBA converts RGBA to its packed ARGB representation.
func Int32FromRGBA(c color.RGBA) int32 {
	return int32(binary.BigEndian.Uint32([]byte{c.A, c.R, c.G, c.B}))
}

// Int32FromRGBAOpaqueBlack converts RGBA to packed ARGB and keeps black opaque.
func Int32FromRGBAOpaqueBlack(c color.RGBA) int32 {
	if c.R == 0 && c.G == 0 && c.B == 0 {
		// Transparent black is invisible on signs, so keep black opaque.
		return int32(-0x1000000)
	}
	return Int32FromRGBA(c)
}

// RGBAFromInt32 converts packed ARGB to RGBA.
func RGBAFromInt32(v int32) color.RGBA {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(v))
	return color.RGBA{A: b[0], R: b[1], G: b[2], B: b[3]}
}

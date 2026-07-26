package nbtconv

import (
	"image/color"

	"github.com/df-mc/dragonfly/server/internal/colour"
)

// Int32FromRGBA converts a color.RGBA into an int32. These int32s are present in, for example, signs.
func Int32FromRGBA(x color.RGBA) int32 {
	return colour.Int32FromRGBAOpaqueBlack(x)
}

// RGBAFromInt32 converts an int32 into a color.RGBA. These int32s are present in, for example, signs.
func RGBAFromInt32(x int32) color.RGBA {
	return colour.RGBAFromInt32(x)
}

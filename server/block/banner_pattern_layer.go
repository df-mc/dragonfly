package block

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
)

// BannerPatternLayer is a wrapper over BannerPatternType with a colour property.
type BannerPatternLayer struct {
	// Type represents the type of banner pattern.
	Type BannerPatternType
	// Colour is the colour the pattern should be rendered in.
	Colour item.Colour
}

// EncodeNBT encodes the given BannerPatternLayer into an NBT map.
func (b BannerPatternLayer) EncodeNBT() map[string]any {
	return map[string]any{
		"Pattern": bannerPatternID(b.Type),
		"Color":   int32(invertColour(b.Colour)),
	}
}

// DecodeNBT decodes the given NBT map into a BannerPatternLayer and returns it.
func (b BannerPatternLayer) DecodeNBT(data map[string]any) any {
	id := nbtconv.String(data, "Pattern")
	pattern, exists := BannerPatternByID(id)
	if !exists {
		panic(fmt.Errorf("unknown banner pattern id %q", id))
	}
	b.Type = pattern
	b.Colour = invertColourID(int16(nbtconv.Int32(data, "Color")))
	return b
}

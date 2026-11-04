package item

import (
	"math"
	"slices"
)

const shieldDamageThreshold = 3

// ShieldBanner holds the banner design applied to a shield.
type ShieldBanner struct {
	// BaseColour is the base colour of the design.
	BaseColour Colour
	// Patterns holds the ordered pattern layers of the design.
	Patterns []ShieldPattern
	// Illager is true for the ominous banner design.
	Illager bool
}

// ShieldPattern is a single banner pattern layer applied to a shield. Pattern is the vanilla banner pattern ID.
type ShieldPattern struct {
	Pattern string
	Colour  Colour
}

// Shield is a defensive item that can block incoming attacks while held.
type Shield struct {
	banner    ShieldBanner
	decorated bool
}

// WithBanner returns the shield with a copy of the banner design applied.
func (s Shield) WithBanner(banner ShieldBanner) Shield {
	s.banner, s.decorated = cloneShieldBanner(banner), true
	return s
}

// Banner returns a copy of the shield's banner design, if present.
func (s Shield) Banner() (ShieldBanner, bool) {
	return cloneShieldBanner(s.banner), s.decorated
}

// Decorated reports whether a banner design is applied without copying it.
func (s Shield) Decorated() bool {
	return s.decorated
}

// BlockDurabilityDamage returns the durability lost when the shield blocks damage.
func (Shield) BlockDurabilityDamage(damage float64) int {
	if damage < shieldDamageThreshold {
		return 0
	}
	return int(math.Floor(damage)) + 1
}

// DurabilityInfo ...
func (Shield) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: 337,
		BrokenItem:    simpleItem(Stack{}),
	}
}

// RepairableBy ...
func (Shield) RepairableBy(i Stack) bool {
	return toolTierRepairable(ToolTierWood)(i)
}

// MaxCount always returns 1.
func (Shield) MaxCount() int {
	return 1
}

// OffHand ...
func (Shield) OffHand() bool {
	return true
}

// HandEquipped returns true so that the shield uses its equipped hand transform.
func (Shield) HandEquipped() bool {
	return true
}

// DecodeNBT decodes an optional banner design from shield item data.
func (s Shield) DecodeNBT(data map[string]any) any {
	_, base := data["Base"]
	_, kind := data["Type"]
	patterns := nbtSlice(data, "Patterns")
	if !base && !kind && patterns == nil {
		s.banner, s.decorated = ShieldBanner{}, false
		return s
	}
	banner := ShieldBanner{
		BaseColour: shieldColourFromNBT(nbtInt32(data, "Base")),
		Illager:    nbtInt32(data, "Type") == 1,
		Patterns:   make([]ShieldPattern, 0, len(patterns)),
	}
	for _, value := range patterns {
		pattern, ok := value.(map[string]any)
		if !ok {
			continue
		}
		banner.Patterns = append(banner.Patterns, ShieldPattern{
			Pattern: nbtString(pattern, "Pattern"),
			Colour:  shieldColourFromNBT(nbtInt32(pattern, "Color")),
		})
	}
	// banner and its patterns are method-local, so they can be adopted without WithBanner's clone.
	s.banner, s.decorated = banner, true
	return s
}

// EncodeNBT encodes the shield's optional banner design.
func (s Shield) EncodeNBT() map[string]any {
	if !s.decorated {
		return nil
	}
	patterns := make([]any, 0, len(s.banner.Patterns))
	for _, pattern := range s.banner.Patterns {
		patterns = append(patterns, map[string]any{
			"Pattern": pattern.Pattern,
			"Color":   shieldColourNBT(pattern.Colour),
		})
	}
	return map[string]any{
		"Base":     shieldColourNBT(s.banner.BaseColour),
		"Patterns": patterns,
		"Type":     int32(boolByte(s.banner.Illager)),
	}
}

// EncodeItem ...
func (Shield) EncodeItem() (name string, meta int16) {
	return "minecraft:shield", 0
}

// shieldColourFromNBT returns the colour for the inverted dye ID used by banner NBT.
func shieldColourFromNBT(id int32) Colour {
	if id < 0 || id > 15 {
		return ColourWhite()
	}
	return invertColourID(int16(id))
}

// shieldColourNBT returns the inverted dye ID used by banner NBT.
func shieldColourNBT(colour Colour) int32 {
	return int32(invertColour(colour))
}

// cloneShieldBanner returns a deep copy of banner.
func cloneShieldBanner(banner ShieldBanner) ShieldBanner {
	banner.Patterns = slices.Clone(banner.Patterns)
	return banner
}

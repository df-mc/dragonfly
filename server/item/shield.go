package item

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
	// Banner is the optional banner design applied to the shield.
	Banner *ShieldBanner
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
		s.Banner = nil
		return s
	}
	banner := &ShieldBanner{
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
	s.Banner = banner
	return s
}

// EncodeNBT encodes the shield's optional banner design.
func (s Shield) EncodeNBT() map[string]any {
	if s.Banner == nil {
		return nil
	}
	patterns := make([]any, 0, len(s.Banner.Patterns))
	for _, pattern := range s.Banner.Patterns {
		patterns = append(patterns, map[string]any{
			"Pattern": pattern.Pattern,
			"Color":   shieldColourNBT(pattern.Colour),
		})
	}
	return map[string]any{
		"Base":     shieldColourNBT(s.Banner.BaseColour),
		"Patterns": patterns,
		"Type":     int32(boolByte(s.Banner.Illager)),
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
	return Colours()[uint8(^id)&0xf]
}

// shieldColourNBT returns the inverted dye ID used by banner NBT.
func shieldColourNBT(colour Colour) int32 {
	return int32(^colour.Uint8() & 0xf)
}

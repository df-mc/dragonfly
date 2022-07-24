package item

// FireworkExplosion represents an explosion of a firework.
type FireworkExplosion struct {
	// Shape represents the shape of the explosion.
	Shape FireworkShape
	// Colour is the colour of the explosion.
	Colour Colour
	// Fade is the colour the explosion should fade into.
	Fade Colour
	// Fades is true if the explosion should fade into the fade colour.
	Fades bool
	// Twinkle is true if the explosion should twinkle on explode.
	Twinkle bool
	// Trail is true if the explosion should have a trail.
	Trail bool
}

// EncodeNBT ...
func (f FireworkExplosion) EncodeNBT() map[string]any {
	data := map[string]any{
		"FireworkType":    f.Shape.Uint8(),
		"FireworkColor":   [1]uint8{uint8(invertColour(f.Colour))},
		"FireworkFade":    [0]uint8{},
		"FireworkFlicker": boolByte(f.Twinkle),
		"FireworkTrail":   boolByte(f.Trail),
	}
	if f.Fades {
		data["FireworkFade"] = [1]uint8{uint8(invertColour(f.Fade))}
	}
	return data
}

// DecodeNBT ...
func (f FireworkExplosion) DecodeNBT(data map[string]any) any {
	f.Shape = FireworkTypes()[data["FireworkType"].(uint8)]
	f.Twinkle = data["FireworkFlicker"].(uint8) == 1
	f.Trail = data["FireworkTrail"].(uint8) == 1

	colours := data["FireworkColor"]
	if diskColour, ok := colours.([1]uint8); ok {
		f.Colour = invertColourID(int16(diskColour[0]))
	} else if networkColours, ok := colours.([]any); ok {
		f.Colour = invertColourID(int16(networkColours[0].(uint8)))
	}

	if fades, ok := data["FireworkFade"].([1]uint8); ok {
		f.Fade, f.Fades = invertColourID(int16(fades[0])), true
	}
	return f
}

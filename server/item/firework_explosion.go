package item

// FireworkExplosion represents an explosion of a firework.
type FireworkExplosion struct {
	// Shape represents the shape of the explosion.
	Shape FireworkShape
	// Colour is the colour of the explosion.
	Colour Colour
	// Fade is the colour the explosion should fade into. Fades must be set to true in order for this to function.
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
	if data == nil {
		return f
	}

	if shapeID, ok := fireworkNBTUint8(data["FireworkType"]); ok {
		shapes := FireworkShapes()
		if int(shapeID) < len(shapes) {
			f.Shape = shapes[shapeID]
		}
	}
	if twinkle, ok := fireworkNBTUint8(data["FireworkFlicker"]); ok {
		f.Twinkle = twinkle == 1
	}
	if trail, ok := fireworkNBTUint8(data["FireworkTrail"]); ok {
		f.Trail = trail == 1
	}

	switch colours := data["FireworkColor"].(type) {
	case [1]uint8:
		f.Colour = invertColourID(int16(colours[0]))
	case []uint8:
		if len(colours) > 0 {
			f.Colour = invertColourID(int16(colours[0]))
		}
	case []any:
		if len(colours) > 0 {
			if c, ok := fireworkNBTUint8(colours[0]); ok {
				f.Colour = invertColourID(int16(c))
			}
		}
	}

	switch fades := data["FireworkFade"].(type) {
	case [1]uint8:
		f.Fade, f.Fades = invertColourID(int16(fades[0])), true
	case []uint8:
		if len(fades) > 0 {
			f.Fade, f.Fades = invertColourID(int16(fades[0])), true
		}
	case []any:
		if len(fades) > 0 {
			if fade, ok := fireworkNBTUint8(fades[0]); ok {
				f.Fade, f.Fades = invertColourID(int16(fade)), true
			}
		}
	}
	return f
}

func fireworkNBTUint8(v any) (uint8, bool) {
	switch value := v.(type) {
	case uint8:
		return value, true
	case int8:
		return uint8(value), true
	case int16:
		return uint8(value), true
	case uint16:
		return uint8(value), true
	case int32:
		return uint8(value), true
	case uint32:
		return uint8(value), true
	case int:
		return uint8(value), true
	case uint:
		return uint8(value), true
	default:
		return 0, false
	}
}

package customblock

// Material represents a single material used for rendering part of a custom block.
type Material struct {
	// texture is the name of the texture for the material.
	texture string
	// renderMethod is the method to use when rendering the material.
	renderMethod Method
	// faceDimming is if the material should be dimmed by the direction it's facing.
	faceDimming bool
	// ambientOcclusion is if the material should have ambient occlusion applied when lighting.
	ambientOcclusion bool
}

// NewMaterial returns a new Material with the provided information. It enables face dimming by default and ambient
// occlusion based on the render method given.
func NewMaterial(texture string, method Method) Material {
	return Material{
		faceDimming:      true,
		texture:          texture,
		renderMethod:     method,
		ambientOcclusion: method.AmbientOcclusion(),
	}
}

// WithFaceDimming returns a copy of the Material with face dimming enabled.
func (m Material) WithFaceDimming() Material {
	m.faceDimming = true
	return m
}

// WithoutFaceDimming returns a copy of the Material with face dimming disabled.
func (m Material) WithoutFaceDimming() Material {
	m.faceDimming = false
	return m
}

// WithAmbientOcclusion returns a copy of the Material with ambient occlusion enabled.
func (m Material) WithAmbientOcclusion() Material {
	m.ambientOcclusion = true
	return m
}

// WithoutAmbientOcclusion returns a copy of the Material with ambient occlusion disabled.
func (m Material) WithoutAmbientOcclusion() Material {
	m.ambientOcclusion = false
	return m
}

// Encode returns the material encoded as a map[string]any.
func (m Material) Encode() map[string]any {
	return map[string]any{
		"texture":           m.texture,
		"render_method":     m.renderMethod.String(),
		"face_dimming":      boolByte(m.faceDimming),
		"ambient_occlusion": boolByte(m.ambientOcclusion),
	}
}

// boolByte returns 1 if the bool passed is true, or 0 if it is false.
func boolByte(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

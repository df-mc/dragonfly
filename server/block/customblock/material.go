package customblock

// Material represents a single material used for rendering part of a custom block.
type Material struct {
	// texture is the name of the texture for the material.
	texture string
	// renderMethod is the method to use when rendering the material.
	renderMethod Method
	// faceDimming is if the material should be dimmed by the direction it's facing.
	faceDimming bool
	// ambientOcclusion is how strongly ambient occlusion is applied to the material when
	// lighting it, with 0 disabling it and 1 being the vanilla strength.
	ambientOcclusion float32
}

// NewMaterial returns a new Material with the provided information. It enables face dimming by default and ambient
// occlusion based on the render method given.
func NewMaterial(texture string, method Method) Material {
	m := Material{
		texture:      texture,
		renderMethod: method,
		faceDimming:  true,
	}
	if method.AmbientOcclusion() {
		m.ambientOcclusion = 1
	}
	return m
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

// WithAmbientOcclusion returns a copy of the Material with ambient occlusion enabled at the
// vanilla strength. Use WithAmbientOcclusionIntensity to pick a different one.
func (m Material) WithAmbientOcclusion() Material {
	m.ambientOcclusion = 1
	return m
}

// WithAmbientOcclusionIntensity returns a copy of the Material with ambient occlusion applied
// at the intensity given, where 0 disables it and 1 is the vanilla strength. The client
// expects an intensity between 0 and 10.
func (m Material) WithAmbientOcclusionIntensity(intensity float32) Material {
	if intensity < 0 || intensity > 10 {
		panic("customblock: ambient occlusion intensity must be between 0 and 10")
	}
	m.ambientOcclusion = intensity
	return m
}

// WithoutAmbientOcclusion returns a copy of the Material with ambient occlusion disabled.
func (m Material) WithoutAmbientOcclusion() Material {
	m.ambientOcclusion = 0
	return m
}

// Encode returns the material encoded as a map that can be sent over the network to the client.
func (m Material) Encode() map[string]any {
	return map[string]any{
		"texture":           m.texture,
		"render_method":     m.renderMethod.String(),
		"face_dimming":      m.faceDimming,
		"ambient_occlusion": m.ambientOcclusion,
	}
}

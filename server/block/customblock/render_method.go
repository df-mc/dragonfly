package customblock

// Method is the method to use when rendering a material for a custom block.
type Method struct {
	renderMethod
}

// OpaqueRenderMethod returns the opaque rendering method for a material. It does not render an alpha layer, meaning it
// does not support transparent or translucent textures, only textures that are fully opaque.
func OpaqueRenderMethod() Method {
	return Method{0}
}

// AlphaTestRenderMethod returns the alpha_test rendering method for a material. It does not allow for translucent
// textures, only textures that are fully opaque or fully transparent, used for blocks such as regular glass. It also
// disables ambient occlusion by default.
func AlphaTestRenderMethod() Method {
	return Method{1}
}

// BlendRenderMethod returns the blend rendering method for a material. It allows for transparent and translucent
// textures, used for blocks such as stained-glass. It also disables ambient occlusion by default.
func BlendRenderMethod() Method {
	return Method{2}
}

// DoubleSidedRenderMethod returns the double_sided rendering method for a material. It is used to completely disable
// backface culling, which would be used for flat faces visible from both sides.
func DoubleSidedRenderMethod() Method {
	return Method{3}
}

type renderMethod uint8

// Uint8 returns the render method as a uint8.
func (m renderMethod) Uint8() uint8 {
	return uint8(m)
}

func (m renderMethod) String() string {
	switch m {
	case 0:
		return "opaque"
	case 1:
		return "alpha_test"
	case 2:
		return "blend"
	case 3:
		return "double_sided"
	}
	panic("should never happen")
}

// AmbientOcclusion returns if ambient occlusion should be enabled by default for a material using this rendering method.
func (m renderMethod) AmbientOcclusion() bool {
	return m != 1 && m != 2
}

package support

// Type is a support type.
type Type interface {
	// Supports returns whether the support type can provide full, center, or edge support.
	Supports() (full, center, edge bool)
}

// Edge is a support type that can support blocks, such as rails, on the edges of a block face.
type Edge struct{}

// Supports ...
func (e Edge) Supports() (full, center, edge bool) {
	return false, false, true
}

// Center is a support type that can support blocks from the center of the block face.
type Center struct{}

// Supports ...
func (c Center) Supports() (full, center, edge bool) {
	return false, true, false
}

// Full is a support type that can support any blocks.
type Full struct{}

// Supports ...
func (f Full) Supports() (full, center, edge bool) {
	return true, true, true
}

// None is a support type that can not support any blocks.
type None struct{}

// Supports ...
func (n None) Supports() (full, center, edge bool) {
	return false, false, false
}

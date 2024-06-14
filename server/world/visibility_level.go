package world

type VisibilityLevel struct {
	visibility
}

// PublicVisibility is the default visibility level where the entity is (in)visible
// depending on how it already is publicly viewed.
func PublicVisibility() VisibilityLevel {
	return VisibilityLevel{0}
}

// EnforceInvisible is the visibility level where the entity is always invisible to the viewer.
func EnforceInvisible() VisibilityLevel {
	return VisibilityLevel{1}
}

// EnforceVisible is the visibility level where the entity is always visible to the viewer.
func EnforceVisible() VisibilityLevel {
	return VisibilityLevel{2}
}

type visibility uint8

// EnforceVisibility returns whether metadata should be changed.
func (v visibility) EnforceVisibility() bool {
	return v > 0
}

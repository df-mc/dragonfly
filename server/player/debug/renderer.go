package debug

// Renderer represents an interface for a renderer that can manage debug shapes.
type Renderer interface {
	// AddDebugShape adds a debug shape to the renderer, which should be rendered to the player. If the shape
	// already exists, it will be updated with the new information.
	AddDebugShape(shape Shape)
	// RemoveDebugShape removes a debug shape from the renderer by its unique identifier.
	RemoveDebugShape(shape Shape)
	// VisibleDebugShapes returns a slice of all debug shapes that are currently being shown by the renderer.
	VisibleDebugShapes() []Shape
	// RemoveAllDebugShapes clears all debug shapes from the renderer, removing them from the view of the player.
	RemoveAllDebugShapes()
}

package hud

// Renderer represents an interface that can manage HUD elements for a player.
type Renderer interface {
	// ShowHudElement shows a HUD element to the renderer if it is not already shown.
	ShowHudElement(e Element)
	// HideHudElement hides a HUD element from the renderer if it is not already hidden.
	HideHudElement(e Element)
	// HudElementHidden checks if a HUD element is currently hidden from the renderer.
	HudElementHidden(e Element) bool
}

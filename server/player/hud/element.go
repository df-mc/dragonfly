package hud

// Element represents a HUD element in the game that can either be hidden or shown.
type Element struct {
	element
}

type element uint8

// PaperDoll is the element that shows the player's paper doll, which is a visual representation of the
// player's character model and equipment, as well as any currently played animations. It is located in the
// top left corner of the screen.
func PaperDoll() Element {
	return Element{0}
}

// Armour is the element that shows the player's armour level, sitting either above the hotbar or at the top
// of the screen on in non-classic views.
func Armour() Element {
	return Element{1}
}

// ToolTips is the element that shows useful hints and tips to the player, such as how to use items or
// how to perform certain actions in the game. These tips are displayed at the top right of the screen.
func ToolTips() Element {
	return Element{2}
}

// TouchControls is the element that shows the touch controls on the screen, which is used for touch-based
// devices.
func TouchControls() Element {
	return Element{3}
}

// Crosshair is the element that shows the crosshair in the middle of the screen, which is used for aiming
// and targeting entities or blocks.
func Crosshair() Element {
	return Element{4}
}

// HotBar is the element that shows all the items in the player's hotbar, located at the bottom of the screen.
func HotBar() Element {
	return Element{5}
}

// Health is the element that shows the player's health bar, sitting either above the hotbar or at the top
// of the screen on in non-classic views.
func Health() Element {
	return Element{6}
}

// ProgressBar is the element that shows the player's experience bar. It is always located just above the
// hotbar.
func ProgressBar() Element {
	return Element{7}
}

// Hunger is the element that shows the player's hunger bar, which indicates how hungry the player is and
// how much food they need to consume to restore their hunger. It is located either above the hotbar or at the
// top of the screen on in non-classic views.
func Hunger() Element {
	return Element{8}
}

// AirBubbles is the element that shows the player's air bubbles, which indicate how much air the player has
// left when underwater. It is located either above the hotbar or at the top of the screen on in non-classic
// views. It is only visible when the player is underwater or they are regenerating air after being underwater.
func AirBubbles() Element {
	return Element{9}
}

// HorseHealth is the element that shows the health of the player's horse, which replaces the player's own
// health bar when riding a horse/other entity with health.
func HorseHealth() Element {
	return Element{10}
}

// StatusEffects is the element that shows the icons of the currently active status effects, located on the
// right side of the screen.
func StatusEffects() Element {
	return Element{11}
}

// ItemText is the element that shows the text of the item currently held in the player's hand, which is
// displayed just above the hotbar when switching to a new item.
func ItemText() Element {
	return Element{12}
}

// Uint8 returns the element type as a uint8.
func (s element) Uint8() uint8 {
	return uint8(s)
}

// All returns all the HUD elements that are available to be shown or hidden in the game.
func All() []Element {
	return []Element{
		PaperDoll(), Armour(), ToolTips(), TouchControls(), Crosshair(), HotBar(), Health(),
		ProgressBar(), Hunger(), AirBubbles(), HorseHealth(), StatusEffects(), ItemText(),
	}
}

package dialogue

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"strings"
)

// Menu represents a npc dialogue. It contains a title, body, entity, and a number of buttons no more than
// 6.
type Menu struct {
	title, body string
	// action represents the action this menu is executing. This value is either packet.NPCDialogueActionOpen or
	// packet.NPCDialogueActionClose.
	npc     world.NPC
	buttons []Button
}

// NewMenu creates a new Menu with the Dialogue passed. Title is formatted with accordance to the rules of fmt.Sprintln.
func NewMenu(npc world.NPC, title ...any) Menu {
	return Menu{
		title: format(title),
		npc:   npc,
	}
}

// WithButtons creates a copy of the dialogue Menu and appends the buttons passed to the existing buttons, after
// which the new dialogue Menu is returned. If the count of the buttons passed and the buttons already within the Menu
// pass the threshold of 6, it will return an empty Menu and an error.
func (m Menu) WithButtons(buttons ...Button) (Menu, error) {
	if len(m.buttons)+len(buttons) > 6 {
		return Menu{}, fmt.Errorf("menu has %v buttons, an addition of %v will pass the 6 buttons threashold", len(m.buttons), len(buttons))
	}
	m.buttons = append(m.buttons, buttons...)
	return m, nil
}

// WithBody creates a copy of the dialogue Menu and replaces the existing body with the body passed, after which the
// new dialogue Menu is returned. The text is formatted following the rules of fmt.Sprintln.
func (m Menu) WithBody(body ...any) Menu {
	m.body = format(body)
	return m
}

// NPC returns the entity associated with this Menu.
func (m Menu) NPC() world.Entity {
	return m.npc
}

// Body returns the formatted body passed to Menu by WithBody()
func (m Menu) Body() string {
	return m.body
}

// Buttons will return all the buttons passed to Menu by WithButtons().
func (m Menu) Buttons() []Button {
	return m.buttons
}

// Title returns the formatted body passed to Menu from NewMenu()
func (m Menu) Title() string {
	return m.title
}

// MarshalJSON ...
func (m Menu) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Buttons())
}

// format is a utility function to format a list of values to have spaces between them, but no newline at the
// end.
func format(a []any) string {
	return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintln(a...), "\n"), "\n")
}

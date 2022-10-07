package form

import (
	"encoding/json"
	"fmt"
)

// Menu represents a menu form. These menus are made up of a title and a body, with a number of buttons which
// come below the body. These buttons may also have images on the side of them.
type buttonData struct {
	btn     Button
	onClick Closer
}

type Menu struct {
	title, body string
	btnData     []buttonData
	onClose     Closer
}

// NewMenu creates a new Menu form using the MenuSubmittable passed to handle the output of the form. The
// title passed is formatted following the rules of fmt.Sprintln.
func NewMenu(title ...any) Menu {
	m := Menu{title: format(title)}
	m.verify()
	return m
}

// MarshalJSON ...
func (m Menu) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    "form",
		"title":   m.title,
		"content": m.body,
		"buttons": m.Buttons(),
	})
}

// WithBody creates a copy of the Menu form and changes its body to the body passed, after which the new Menu
// form is returned. The text is formatted following the rules of fmt.Sprintln.
func (m Menu) WithBody(body ...any) Menu {
	m.body = format(body)
	return m
}

// WithButtons creates a copy of the Menu form and appends the buttons passed to the existing buttons, after
// which the new Menu form is returned.
func (m Menu) Button(btn Button, onClick func(Submitter)) Menu {
	m.btnData = append(m.btnData, buttonData{btn, onClick})
	return m
}

func (m Menu) OnClose(onClose Closer) Menu {
	m.onClose = onClose
	return m
}

// Title returns the formatted title passed to the menu upon construction using NewMenu().
func (m Menu) Title() string {
	return m.title
}

// Body returns the formatted text in the body passed to the menu using WithBody().
func (m Menu) Body() string {
	return m.body
}

// Buttons returns a list of all buttons of the MenuSubmittable. It parses them from the fields using
// reflection and returns them.
func (m Menu) Buttons() []Button {
	buttons := make([]Button, len(m.btnData))
	for i, data := range m.btnData {
		buttons[i] = data.btn
	}
	return buttons
}

// SubmitJSON submits a JSON value to the menu, containing the index of the button clicked.
func (m Menu) SubmitJSON(b []byte, submitter Submitter) error {
	if b == nil {
		if m.onClose != nil {
			m.onClose(submitter)
		}
		return nil
	}

	var index uint
	err := json.Unmarshal(b, &index)
	if err != nil {
		return fmt.Errorf("cannot parse button index as int: %w", err)
	}
	btnData := m.btnData
	if index >= uint(len(btnData)) {
		return fmt.Errorf("button index points to inexistent button: %v (only %v buttons present)", index, len(btnData))
	}
	call := btnData[index].onClick
	if call != nil {
		call(submitter)
	}
	return nil
}

// verify verifies if the form is valid, checking all fields are of the type Button. It panics if the form is
// not valid.
func (m Menu) verify() {

}

func (m Menu) __() {}

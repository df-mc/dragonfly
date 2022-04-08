package form

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Menu represents a menu form. These menus are made up of a title and a body, with a number of buttons which
// come below the body. These buttons may also have images on the side of them.
type Menu struct {
	title, body string
	submittable MenuSubmittable
	buttons     []Button
}

// NewMenu creates a new Menu form using the MenuSubmittable passed to handle the output of the form. The
// title passed is formatted following the rules of fmt.Sprintln.
func NewMenu(submittable MenuSubmittable, title ...any) Menu {
	t := reflect.TypeOf(submittable)
	if t.Kind() != reflect.Struct {
		panic("submittable must be struct")
	}
	m := Menu{title: format(title), submittable: submittable}
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
func (m Menu) WithButtons(buttons ...Button) Menu {
	m.buttons = append(m.buttons, buttons...)
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
	v := reflect.New(reflect.TypeOf(m.submittable)).Elem()
	v.Set(reflect.ValueOf(m.submittable))

	buttons := make([]Button, 0)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}
		// Each exported field is guaranteed to be of type Button.
		buttons = append(buttons, field.Interface().(Button))
	}
	buttons = append(buttons, m.buttons...)
	return buttons
}

// SubmitJSON submits a JSON value to the menu, containing the index of the button clicked.
func (m Menu) SubmitJSON(b []byte, submitter Submitter) error {
	if b == nil {
		if closer, ok := m.submittable.(Closer); ok {
			closer.Close(submitter)
		}
		return nil
	}

	var index uint
	err := json.Unmarshal(b, &index)
	if err != nil {
		return fmt.Errorf("cannot parse button index as int: %w", err)
	}
	buttons := m.Buttons()
	if index >= uint(len(buttons)) {
		return fmt.Errorf("button index points to inexistent button: %v (only %v buttons present)", index, len(buttons))
	}
	m.submittable.Submit(submitter, buttons[index])
	return nil
}

// verify verifies if the form is valid, checking all fields are of the type Button. It panics if the form is
// not valid.
func (m Menu) verify() {
	v := reflect.New(reflect.TypeOf(m.submittable)).Elem()
	v.Set(reflect.ValueOf(m.submittable))
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanSet() {
			continue
		}
		if _, ok := v.Field(i).Interface().(Button); !ok {
			panic("all exported fields must be of the type form.Button")
		}
	}
}

func (m Menu) __() {}

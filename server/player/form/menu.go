package form

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/df-mc/dragonfly/server/world"
)

// Menu represents a menu form. These menus are made up of a title and a body, with a number of elements which
// come below the body. These elements can include buttons, dividers, headers, and labels.
type Menu struct {
	title, body string
	submittable MenuSubmittable
	elements    []MenuElement
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
		"type":     "form",
		"title":    m.title,
		"content":  m.body,
		"elements": m.Elements(),
	})
}

// WithBody creates a copy of the Menu form and changes its body to the body passed, after which the new Menu
// form is returned. The text is formatted following the rules of fmt.Sprintln.
func (m Menu) WithBody(body ...any) Menu {
	m.body = format(body)
	return m
}

// AddButton appends a button to the menu's element list and returns the updated Menu.
func (m Menu) AddButton(button Button) Menu {
	m.elements = append(m.elements, button)
	return m
}

// AddDivider appends a divider to the menu's element list and returns the updated Menu.
func (m Menu) AddDivider(divider Divider) Menu {
	m.elements = append(m.elements, divider)
	return m
}

// AddHeader appends a header to the menu's element list and returns the updated Menu.
func (m Menu) AddHeader(header Header) Menu {
	m.elements = append(m.elements, header)
	return m
}

// AddLabel appends a label to the menu's element list and returns the updated Menu.
func (m Menu) AddLabel(label Label) Menu {
	m.elements = append(m.elements, label)
	return m
}

// WithButtons creates a copy of the Menu form and appends the buttons passed to the existing elements, after
// which the new Menu form is returned.
func (m Menu) WithButtons(buttons ...Button) Menu {
	for _, b := range buttons {
		m.elements = append(m.elements, b)
	}
	return m
}

// WithElements creates a copy of the Menu form and appends the elements passed to the existing elements, after
// which the new Menu form is returned. This allows adding any MenuElement type.
func (m Menu) WithElements(elements ...MenuElement) Menu {
	m.elements = append(m.elements, elements...)
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

// Buttons returns a list of all buttons of the MenuSubmittable. It collects buttons from the MenuSubmittable
// fields and any buttons added via WithButtons(), AddButton().
func (m Menu) Buttons() []Button {
	v := reflect.New(reflect.TypeOf(m.submittable)).Elem()
	v.Set(reflect.ValueOf(m.submittable))

	buttons := make([]Button, 0)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}
		if b, ok := field.Interface().(Button); ok {
			buttons = append(buttons, b)
		}
	}
	for _, elem := range m.elements {
		if b, ok := elem.(Button); ok {
			buttons = append(buttons, b)
		}
	}
	return buttons
}

// Elements returns all elements of this menu form. It collects elements from the MenuSubmittable
// fields and any elements added via WithElements().
func (m Menu) Elements() []MenuElement {
	v := reflect.New(reflect.TypeOf(m.submittable)).Elem()
	v.Set(reflect.ValueOf(m.submittable))

	elements := make([]MenuElement, 0, v.NumField()+len(m.elements))
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}
		elements = append(elements, field.Interface().(MenuElement))
	}
	elements = append(elements, m.elements...)
	return elements
}

// SubmitJSON submits a JSON value to the menu, containing the index of the button clicked.
func (m Menu) SubmitJSON(b []byte, submitter Submitter, tx *world.Tx) error {
	if b == nil {
		if closer, ok := m.submittable.(Closer); ok {
			closer.Close(submitter, tx)
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
	m.submittable.Submit(submitter, buttons[index], tx)
	return nil
}

// verify verifies if the form is valid, checking all exported fields implement MenuElement.
// It panics if the form is not valid.
func (m Menu) verify() {
	v := reflect.New(reflect.TypeOf(m.submittable)).Elem()
	v.Set(reflect.ValueOf(m.submittable))
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanSet() {
			continue
		}
		if _, ok := v.Field(i).Interface().(MenuElement); !ok {
			panic("all exported fields must implement form.MenuElement")
		}
	}
}

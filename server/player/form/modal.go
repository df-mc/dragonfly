package form

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Modal represents a modal form. These forms have a body with text and two buttons at the end, typically one
// for Yes and one for No. These buttons may have custom text, but can, unlike with a Menu form, not have
// images next to them.
type Modal struct {
	title, body string
	submittable ModalSubmittable
}

// NewModal creates a new Modal form using the ModalSubmittable passed to handle the output of the form. The
// title passed is formatted following the fmt.Sprintln rules.
// Default 'yes' and 'no' buttons may be passed by setting the two exported struct fields of the submittable
// to YesButton() and NoButton() respectively.
func NewModal(submittable ModalSubmittable, title ...any) Modal {
	t := reflect.TypeOf(submittable)
	if t.Kind() != reflect.Struct {
		panic("submittable must be struct")
	}
	m := Modal{title: format(title), submittable: submittable}
	m.verify()
	return m
}

// YesButton returns a Button which may be used as a default 'yes' button for a modal form.
func YesButton() Button {
	return Button{Text: "gui.yes"}
}

// NoButton returns a Button which may be used as a default 'no' button for a modal form.
func NoButton() Button {
	return Button{Text: "gui.no"}
}

// MarshalJSON ...
func (m Modal) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    "modal",
		"title":   m.title,
		"content": m.body,
		"button1": m.Buttons()[0].Text,
		"button2": m.Buttons()[1].Text,
	})
}

// WithBody creates a copy of the Modal form and changes its body to the body passed, after which the new Modal
// form is returned. The text is formatted following the rules of fmt.Sprintln.
func (m Modal) WithBody(body ...any) Modal {
	m.body = format(body)
	return m
}

// Title returns the formatted title passed to the menu upon construction using NewModal().
func (m Modal) Title() string {
	return m.title
}

// Body returns the formatted text in the body passed to the menu using WithBody().
func (m Modal) Body() string {
	return m.body
}

// SubmitJSON submits a JSON byte slice to the modal form. This byte slice contains a JSON encoded bool in it,
// which is used to determine which button was clicked.
func (m Modal) SubmitJSON(b []byte, submitter Submitter) error {
	if b == nil {
		if closer, ok := m.submittable.(Closer); ok {
			closer.Close(submitter)
		}
		return nil
	}

	var value bool
	if err := json.Unmarshal(b, &value); err != nil {
		return fmt.Errorf("error parsing JSON as bool: %w", err)
	}
	if value {
		m.submittable.Submit(submitter, m.Buttons()[0])
		return nil
	}
	m.submittable.Submit(submitter, m.Buttons()[1])
	return nil
}

// Buttons returns a list of all buttons of the Modal form, which will always be a total of two buttons.
func (m Modal) Buttons() []Button {
	v := reflect.New(reflect.TypeOf(m.submittable)).Elem()
	v.Set(reflect.ValueOf(m.submittable))

	buttons := make([]Button, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}
		// Each exported field is guaranteed to be of type Button.
		buttons = append(buttons, field.Interface().(Button))
	}
	return buttons
}

// verify verifies that the Modal form is valid. It checks if exactly two exported fields are present and
// ensures that both have the Button type.
func (m Modal) verify() {
	var count int

	v := reflect.New(reflect.TypeOf(m.submittable)).Elem()
	v.Set(reflect.ValueOf(m.submittable))

	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanSet() {
			continue
		}
		if _, ok := v.Field(i).Interface().(Button); !ok {
			panic("both exported fields must be of the type form.Button")
		}
		count++
	}
	if count != 2 {
		panic("modal form must have exactly two exported fields of the type form.Button")
	}
}

func (Modal) __() {}

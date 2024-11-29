package dialogue

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"reflect"
	"strings"
)

// Dialogue represents a dialogue menu. This menu can consist of a title, a
// body and up to 6 different buttons. The menu also shows a 3D render of the
// entity that is sending the dialogue.
type Dialogue struct {
	title, body string
	submittable Submittable
	buttons     []Button
	display     DisplaySettings
}

// New creates a new Dialogue menu using the Submittable passed to handle the
// dialogue interactions. The title passed is formatted following the rules of
// fmt.Sprintln.
func New(submittable Submittable, title ...any) Dialogue {
	t := reflect.TypeOf(submittable)
	if t.Kind() != reflect.Struct {
		panic("submittable must be struct")
	}
	m := Dialogue{title: format(title), submittable: submittable}
	m.verify()
	return m
}

// MarshalJSON ...
func (m Dialogue) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Buttons())
}

// WithBody creates a copy of the Dialogue and changes its body to the body
// passed, after which the new Dialogue is returned. The text is formatted
// following the rules of fmt.Sprintln.
func (m Dialogue) WithBody(body ...any) Dialogue {
	m.body = format(body)
	return m
}

// WithDisplay returns a new Dialogue with the DisplaySettings passed.
func (m Dialogue) WithDisplay(display DisplaySettings) Dialogue {
	m.display = display
	return m
}

// WithButtons creates a copy of the Dialogue and appends the buttons passed to
// the existing buttons, after which the new Dialogue is returned.
func (m Dialogue) WithButtons(buttons ...Button) Dialogue {
	m.buttons = append(m.buttons, buttons...)
	m.verify()
	return m
}

// Title returns the formatted title passed to the dialogue upon construction
// using New().
func (m Dialogue) Title() string {
	return m.title
}

// Body returns the formatted text in the body passed to the menu using
// WithBody().
func (m Dialogue) Body() string {
	return m.body
}

// Display returns the DisplaySettings of the Dialogue as specified using
// WithDisplay().
func (m Dialogue) Display() DisplaySettings {
	return m.display
}

// Buttons returns a slice of buttons of the Submittable. It parses them from
// the fields using reflection and returns them.
func (m Dialogue) Buttons() []Button {
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

// Submit submits an index of the pressed button to the Submittable. If the
// index is invalid, an error is returned.
func (m Dialogue) Submit(index uint, submitter Submitter, tx *world.Tx) error {
	buttons := m.Buttons()
	if index >= uint(len(buttons)) {
		return fmt.Errorf("button index points to inexistent button: %v (only %v buttons present)", index, len(buttons))
	}
	m.submittable.Submit(submitter, buttons[index], tx)
	return nil
}

// Close closes the dialogue, calling the Close method on the Submittable if it
// implements the Closer interface.
func (m Dialogue) Close(submitter Submitter, tx *world.Tx) {
	if closer, ok := m.submittable.(Closer); ok {
		closer.Close(submitter, tx)
	}
}

// verify verifies if the dialogue is valid, checking all fields are of the
// type Button and there are no more than 6 buttons in total. It panics if the
// dialogue is invalid.
func (m Dialogue) verify() {
	v := reflect.New(reflect.TypeOf(m.submittable)).Elem()
	v.Set(reflect.ValueOf(m.submittable))
	var buttons int
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanSet() {
			continue
		}
		if _, ok := v.Field(i).Interface().(Button); !ok {
			panic("all exported fields must be of the type dialogue.Button")
		}
		buttons++
	}
	if buttons+len(m.buttons) > 6 {
		panic("maximum of 6 buttons allowed")
	}
}

// format is a utility function to format a list of values to have spaces
// between them, but no newline at the end.
func format(a []any) string {
	return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintln(a...), "\n"), "\n")
}

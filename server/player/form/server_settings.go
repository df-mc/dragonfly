package form

import (
	"reflect"

	"github.com/go-jose/go-jose/v4/json"
)

// ServerSettings represents a form that may be sent to a player when they open their settings.
type ServerSettings struct {
	// Image holds the image displayed on the button.
	Image Image
	// We inherit other properties from Custom forms, since they are identical.
	Custom
}

// NewServerSettings creates a new (server settings) form with the title passed and returns it. The title is
// formatted according to the rules of fmt.Sprintln.
// The submittable passed is used to create the structure of the form. The values of the Submittable's form
// fields are used to set text, defaults and placeholders. If the Submittable passed is not a struct, New
// panics. NewServerSettings also panics if one of the exported field types of the Submittable is not one
// that implements the Element interface.
func NewServerSettings(submittable Submittable, title ...any) ServerSettings {
	t := reflect.TypeOf(submittable)
	if t.Kind() != reflect.Struct {
		panic("submittable must be struct")
	}
	f := ServerSettings{}
	f.title = format(title)
	f.submittable = submittable
	f.verify()
	return f
}

// MarshalJSON ...
func (f ServerSettings) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"type":    "custom_form",
		"title":   f.title,
		"content": f.Elements(),
	}
	if f.Image.Path != "" {
		m["image"] = f.Image
	}
	return json.Marshal(m)
}

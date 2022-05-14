package form

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
)

// Form represents a form that may be sent to a Submitter. The three types of forms, custom forms, menu forms
// and modal forms implement this interface.
type Form interface {
	json.Marshaler
	SubmitJSON(b []byte, submitter Submitter) error
	__()
}

// Custom represents a form that may be sent to a player and has fields that should be filled out by the
// player that the form is sent to.
type Custom struct {
	title       string
	submittable Submittable
}

// MarshalJSON ...
func (f Custom) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    "custom_form",
		"title":   f.title,
		"content": f.Elements(),
	})
}

// New creates a new (custom) form with the title passed and returns it. The title is formatted according to
// the rules of fmt.Sprintln.
// The submittable passed is used to create the structure of the form. The values of the Submittable's form
// fields are used to set text, defaults and placeholders. If the Submittable passed is not a struct, New
// panics. New also panics if one of the exported field types of the Submittable is not one that implements
// the Element interface.
func New(submittable Submittable, title ...any) Custom {
	t := reflect.TypeOf(submittable)
	if t.Kind() != reflect.Struct {
		panic("submittable must be struct")
	}
	f := Custom{title: format(title), submittable: submittable}
	f.verify()
	return f
}

// Title returns the formatted title passed when the form was created using New().
func (f Custom) Title() string {
	return f.title
}

// Elements returns a list of all elements as set in the Submittable passed to form.New().
func (f Custom) Elements() []Element {
	v := reflect.New(reflect.TypeOf(f.submittable)).Elem()
	v.Set(reflect.ValueOf(f.submittable))
	n := v.NumField()

	elements := make([]Element, 0, n)
	for i := 0; i < n; i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}
		// Each exported field is guaranteed to implement the Element interface.
		elements = append(elements, field.Interface().(Element))
	}
	return elements
}

// SubmitJSON submits a JSON data slice to the form. The form will check all values in the JSON array passed,
// making sure their values are valid for the form's elements.
// If the values are valid and can be parsed properly, the Submittable.Submit() method of the form's Submittable is
// called and the fields of the Submittable will be filled out.
func (f Custom) SubmitJSON(b []byte, submitter Submitter) error {
	if b == nil {
		if closer, ok := f.submittable.(Closer); ok {
			closer.Close(submitter)
		}
		return nil
	}

	dec := json.NewDecoder(bytes.NewBuffer(b))
	dec.UseNumber()

	var data []any
	if err := dec.Decode(&data); err != nil {
		return fmt.Errorf("error decoding JSON data to slice: %w", err)
	}

	v := reflect.New(reflect.TypeOf(f.submittable)).Elem()
	v.Set(reflect.ValueOf(f.submittable))

	for i := 0; i < v.NumField(); i++ {
		fieldV := v.Field(i)
		if !fieldV.CanSet() {
			continue
		}
		if len(data) == 0 {
			return fmt.Errorf("form JSON data array does not have enough values")
		}
		elem, err := f.parseValue(fieldV.Interface().(Element), data[0])
		if err != nil {
			return fmt.Errorf("error parsing form response value: %w", err)
		}
		fieldV.Set(elem)
		data = data[1:]
	}

	v.Interface().(Submittable).Submit(submitter)

	return nil
}

// parseValue parses a value into the Element passed and returns it as a reflection Value. If the value is not
// valid for the element, an error is returned.
func (f Custom) parseValue(elem Element, s any) (reflect.Value, error) {
	var ok bool
	var value reflect.Value

	switch element := elem.(type) {
	case Label:
		value = reflect.ValueOf(element)
	case Input:
		element.value, ok = s.(string)
		if !ok {
			return value, fmt.Errorf("value %v is not allowed for input element", s)
		}
		if !utf8.ValidString(element.value) {
			return value, fmt.Errorf("value %v is not valid UTF8", s)
		}
		value = reflect.ValueOf(element)
	case Toggle:
		element.value, ok = s.(bool)
		if !ok {
			return value, fmt.Errorf("value %v is not allowed for toggle element", s)
		}
		value = reflect.ValueOf(element)
	case Slider:
		v, ok := s.(json.Number)
		f, err := v.Float64()
		if !ok || err != nil {
			return value, fmt.Errorf("value %v is not allowed for slider element", s)
		}
		if f > element.Max || f < element.Min {
			return value, fmt.Errorf("slider value %v is out of range %v-%v", f, element.Min, element.Max)
		}
		element.value = f
		value = reflect.ValueOf(element)
	case Dropdown:
		v, ok := s.(json.Number)
		f, err := v.Int64()
		if !ok || err != nil {
			return value, fmt.Errorf("value %v is not allowed for dropdown element", s)
		}
		if f < 0 || int(f) >= len(element.Options) {
			return value, fmt.Errorf("dropdown value %v is out of range %v-%v", f, 0, len(element.Options)-1)
		}
		element.value = int(f)
		value = reflect.ValueOf(element)
	case StepSlider:
		v, ok := s.(json.Number)
		f, err := v.Int64()
		if !ok || err != nil {
			return value, fmt.Errorf("value %v is not allowed for dropdown element", s)
		}
		if f < 0 || int(f) >= len(element.Options) {
			return value, fmt.Errorf("dropdown value %v is out of range %v-%v", f, 0, len(element.Options)-1)
		}
		element.value = int(f)
		value = reflect.ValueOf(element)
	}
	return value, nil
}

// verify verifies if the form is valid, checking if the fields all implement the Element interface. It panics
// if the form is not valid.
func (f Custom) verify() {
	el := reflect.TypeOf((*Element)(nil)).Elem()

	v := reflect.New(reflect.TypeOf(f.submittable)).Elem()
	v.Set(reflect.ValueOf(f.submittable))

	t := reflect.TypeOf(f.submittable)
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanSet() {
			continue
		}
		if !t.Field(i).Type.Implements(el) {
			panic("all exported fields must implement form.Element interface")
		}
	}
}

// format is a utility function to format a list of values to have spaces between them, but no newline at the
// end.
func format(a []any) string {
	return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintln(a...), "\n"), "\n")
}

func (f Custom) __() {}

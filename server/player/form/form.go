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

type ResponseData = interface{}

// Custom represents a form that may be sent to a player and has fields that should be filled out by the
// player that the form is sent to.
type Custom[T ResponseData] struct {
	title    string
	elements []Element
	data     T
	onClose  Handler
	onSubmit func(Submitter, T)
}

// MarshalJSON ...
func (f Custom[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    "custom_form",
		"title":   f.title,
		"content": f.Elements(),
	})
}

// NewCustom creates a new (custom) form with the title passed and returns it. The title is formatted according to
// the rules of fmt.Sprintln.
// The data passed is used to create the structure of the form. The fields of the data struct
// are used to filled by parsed response data. If the data passed is not a struct, NewCustom
// panics.
func NewCustom[T ResponseData](data T, title ...any) Custom[T] {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Struct {
		panic(fmt.Sprintf("invalid kind of data: %v", v.Kind().String()))
	}
	f := Custom[T]{title: format(title), data: data}
	f.verify()
	return f
}

// WithElements creates a copy of the Custom form and append the elements passed.
func (f Custom[T]) WithElements(elem ...Element) Custom[T] {
	f.elements = append(f.elements, elem...)
	return f
}

// OnClose creates a copy of the Menu form and set the form close callback to the passed one.
func (f Custom[T]) OnClose(c Handler) Custom[T] {
	f.onClose = c
	return f
}

// OnSubmit creates a copy of the Menu form and set the form submit callback to the passed one.
func (f Custom[T]) OnSubmit(c func(Submitter, T)) Custom[T] {
	f.onSubmit = c
	return f
}

// Title returns the formatted title passed when the form was created using NewCustom().
func (f Custom[T]) Title() string {
	return f.title
}

// Elements returns a list of all elements passed in WithElements.
func (f Custom[T]) Elements() []Element {
	return f.elements
}

// SubmitJSON submits a JSON data slice to the form. The form will check all values in the JSON array passed,
// making sure their values are valid for the form's elements.
// If the values are valid and can be parsed properly, the fields of the data will be filled out, and
// the onSubmit callback will be called.
func (f Custom[T]) SubmitJSON(b []byte, submitter Submitter) error {
	if b == nil {
		f.onClose.Call(submitter)
		return nil
	}

	dec := json.NewDecoder(bytes.NewBuffer(b))
	dec.UseNumber()

	var data []any
	if err := dec.Decode(&data); err != nil {
		return fmt.Errorf("error decoding JSON data to slice: %w", err)
	}

	elem := f.Elements()
	v := reflect.ValueOf(&f.data).Elem()

	for i := 0; i < v.NumField() && i < len(elem); i++ {
		fieldV := v.Field(i)
		if !fieldV.CanSet() {
			continue
		}
		if len(data) == 0 {
			return fmt.Errorf("form JSON data array does not have enough values")
		}
		val, hasValue, err := f.parseValue(elem[i], data[0])
		if !hasValue {
			continue
		}
		if err != nil {
			return fmt.Errorf("error parsing form response value: %w", err)
		}
		fieldV.Set(reflect.ValueOf(val))
		data = data[1:]
	}

	if f.onSubmit != nil {
		f.onSubmit(submitter, f.data)
	}
	return nil
}

// parseValue parses a value into the Element passed and returns it as a parsed Value. If the value is not
// valid for the element, no value, an error is returned.
func (f Custom[T]) parseValue(elem Element, s any) (interface{}, bool, error) {
	var ok bool
	var value interface{}

	switch element := elem.(type) {
	case Label:
		value = nil
		return value, false, nil
	case Input:
		value, ok = s.(string)
		if !ok {
			return value, false, fmt.Errorf("value %v is not allowed for input element", s)
		}
		if !utf8.ValidString(value.(string)) {
			return value, false, fmt.Errorf("value %v is not valid UTF8", s)
		}
	case Toggle:
		value, ok = s.(bool)
		if !ok {
			return value, false, fmt.Errorf("value %v is not allowed for toggle element", s)
		}
	case Slider:
		v, ok := s.(json.Number)
		f, err := v.Float64()
		if !ok || err != nil {
			return value, false, fmt.Errorf("value %v is not allowed for slider element", s)
		}
		if f > element.Max || f < element.Min {
			return value, false, fmt.Errorf("slider value %v is out of range %v-%v", f, element.Min, element.Max)
		}
		value = f
	case Dropdown:
		v, ok := s.(json.Number)
		f, err := v.Int64()
		if !ok || err != nil {
			return value, false, fmt.Errorf("value %v is not allowed for dropdown element", s)
		}
		if f < 0 || int(f) >= len(element.Options) {
			return value, false, fmt.Errorf("dropdown value %v is out of range %v-%v", f, 0, len(element.Options)-1)
		}
		value = element.Options[f]
	case StepSlider:
		v, ok := s.(json.Number)
		f, err := v.Int64()
		if !ok || err != nil {
			return value, false, fmt.Errorf("value %v is not allowed for dropdown element", s)
		}
		if f < 0 || int(f) >= len(element.Options) {
			return value, false, fmt.Errorf("dropdown value %v is out of range %v-%v", f, 0, len(element.Options)-1)
		}
		value = element.Options[f]
	}
	return value, true, nil
}

// verify verifies if the form is valid, checking if the fields all implement the Element interface. It panics
// if the form is not valid.
func (f Custom[T]) verify() {
	//TODO
}

// format is a utility function to format a list of values to have spaces between them, but no newline at the
// end.
func format(a []any) string {
	return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintln(a...), "\n"), "\n")
}

func (f Custom[T]) __() {}

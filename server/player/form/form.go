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
	title    string
	elements []Element
	onClose  Handler
	onSubmit *reflect.Value
}

// MarshalJSON ...
func (f Custom) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    "custom_form",
		"title":   f.title,
		"content": f.Elements(),
	})
}

// NewCustom creates a new (custom) form with the title passed and returns it. The title is formatted according to
// the rules of fmt.Sprintln.
func NewCustom(title ...any) Custom {
	f := Custom{title: format(title)}
	return f
}

// WithElements creates a copy of the Custom form and append the elements passed.
func (f Custom) WithElements(elem ...Element) Custom {
	f.elements = append(f.elements, elem...)
	return f
}

// OnClose creates a copy of the Menu form and set the form close callback to the passed one.
func (f Custom) OnClose(c Handler) Custom {
	f.onClose = c
	return f
}

// OnSubmit creates a copy of the Menu form and set the form submit callback to the passed one.
// will panic if passing c isn't a func
func (f Custom) OnSubmit(c interface{}) Custom {
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Func {
		panic("passing a invalid func")
	}
	f.onSubmit = &v
	f.verify()
	return f
}

// Title returns the formatted title passed when the form was created using NewCustom().
func (f Custom) Title() string {
	return f.title
}

// Elements returns a list of all elements passed in WithElements.
func (f Custom) Elements() []Element {
	return f.elements
}

// SubmitJSON submits a JSON data slice to the form. The form will check all values in the JSON array passed,
// making sure their values are valid for the form's elements.
// If the values are valid and can be parsed properly, the fields of the data will be filled out, and
// the onSubmit callback will be called.
func (f Custom) SubmitJSON(b []byte, submitter Submitter) error {
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
	params := []reflect.Value{reflect.ValueOf(submitter)}

	for i := 0; i < len(elem); i++ {
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
		params = append(params, reflect.ValueOf(val))
		data = data[1:]
	}

	if f.onSubmit.Type().NumIn() != len(params) {
		return fmt.Errorf("error form response data: %v parsed, expected %v", len(params), f.onSubmit.Type().NumIn())
	}

	if f.onSubmit != nil {
		f.onSubmit.Call(params)
	}
	return nil
}

// parseValue parses a value into the Element passed and returns it as a parsed Value. If the value is not
// valid for the element, no value, an error is returned.
func (f Custom) parseValue(elem Element, s any) (interface{}, bool, error) {
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

// validateParam validate a param from onSubmit by the Element passed.
// if the param is valid for the element, true will be returned, or else false.
// will panic if the element is invalid for example button (unavailable for Custom)
func (f Custom) validateParam(param reflect.Type, elem Element) (valid bool) {
	switch elem.(type) {
	case Label:
		return true
	case Input:
		return param.Kind() == reflect.String
	case Toggle:
		return param.Kind() == reflect.Bool
	case Slider:
		return param.Kind() == reflect.Float64
	case Dropdown:
		return param.Kind() == reflect.String
	case StepSlider:
		return param.Kind() == reflect.String
	}
	return false
}

// verify verifies if the form is valid, checking if the fields all implement the Element interface. It panics
// if the form is not valid.
func (f Custom) verify() {
	elems := f.Elements()
	cal := f.onSubmit
	calType := cal.Type()
	submitterType := reflect.TypeOf((*Submitter)(nil)).Elem()

	if calType.NumIn() < 1 {
		panic(fmt.Errorf("no param given in OnSubmit"))
	}
	if !calType.In(0).Implements(submitterType) {
		panic(fmt.Errorf("invalid submitter type: %v", calType.In(0).String()))
	}
	elemIndex := 0
	validParamCount := 0
	for i := 1; i < calType.NumIn(); i++ {
		in := calType.In(i)
		if elemIndex >= len(elems) {
			panic(fmt.Errorf("mismatched params given in OnSubmit"))
		}
		for !elems[elemIndex].haveData() {
			elemIndex++
		}
		valid := f.validateParam(in, elems[elemIndex])
		if !valid {
			panic(fmt.Errorf("invalid param %v(%v) for element %T", i, in.String(), elems[elemIndex]))
		}
		validParamCount++
		elemIndex++
	}
	if validParamCount != calType.NumIn()-1 {
		panic(fmt.Errorf("mismatched params given in OnSubmit: expected %v given %v", validParamCount, calType.NumIn()-1))
	}
}

// format is a utility function to format a list of values to have spaces between them, but no newline at the
// end.
func format(a []any) string {
	return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintln(a...), "\n"), "\n")
}

func (f Custom) __() {}

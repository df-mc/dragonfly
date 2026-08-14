package form

import (
	"encoding/json"
	"strings"
)

// Element represents an element that may be added to a Form. Any of the types in this package that implement
// the element interface may be used as struct fields when passing the form structure to form.New().
type Element interface {
	json.Marshaler
	elem()
}

// MenuElement represents an element that may be added to a Menu form. This includes buttons, dividers,
// headers, and labels.
type MenuElement interface {
	json.Marshaler
	menuElem()
}

// Divider represents a visual separator element on a form. It displays a horizontal line.
type Divider struct{}

// MarshalJSON ...
func (d Divider) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type": "divider",
		"text": "",
	})
}

// Header represents a header element on a form. It displays larger, emphasised text for section titles.
type Header struct {
	// Text is the text held by the header. The text may contain Minecraft formatting codes.
	Text string
}

// NewHeader creates and returns a new Header with the text passed.
func NewHeader(text string) Header {
	return Header{Text: text}
}

// MarshalJSON ...
func (h Header) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type": "header",
		"text": h.Text,
	})
}

// Label represents a static label on a form. It serves only to display a box of text, and users cannot
// submit values to it.
type Label struct {
	// Text is the text held by the label. The text may contain Minecraft formatting codes.
	Text string
}

// NewLabel creates and returns a new Label with the values passed.
func NewLabel(text string) Label {
	return Label{Text: text}
}

// MarshalJSON ...
func (l Label) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type": "label",
		"text": l.Text,
	})
}

// Input represents a text input box element. Submitters may write any text in these boxes with no specific
// length.
type Input struct {
	// Text is the text displayed over the input element. The text may contain Minecraft formatting codes.
	Text string
	// Default is the default value filled out in the input. The user may remove this value and fill out its
	// own text. The text may contain Minecraft formatting codes.
	Default string
	// Placeholder is the text displayed in the input box if it does not contain any text filled out by the
	// user. The text may contain Minecraft formatting codes.
	Placeholder string
	// Tooltip is an optional text displayed when hovering over the element's info icon. The icon only
	// appears when a tooltip is set.
	Tooltip string

	value string
}

// NewInput creates and returns a new Input with the values passed.
func NewInput(text, defaultValue, placeholder string) Input {
	return Input{Text: text, Default: defaultValue, Placeholder: placeholder}
}

// WithTooltip returns a copy of the Input with the tooltip set.
func (i Input) WithTooltip(tooltip string) Input {
	i.Tooltip = tooltip
	return i
}

// MarshalJSON ...
func (i Input) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"type":        "input",
		"text":        i.Text,
		"default":     i.Default,
		"placeholder": i.Placeholder,
	}
	if i.Tooltip != "" {
		m["tooltip"] = i.Tooltip
	}
	return json.Marshal(m)
}

// Value returns the value filled out by the user.
func (i Input) Value() string {
	return i.value
}

// Toggle represents an on-off button element. Submitters may either toggle this on or off, which will then
// hold a value of true or false respectively.
type Toggle struct {
	// Text is the text displayed over the toggle element. The text may contain Minecraft formatting codes.
	Text string
	// Default determines if the toggle should be on/off by default.
	Default bool
	// Tooltip is an optional text displayed when hovering over the element's info icon. The icon only
	// appears when a tooltip is set.
	Tooltip string

	value bool
}

// NewToggle creates and returns a new Toggle with the values passed.
func NewToggle(text string, defaultValue bool) Toggle {
	return Toggle{Text: text, Default: defaultValue}
}

// WithTooltip returns a copy of the Toggle with the tooltip set.
func (t Toggle) WithTooltip(tooltip string) Toggle {
	t.Tooltip = tooltip
	return t
}

// MarshalJSON ...
func (t Toggle) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"type":    "toggle",
		"text":    t.Text,
		"default": t.Default,
	}
	if t.Tooltip != "" {
		m["tooltip"] = t.Tooltip
	}
	return json.Marshal(m)
}

// Value returns the value filled out by the user.
func (t Toggle) Value() bool {
	return t.value
}

// Slider represents a slider element. Submitters may move the slider to values within the range of the slider
// to select a value.
type Slider struct {
	// Text is the text displayed over the slider element. The text may contain Minecraft formatting codes.
	Text string
	// Min and Max are used to specify the minimum and maximum range of the slider. A value lower or higher
	// than these values cannot be selected.
	Min, Max float64
	// StepSize is the size that one step of the slider takes up. When set to 1.0 for example, a submitter
	// will be able to select only whole values.
	StepSize float64
	// Default is the default value filled out for the slider.
	Default float64
	// Tooltip is an optional text displayed when hovering over the element's info icon. The icon only
	// appears when a tooltip is set.
	Tooltip string

	value float64
}

// NewSlider creates and returns a new Slider using the values passed.
func NewSlider(text string, min, max, stepSize, defaultValue float64) Slider {
	return Slider{Text: text, Min: min, Max: max, StepSize: stepSize, Default: defaultValue}
}

// WithTooltip returns a copy of the Slider with the tooltip set.
func (s Slider) WithTooltip(tooltip string) Slider {
	s.Tooltip = tooltip
	return s
}

// MarshalJSON ...
func (s Slider) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"type":    "slider",
		"text":    s.Text,
		"min":     s.Min,
		"max":     s.Max,
		"step":    s.StepSize,
		"default": s.Default,
	}
	if s.Tooltip != "" {
		m["tooltip"] = s.Tooltip
	}
	return json.Marshal(m)
}

// Value returns the value filled out by the user.
func (s Slider) Value() float64 {
	return s.value
}

// Dropdown represents a dropdown which, when clicked, opens a window with the options set in the Options
// field. Submitters may select one of the options.
type Dropdown struct {
	// Text is the text displayed over the dropdown element. The text may contain Minecraft formatting codes.
	Text string
	// Options holds a list of options that a Submitter may select. The order of these options is retained
	// when shown to the submitter of the form.
	Options []string
	// DefaultIndex is the index in the Options slice that is used as default. When sent to a Submitter, the
	// value at this index in the Options slice will be selected.
	DefaultIndex int
	// Tooltip is an optional text displayed when hovering over the element's info icon. The icon only
	// appears when a tooltip is set.
	Tooltip string

	value int
}

// NewDropdown creates and returns new Dropdown using the values passed.
func NewDropdown(text string, options []string, defaultIndex int) Dropdown {
	return Dropdown{Text: text, Options: options, DefaultIndex: defaultIndex}
}

// WithTooltip returns a copy of the Dropdown with the tooltip set.
func (d Dropdown) WithTooltip(tooltip string) Dropdown {
	d.Tooltip = tooltip
	return d
}

// MarshalJSON ...
func (d Dropdown) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"type":    "dropdown",
		"text":    d.Text,
		"default": d.DefaultIndex,
		"options": d.Options,
	}
	if d.Tooltip != "" {
		m["tooltip"] = d.Tooltip
	}
	return json.Marshal(m)
}

// Value returns the value that the Submitter submitted. The value is an index pointing to the selected option
// in the Options slice.
func (d Dropdown) Value() int {
	return d.value
}

// StepSlider represents a slider that has a number of options that may be selected. It is essentially a
// combination of a Dropdown and a Slider, looking like a slider but having properties like a dropdown.
type StepSlider Dropdown

// NewStepSlider creates and returns new StepSlider using the values passed.
func NewStepSlider(text string, options []string, defaultIndex int) StepSlider {
	return StepSlider{Text: text, Options: options, DefaultIndex: defaultIndex}
}

// WithTooltip returns a copy of the StepSlider with the tooltip set.
func (s StepSlider) WithTooltip(tooltip string) StepSlider {
	s.Tooltip = tooltip
	return s
}

// MarshalJSON ...
func (s StepSlider) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"type":    "step_slider",
		"text":    s.Text,
		"default": s.DefaultIndex,
		"steps":   s.Options,
	}
	if s.Tooltip != "" {
		m["tooltip"] = s.Tooltip
	}
	return json.Marshal(m)
}

// Value returns the value that the Submitter submitted. The value is an index pointing to the selected option
// in the Options slice.
func (s StepSlider) Value() int {
	return s.value
}

// Button represents a button added to a Menu or Modal form. The button has text on it and an optional image,
// which may be either retrieved from a website or the local assets of the game.
type Button struct {
	// Text holds the text displayed on the button. It may use Minecraft formatting codes and may have
	// newlines.
	Text string
	// Image holds a path to an image for the button. The Image may either be a URL pointing to an image,
	// such as 'https://someimagewebsite.com/someimage.png', or a path pointing to a local asset, such as
	// 'textures/blocks/grass_carried'.
	Image string
}

// NewButton creates and returns a new Button using the text and image passed.
func NewButton(text, image string) Button {
	return Button{Text: text, Image: image}
}

// MarshalJSON ...
func (b Button) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"type": "button",
		"text": b.Text,
	}
	if b.Image != "" {
		buttonType := "path"
		if strings.HasPrefix(b.Image, "http:") || strings.HasPrefix(b.Image, "https:") {
			buttonType = "url"
		}
		m["image"] = map[string]any{"type": buttonType, "data": b.Image}
	}
	return json.Marshal(m)
}

func (Divider) elem()    {}
func (Header) elem()     {}
func (Label) elem()      {}
func (Input) elem()      {}
func (Toggle) elem()     {}
func (Slider) elem()     {}
func (Dropdown) elem()   {}
func (StepSlider) elem() {}

func (Divider) menuElem() {}
func (Header) menuElem()  {}
func (Label) menuElem()   {}
func (Button) menuElem()  {}

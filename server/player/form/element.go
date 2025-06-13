package form

import (
	"encoding/json"
	"strings"
)

// Element represents an element that may be added to a Form. Any of the types in this package that implement
// the element interface may be used as struct fields when passing the form structure to form.New().
type Element interface {
	json.Marshaler
	ReadOnly() bool
	elem()
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

func (Label) ReadOnly() bool {
	return true
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

	value string
}

// NewInput creates and returns a new Input with the values passed.
func NewInput(text, defaultValue, placeholder string) Input {
	return Input{Text: text, Default: defaultValue, Placeholder: placeholder}
}

// MarshalJSON ...
func (i Input) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":        "input",
		"text":        i.Text,
		"default":     i.Default,
		"placeholder": i.Placeholder,
	})
}

// Value returns the value filled out by the user.
func (i Input) Value() string {
	return i.value
}

func (Input) ReadOnly() bool {
	return false
}

// Toggle represents an on-off button element. Submitters may either toggle this on or off, which will then
// hold a value of true or false respectively.
type Toggle struct {
	// Text is the text displayed over the toggle element. The text may contain Minecraft formatting codes.
	Text string
	// Default determines if the toggle should be on/off by default.
	Default bool

	value bool
}

// NewToggle creates and returns a new Toggle with the values passed.
func NewToggle(text string, defaultValue bool) Toggle {
	return Toggle{Text: text, Default: defaultValue}
}

// MarshalJSON ...
func (t Toggle) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    "toggle",
		"text":    t.Text,
		"default": t.Default,
	})
}

// Value returns the value filled out by the user.
func (t Toggle) Value() bool {
	return t.value
}

func (Toggle) ReadOnly() bool {
	return false
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

	value float64
}

// NewSlider creates and returns a new Slider using the values passed.
func NewSlider(text string, min, max, stepSize, defaultValue float64) Slider {
	return Slider{Text: text, Min: min, Max: max, StepSize: stepSize, Default: defaultValue}
}

// MarshalJSON ...
func (s Slider) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    "slider",
		"text":    s.Text,
		"min":     s.Min,
		"max":     s.Max,
		"step":    s.StepSize,
		"default": s.Default,
	})
}

// Value returns the value filled out by the user.
func (s Slider) Value() float64 {
	return s.value
}

func (Slider) ReadOnly() bool {
	return false
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

	value int
}

// NewDropdown creates and returns new Dropdown using the values passed.
func NewDropdown(text string, options []string, defaultIndex int) Dropdown {
	return Dropdown{Text: text, Options: options, DefaultIndex: defaultIndex}
}

// MarshalJSON ...
func (d Dropdown) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    "dropdown",
		"text":    d.Text,
		"default": d.DefaultIndex,
		"options": d.Options,
	})
}

// Value returns the value that the Submitter submitted. The value is an index pointing to the selected option
// in the Options slice.
func (d Dropdown) Value() int {
	return d.value
}

func (Dropdown) ReadOnly() bool {
	return false
}

// StepSlider represents a slider that has a number of options that may be selected. It is essentially a
// combination of a Dropdown and a Slider, looking like a slider but having properties like a dropdown.
type StepSlider Dropdown

// NewStepSlider creates and returns new StepSlider using the values passed.
func NewStepSlider(text string, options []string, defaultIndex int) StepSlider {
	return StepSlider{Text: text, Options: options, DefaultIndex: defaultIndex}
}

// MarshalJSON ...
func (s StepSlider) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    "step_slider",
		"text":    s.Text,
		"default": s.DefaultIndex,
		"steps":   s.Options,
	})
}

// Value returns the value that the Submitter submitted. The value is an index pointing to the selected option
// in the Options slice.
func (s StepSlider) Value() int {
	return s.value
}

func (StepSlider) ReadOnly() bool {
	return false
}

const (
	ImageTypePath  = "path"
	ImageTypeImage = "image"
)

// Image represents an image attached to a form element such as a button.
type Image struct {
	// Type is the type of the image. It is always one of the constants above. If left empty, it will be inferred.
	Type string
	// Path holds the path to the image. It can be a URL like "https://example.com/image.png"
	// or a local path like "textures/blocks/grass_carried".
	Path string
}

// NewImage creates and returns new Image using the values passed.
func NewImage(path string) Image {
	return Image{Path: path}
}

// MarshalJSON ...
func (img Image) MarshalJSON() ([]byte, error) {
	imgType := img.Type
	if imgType == "" {
		imgType = ImageTypePath
		if strings.HasPrefix(img.Path, "http:") || strings.HasPrefix(img.Path, "https:") {
			imgType = ImageTypeImage
		}
	}
	return json.Marshal(map[string]any{
		"type": imgType,
		"data": img.Path,
	})
}

// Button represents a button added to a Menu or Modal form. The button has text on it and an optional image,
// which may be either retrieved from a website or the local assets of the game.
type Button struct {
	// Text holds the text displayed on the button. It may use Minecraft formatting codes and may have
	// newlines.
	Text string
	// Image holds the image displayed on the button.
	Image Image
}

// NewButton creates and returns a new Button using the text and image passed.
func NewButton(text string, image Image) Button {
	return Button{Text: text, Image: image}
}

// MarshalJSON ...
func (b Button) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"text":  b.Text,
		"image": b.Image,
	})
}

func (Label) elem()      {}
func (Input) elem()      {}
func (Toggle) elem()     {}
func (Slider) elem()     {}
func (Dropdown) elem()   {}
func (StepSlider) elem() {}

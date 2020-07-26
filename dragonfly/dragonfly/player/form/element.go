package form

// Element represents an element that may be added to a Form. Any of the types in this package that implement
// the element interface may be used as struct fields when passing the form structure to form.New().
type Element interface {
	__()
}

// Label represents a static label on a form. It serves only to display a box of text, and users cannot
// submit values to it.
type Label struct {
	// Text is the text held by the label. The text may contain Minecraft formatting codes.
	Text string
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

// Value returns the value filled out by the user.
func (i Input) Value() string {
	return i.value
}

// Toggle represents an on-off button element. Submitters may either toggle this on or off, which will then
// hold a value of true or false respectively.
type Toggle struct {
	// Text is the text displayed over the toggle element. The text may contain Minecraft formatting codes.
	Text string
	// Default is the default value filled out in the input. The user may remove this value and fill out its
	// own text. The text may contain Minecraft formatting codes.
	Default bool

	value bool
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

	value float64
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

	value int
}

// Value returns the value that the Submitter submitted. The value is an index pointing to the selected option
// in the Options slice.
func (d Dropdown) Value() int {
	return d.value
}

// StepSlider represents a slider that has a number of options that may be selected. It is essentially a
// combination of a Dropdown and a Slider, looking like a slider but having properties like a dropdown.
type StepSlider Dropdown

// Value returns the value that the Submitter submitted. The value is an index pointing to the selected option
// in the Options slice.
func (s StepSlider) Value() int {
	return s.value
}

func (Label) __()      {}
func (Input) __()      {}
func (Toggle) __()     {}
func (Slider) __()     {}
func (Dropdown) __()   {}
func (StepSlider) __() {}

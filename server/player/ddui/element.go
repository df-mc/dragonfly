package ddui

// DropdownOption is a selectable item in a Dropdown element.
type DropdownOption struct {
	Label string
	Value int
}

// TextFieldOption configures optional properties of a TextField element.
type TextFieldOption struct{ description string }

// WithDescription sets the description tooltip for a TextField.
func WithDescription(description string) TextFieldOption {
	return TextFieldOption{description: description}
}

// SliderOption configures optional properties of a Slider element.
type SliderOption struct {
	description string
	step        float64
}

// WithSliderDescription sets the description tooltip for a Slider.
func WithSliderDescription(description string) SliderOption {
	return SliderOption{description: description}
}

// WithStep sets the step increment for a Slider.
func WithStep(step float64) SliderOption {
	return SliderOption{step: step}
}

// element is the internal interface implemented by all CustomForm element types.
type element interface {
	describe() ElementDescriptor
	handleUpdate(property string, value UpdateValue)
	bindSend(pathPrefix string, fn func(UpdateNotification))
}

type spacerElement struct{}

func (s spacerElement) describe() ElementDescriptor                   { return ElementDescriptor{Kind: ElementSpacer} }
func (s spacerElement) handleUpdate(_ string, _ UpdateValue)          {}
func (s spacerElement) bindSend(_ string, _ func(UpdateNotification)) {}
func (s spacerElement) applyForm(f *CustomForm)                       { f.elements = append(f.elements, s) }

// Spacer adds a blank vertical spacer element.
func Spacer() FormOption { return spacerElement{} }

type dividerElement struct{}

func (d dividerElement) describe() ElementDescriptor                   { return ElementDescriptor{Kind: ElementDivider} }
func (d dividerElement) handleUpdate(_ string, _ UpdateValue)          {}
func (d dividerElement) bindSend(_ string, _ func(UpdateNotification)) {}
func (d dividerElement) applyForm(f *CustomForm)                       { f.elements = append(f.elements, d) }

// Divider adds a horizontal divider line element.
func Divider() FormOption { return dividerElement{} }

type labelElement struct{ text *Observable[string] }

func (l *labelElement) describe() ElementDescriptor {
	return ElementDescriptor{Kind: ElementLabel, StringValue: l.text.Get()}
}
func (l *labelElement) handleUpdate(_ string, _ UpdateValue) {}
func (l *labelElement) bindSend(pathPrefix string, fn func(UpdateNotification)) {
	l.text.bindSend(func(v string) {
		fn(UpdateNotification{Path: pathPrefix + ".text", Value: UpdateValue{Kind: UpdateKindString, String: v}})
	})
}
func (l *labelElement) applyForm(f *CustomForm) { f.elements = append(f.elements, l) }

// Label adds a text label element with a static value.
func Label(text string) FormOption { return &labelElement{text: NewObservable(text, false)} }

// LabelObs adds a text label element bound to obs. Calling obs.Set updates the label on screen immediately.
func LabelObs(obs *Observable[string]) FormOption { return &labelElement{text: obs} }

type textFieldElement struct {
	label, description string
	value              *Observable[string]
}

func (t *textFieldElement) describe() ElementDescriptor {
	return ElementDescriptor{
		Kind:        ElementTextField,
		Label:       t.label,
		Description: t.description,
		StringValue: t.value.Get(),
	}
}

func (t *textFieldElement) handleUpdate(property string, u UpdateValue) {
	if property == "text" && t.value.clientWritable {
		t.value.update(u.String)
	}
}

func (t *textFieldElement) bindSend(pathPrefix string, fn func(UpdateNotification)) {
	t.value.bindSend(func(v string) {
		fn(UpdateNotification{
			Path:  pathPrefix + ".text",
			Value: UpdateValue{Kind: UpdateKindString, String: v},
		})
	})
}

func (t *textFieldElement) applyForm(f *CustomForm) { f.elements = append(f.elements, t) }

// TextField adds a text input element bound to value.
func TextField(label string, value *Observable[string], opts ...TextFieldOption) FormOption {
	e := &textFieldElement{label: label, value: value}
	for _, o := range opts {
		if o.description != "" {
			e.description = o.description
		}
	}
	return e
}

type dropdownElement struct {
	label   string
	value   *Observable[int]
	options []DropdownOption
}

func (d *dropdownElement) describe() ElementDescriptor {
	return ElementDescriptor{
		Kind:     ElementDropdown,
		Label:    d.label,
		IntValue: d.value.Get(),
		Options:  d.options,
	}
}

func (d *dropdownElement) handleUpdate(property string, u UpdateValue) {
	if property == "value" && d.value.clientWritable {
		d.value.update(int(u.Float))
	}
}

func (d *dropdownElement) bindSend(pathPrefix string, fn func(UpdateNotification)) {
	d.value.bindSend(func(v int) {
		fn(UpdateNotification{
			Path:  pathPrefix + ".value",
			Value: UpdateValue{Kind: UpdateKindFloat, Float: float64(v)},
		})
	})
}

func (d *dropdownElement) applyForm(f *CustomForm) { f.elements = append(f.elements, d) }

// Dropdown adds a dropdown selection element bound to value.
func Dropdown(label string, value *Observable[int], options []DropdownOption) FormOption {
	return &dropdownElement{label: label, value: value, options: options}
}

type toggleElement struct {
	label string
	value *Observable[bool]
}

func (t *toggleElement) describe() ElementDescriptor {
	return ElementDescriptor{
		Kind:      ElementToggle,
		Label:     t.label,
		BoolValue: t.value.Get(),
	}
}

func (t *toggleElement) handleUpdate(property string, u UpdateValue) {
	if property == "toggled" && t.value.clientWritable {
		t.value.update(u.Bool)
	}
}

func (t *toggleElement) bindSend(pathPrefix string, fn func(UpdateNotification)) {
	t.value.bindSend(func(v bool) {
		fn(UpdateNotification{
			Path:  pathPrefix + ".toggled",
			Value: UpdateValue{Kind: UpdateKindBool, Bool: v},
		})
	})
}

func (t *toggleElement) applyForm(f *CustomForm) { f.elements = append(f.elements, t) }

// Toggle adds a boolean toggle element bound to value.
func Toggle(label string, value *Observable[bool]) FormOption {
	return &toggleElement{label: label, value: value}
}

type sliderElement struct {
	label, description string
	value              *Observable[float64]
	min, max, step     float64
}

func (s *sliderElement) describe() ElementDescriptor {
	return ElementDescriptor{
		Kind:        ElementSlider,
		Label:       s.label,
		Description: s.description,
		FloatValue:  s.value.Get(),
		Min:         s.min,
		Max:         s.max,
		Step:        s.step,
	}
}

func (s *sliderElement) handleUpdate(property string, u UpdateValue) {
	if property == "value" && s.value.clientWritable {
		s.value.update(u.Float)
	}
}

func (s *sliderElement) bindSend(pathPrefix string, fn func(UpdateNotification)) {
	s.value.bindSend(func(v float64) {
		fn(UpdateNotification{
			Path:  pathPrefix + ".value",
			Value: UpdateValue{Kind: UpdateKindFloat, Float: v},
		})
	})
}

func (s *sliderElement) applyForm(f *CustomForm) { f.elements = append(f.elements, s) }

// Slider adds a numeric slider element bound to value. min and max define the range.
func Slider(label string, value *Observable[float64], min, max float64, opts ...SliderOption) FormOption {
	e := &sliderElement{label: label, value: value, min: min, max: max, step: 1}
	for _, o := range opts {
		if o.description != "" {
			e.description = o.description
		}
		if o.step != 0 {
			e.step = o.step
		}
	}
	return e
}

type buttonElement struct {
	label   string
	onClick func()
}

func (b *buttonElement) describe() ElementDescriptor {
	return ElementDescriptor{Kind: ElementButton, Label: b.label}
}

func (b *buttonElement) handleUpdate(property string, _ UpdateValue) {
	if property == "onClick" && b.onClick != nil {
		b.onClick()
	}
}

func (b *buttonElement) bindSend(_ string, _ func(UpdateNotification)) {}

func (b *buttonElement) applyForm(f *CustomForm) { f.elements = append(f.elements, b) }

// Button adds a clickable button element. onClick is called when the client clicks it.
func Button(label string, onClick func()) FormOption {
	return &buttonElement{label: label, onClick: onClick}
}

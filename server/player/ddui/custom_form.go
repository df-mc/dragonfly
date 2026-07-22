package ddui

import "strconv"

// FormOption configures a CustomForm. The interface is sealed; use the
// provided constructors (CloseButton, Spacer, Label, etc.) to build options.
type FormOption interface {
	applyForm(f *CustomForm)
}

// closeButtonOption enables the close button on a CustomForm.
type closeButtonOption struct{}

func (closeButtonOption) applyForm(f *CustomForm) { f.hasClose = true }

// CloseButton adds a visible close button to the top-right corner of the form.
func CloseButton() FormOption { return closeButtonOption{} }

// CustomForm is a fully customisable data-driven UI form. It supports real-time
// value updates in both directions via Observables while the form is open.
type CustomForm struct {
	title        string
	hasClose     bool
	elements     []element
	closeHandler func(reason int)
}

// New creates a CustomForm with the given title and options.
func New(title string, opts ...FormOption) *CustomForm {
	f := &CustomForm{title: title}
	for _, o := range opts {
		o.applyForm(f)
	}
	return f
}

// ScreenID ...
func (f *CustomForm) ScreenID() string { return "minecraft:custom_form" }

// Describe ...
func (f *CustomForm) Describe() FormDescriptor {
	descs := make([]ElementDescriptor, len(f.elements))
	for i, e := range f.elements {
		descs[i] = e.describe()
	}
	return FormDescriptor{
		Title:          f.title,
		HasCloseButton: f.hasClose,
		Elements:       descs,
	}
}

// HandleUpdate ...
func (f *CustomForm) HandleUpdate(path string, value UpdateValue) bool {
	idx, property, ok := parseLayoutPath(path)
	if !ok || idx < 0 || idx >= len(f.elements) {
		return false
	}
	f.elements[idx].handleUpdate(property, value)
	return false
}

// BindSend ...
func (f *CustomForm) BindSend(fn func(UpdateNotification)) {
	for i, e := range f.elements {
		e.bindSend("layout["+strconv.Itoa(i)+"]", fn)
	}
}

// OnClose ...
func (f *CustomForm) OnClose(reason int) {
	if f.closeHandler != nil {
		f.closeHandler(reason)
	}
}

// parseLayoutPath parses a path of the form "layout[N].property".
func parseLayoutPath(path string) (idx int, property string, ok bool) {
	const prefix = "layout["
	if len(path) <= len(prefix) || path[:len(prefix)] != prefix {
		return 0, "", false
	}
	rest := path[len(prefix):]
	bracket := -1
	for i, c := range rest {
		if c == ']' {
			bracket = i
			break
		}
	}
	if bracket < 0 || bracket+2 >= len(rest) || rest[bracket+1] != '.' {
		return 0, "", false
	}
	n, err := strconv.Atoi(rest[:bracket])
	if err != nil {
		return 0, "", false
	}
	return n, rest[bracket+2:], true
}

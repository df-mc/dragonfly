package ddui

// MessageBoxOption configures a MessageBox. The interface is sealed; use the
// provided constructors (Body, Button1, Button2, Handler) to build options.
type MessageBoxOption interface {
	applyMessageBox(m *MessageBox)
}

// bodyOption sets the body text of a MessageBox.
type bodyOption struct{ text string }

func (b bodyOption) applyMessageBox(m *MessageBox) { m.body = b.text }

// Body sets the body text displayed in the message box.
func Body(text string) MessageBoxOption { return bodyOption{text: text} }

// button1Option sets the first button of a MessageBox.
type button1Option struct {
	label string
	opts  []MessageBoxButtonOption
}

func (b button1Option) applyMessageBox(m *MessageBox) {
	m.btn1.label = b.label
	for _, o := range b.opts {
		m.btn1.tooltip = o.tooltip
	}
}

// Button1 sets the label and optional tooltip for the first button.
func Button1(label string, opts ...MessageBoxButtonOption) MessageBoxOption {
	return button1Option{label: label, opts: opts}
}

// button2Option sets the second button of a MessageBox.
type button2Option struct {
	label string
	opts  []MessageBoxButtonOption
}

func (b button2Option) applyMessageBox(m *MessageBox) {
	m.btn2.label = b.label
	for _, o := range b.opts {
		m.btn2.tooltip = o.tooltip
	}
}

// Button2 sets the label and optional tooltip for the second button.
func Button2(label string, opts ...MessageBoxButtonOption) MessageBoxOption {
	return button2Option{label: label, opts: opts}
}

// MessageBox is a two-button confirmation dialog. The Handler is called with
// the selected button (1 or 2) when the form closes, or 0 if cancelled.
type MessageBox struct {
	title, body string
	btn1, btn2  messageBoxButton
	selection   int
	handler     func(selection int)
}

type messageBoxButton struct {
	label, tooltip string
}

// MessageBoxButtonOption configures optional properties of a MessageBox button.
type MessageBoxButtonOption struct{ tooltip string }

// WithTooltip sets the tooltip text shown when hovering over a button.
func WithTooltip(tooltip string) MessageBoxButtonOption {
	return MessageBoxButtonOption{tooltip: tooltip}
}

// NewMessageBox creates a MessageBox with the given title and options.
func NewMessageBox(title string, opts ...MessageBoxOption) *MessageBox {
	m := &MessageBox{title: title}
	for _, o := range opts {
		o.applyMessageBox(m)
	}
	return m
}

// ScreenID implements Form.
func (m *MessageBox) ScreenID() string { return "minecraft:message_box" }

// Describe implements Form.
func (m *MessageBox) Describe() FormDescriptor {
	return FormDescriptor{
		Title: m.title,
		Body:  m.body,
		Button1: ButtonDescriptor{
			Label:   m.btn1.label,
			Tooltip: m.btn1.tooltip,
		},
		Button2: ButtonDescriptor{
			Label:   m.btn2.label,
			Tooltip: m.btn2.tooltip,
		},
	}
}

// HandleUpdate implements Form. Returns true when a button is clicked, signalling
// the session to close the form and call OnClose immediately.
func (m *MessageBox) HandleUpdate(path string, _ UpdateValue) bool {
	switch path {
	case "button1.onClick":
		m.selection = 1
		return true
	case "button2.onClick":
		m.selection = 2
		return true
	}
	return false
}

// BindSend implements Form.
func (m *MessageBox) BindSend(_ func(UpdateNotification)) {}

// OnClose implements Form.
func (m *MessageBox) OnClose(_ int) {
	if m.handler != nil {
		m.handler(m.selection)
	}
}

package ddui

// MessageBoxOption configures a MessageBox. The interface is sealed; use the
// provided constructors (Body, Button1, Button2, Handler).
type MessageBoxOption interface {
	applyMessageBox(m *MessageBox)
}

type bodyOption struct{ obs *Observable[string] }

func (b bodyOption) applyMessageBox(m *MessageBox) { m.body = b.obs }

// Body sets the body text of the message box. Accepts a plain string or an
// *Observable[string]; passing an observable allows live updates while the form is open.
func Body[T string | *Observable[string]](text T) MessageBoxOption {
	return bodyOption{obs: toStringObs(text)}
}

type button1Option struct {
	label *Observable[string]
	opts  []MessageBoxButtonOption
}

func (b button1Option) applyMessageBox(m *MessageBox) {
	m.btn1.label = b.label
	for _, o := range b.opts {
		if o.tooltip != nil {
			m.btn1.tooltip = o.tooltip
		}
	}
}

// Button1 sets the first button's label and optional tooltip. Accepts a plain
// string or an *Observable[string] for the label.
func Button1[T string | *Observable[string]](label T, opts ...MessageBoxButtonOption) MessageBoxOption {
	return button1Option{label: toStringObs(label), opts: opts}
}

type button2Option struct {
	label *Observable[string]
	opts  []MessageBoxButtonOption
}

func (b button2Option) applyMessageBox(m *MessageBox) {
	m.btn2.label = b.label
	for _, o := range b.opts {
		if o.tooltip != nil {
			m.btn2.tooltip = o.tooltip
		}
	}
}

// Button2 sets the second button's label and optional tooltip. Accepts a plain
// string or an *Observable[string] for the label.
func Button2[T string | *Observable[string]](label T, opts ...MessageBoxButtonOption) MessageBoxOption {
	return button2Option{label: toStringObs(label), opts: opts}
}

// MessageBox is a two-button confirmation dialog. The Handler is called with
// the selected button (1 or 2) when the form closes, or 0 if cancelled.
type MessageBox struct {
	title      *Observable[string]
	body       *Observable[string]
	btn1, btn2 messageBoxButton
	selection  int
	handler    func(selection int)
}

type messageBoxButton struct {
	label   *Observable[string]
	tooltip *Observable[string]
}

// MessageBoxButtonOption configures optional properties of a MessageBox button.
type MessageBoxButtonOption struct{ tooltip *Observable[string] }

// WithTooltip sets the button tooltip. Accepts a plain string or an *Observable[string].
func WithTooltip[T string | *Observable[string]](tooltip T) MessageBoxButtonOption {
	return MessageBoxButtonOption{tooltip: toStringObs(tooltip)}
}

// NewMessageBox creates a MessageBox with the given title and options. title
// accepts a plain string or an *Observable[string].
func NewMessageBox[T string | *Observable[string]](title T, opts ...MessageBoxOption) *MessageBox {
	m := &MessageBox{title: toStringObs(title)}
	for _, o := range opts {
		o.applyMessageBox(m)
	}
	return m
}

// ScreenID implements Form.
func (m *MessageBox) ScreenID() string { return "minecraft:message_box" }

// Describe implements Form.
func (m *MessageBox) Describe() FormDescriptor {
	desc := FormDescriptor{Title: m.title.Get()}
	if m.body != nil {
		desc.Body = m.body.Get()
	}
	if m.btn1.label != nil {
		desc.Button1.Label = m.btn1.label.Get()
	}
	if m.btn1.tooltip != nil {
		desc.Button1.Tooltip = m.btn1.tooltip.Get()
	}
	if m.btn2.label != nil {
		desc.Button2.Label = m.btn2.label.Get()
	}
	if m.btn2.tooltip != nil {
		desc.Button2.Tooltip = m.btn2.tooltip.Get()
	}
	return desc
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
func (m *MessageBox) BindSend(fn func(UpdateNotification)) {
	bind := func(obs *Observable[string], path string) {
		if obs == nil {
			return
		}
		obs.bindSend(func(v string) {
			fn(UpdateNotification{Path: path, Value: UpdateValue{Kind: UpdateKindString, String: v}})
		})
	}
	bind(m.title, "title")
	bind(m.body, "body")
	bind(m.btn1.label, "button1.label")
	bind(m.btn1.tooltip, "button1.tooltip")
	bind(m.btn2.label, "button2.label")
	bind(m.btn2.tooltip, "button2.tooltip")
}

// OnClose implements Form.
func (m *MessageBox) OnClose(_ int) {
	if m.handler != nil {
		m.handler(m.selection)
	}
}

// toStringObs converts a string or *Observable[string] to *Observable[string].
func toStringObs[T string | *Observable[string]](v T) *Observable[string] {
	switch val := any(v).(type) {
	case string:
		return NewObservable(val, false)
	case *Observable[string]:
		return val
	}
	panic("unreachable")
}

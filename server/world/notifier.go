package world

// Notifier ...
type Notifier interface {
	Notify(message any, stack []byte)
}

// NopNotifier ...
type NopNotifier struct{}

// Notify ...
func (NopNotifier) Notify(any, []byte) {}

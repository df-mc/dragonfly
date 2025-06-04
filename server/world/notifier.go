package world

// Notifier receives notifications of panics or errors with a message and stack trace.
type Notifier interface {
	Notify(message any, stack []byte)
}

// NopNotifier is a no-operation implementation of Notifier that ignores all notifications.
type NopNotifier struct{}

// Notify ...
func (NopNotifier) Notify(any, []byte) {}

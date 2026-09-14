package ddui

import "sync"

// Observable holds a value that can be observed for changes. When Set is called,
// all registered listeners and the bound send function are notified.
type Observable[T string | int | float64 | bool] struct {
	listeners      []func(value T)
	sendFn         func(value T)
	clientWritable bool
	value          T
	mut            sync.RWMutex
}

// NewObservable creates a new Observable with the given initial value.
// clientWritable controls whether client-originated packets may write back into
// this observable. Set it to false for server-authoritative values.
func NewObservable[T string | int | float64 | bool](initialValue T, clientWritable bool) *Observable[T] {
	return &Observable[T]{
		listeners:      make([]func(value T), 0),
		clientWritable: clientWritable,
		value:          initialValue,
	}
}

// Set updates the current value and notifies all listeners and the send function.
func (o *Observable[T]) Set(value T) {
	o.mut.Lock()
	o.value = value
	listeners := o.listeners
	sendFn := o.sendFn
	o.mut.Unlock()

	for _, fn := range listeners {
		fn(value)
	}
	if sendFn != nil {
		sendFn(value)
	}
}

// Get returns the current value.
func (o *Observable[T]) Get() T {
	o.mut.RLock()
	defer o.mut.RUnlock()
	return o.value
}

// Listen adds a listener that is called when the value changes via Set.
func (o *Observable[T]) Listen(fn func(value T)) {
	o.mut.Lock()
	o.listeners = append(o.listeners, fn)
	o.mut.Unlock()
}

// update sets the value and notifies listeners without calling the send function.
// Used to apply changes without echoing back to the client.
func (o *Observable[T]) update(value T) {
	o.mut.Lock()
	o.value = value
	listeners := o.listeners
	o.mut.Unlock()

	for _, fn := range listeners {
		fn(value)
	}
}

// bindSend registers the send callback. It is called by form types after the
// session attaches a send function via BindSend.
func (o *Observable[T]) bindSend(fn func(T)) {
	o.mut.Lock()
	o.sendFn = fn
	o.mut.Unlock()
}

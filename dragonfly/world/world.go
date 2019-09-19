package world

import "sync"

// World implements a Minecraft world. It manages all aspects of what players can see, such as blocks,
// entities and particles.
type World struct {
	name string

	hMutex sync.RWMutex
	h      Handler
	pMutex sync.RWMutex
	p      Provider
}

// New creates a new initialised world. The world may be used right away, but it will not be saved or loaded
// from files until it has been given a different provider than the default. (NoIOProvider)
// By default, the name of the world will be 'World'.
func New() *World {
	return &World{name: "World", p: NoIOProvider{}}
}

// Name returns the display name of the world. Generally, this name is displayed at the top of the player list
// in the pause screen in-game.
// If a provider is set, the name will be updated according to the name that it provides.
func (w *World) Name() string {
	w.pMutex.RLock()
	defer w.pMutex.RUnlock()
	return w.name
}

// Provider changes the provider of the world to the provider passed. If nil is passed, the NoIOProvider
// will be set, which does not read or write any data.
func (w *World) Provider(p Provider) {
	w.pMutex.Lock()
	defer w.pMutex.Unlock()

	if p == nil {
		p = NoIOProvider{}
	}
	w.p = p
	w.name = p.WorldName()
}

// Handle changes the current handler of the world. As a result, events called by the world will call
// handlers of the Handler passed.
// Handle sets the world's handler to NopHandler if nil is passed.
func (w *World) Handle(h Handler) {
	w.hMutex.Lock()
	defer w.hMutex.Unlock()

	if h == nil {
		h = NopHandler{}
	}
	w.h = h
}

// provider returns the provider of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) provider() Provider {
	w.pMutex.RLock()
	provider := w.p
	w.pMutex.RUnlock()
	return provider
}

// handler returns the handler of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) handler() Handler {
	w.hMutex.RLock()
	handler := w.h
	w.hMutex.RUnlock()
	return handler
}

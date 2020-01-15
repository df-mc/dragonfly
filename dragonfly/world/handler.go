package world

// Handler handles events that are called by a world. Implementations of Handler may be used to listen to
// specific events such as when an entity is added to the world.
type Handler interface {
}

// NopHandler implements the Handler interface but does not execute any code when an event is called. The
// default handler of worlds is set to NopHandler.
// Users may embed NopHandler to avoid having to implement each method.
type NopHandler struct{}

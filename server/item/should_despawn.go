package item

// ShouldDespawn represents an item that has a configurable despawn behaviour while floating in the world.
type ShouldDespawn interface {
	// ShouldDespawn returns whether the item should eventually despawn while floating in the world.
	ShouldDespawn() bool
}

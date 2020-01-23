package state

// State represents a part of the state of an entity. Entities may hold a combination of these to indicate
// things such as whether it is sprinting or on fire.
type State interface {
	__()
}

// Sneaking makes the entity show up as if it is sneaking.
type Sneaking struct{}

// Breathing makes an entity breath: This state will not show up for entities other than players.
type Breathing struct{}

func (Sneaking) __()  {}
func (Breathing) __() {}

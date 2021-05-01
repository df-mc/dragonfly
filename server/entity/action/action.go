package action

// Action represents an action that may be performed by an entity. Typically, these actions are sent to
// viewers in a world so that they can see these actions.
type Action interface {
	__()
}

// SwingArm makes an entity or player swing its arm.
type SwingArm struct{ action }

// Hurt makes an entity display the animation for being hurt. The entity will be shown as red for a short
// duration.
type Hurt struct{ action }

// Death makes an entity display the death animation. After this animation, the entity disappears from viewers
// watching it.
type Death struct{ action }

// Eat makes an entity display the eating particles at its mouth to viewers with the item in its hand being
// eaten.
type Eat struct{ action }

// PickedUp makes an item get picked up by a collector. After this animation, the item disappears from viewers
// watching it.
type PickedUp struct {
	// Collector is the entity that collected the item.
	Collector interface{}

	action
}

// action implements the Action interface. Structures in this package may embed it to gets its functionality
// out of the box.
type action struct{}

func (action) __() {}

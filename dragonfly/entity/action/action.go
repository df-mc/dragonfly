package action

// Action represents an action that may be performed by an entity. Typically, these actions are sent to
// viewers in a world so that they can see these actions.
type Action interface {
	__()
}

// SwingArm makes an entity or player swing its arm.
type SwingArm struct{ action }

// action implements the Action interface. Structures in this package may embed it to gets its functionality
// out of the box.
type action struct{}

func (action) __() {}

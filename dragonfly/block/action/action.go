package action

// Action represents an action that may be performed by a block. Typically, these actions are sent to
// viewers in a world so that they can see these actions.
type Action interface {
	__()
}

// Open is an action to open a block at a position. It is sent for blocks such as chests.
type Open struct{ action }

// Close is an action to close a block at a position, complementary to the Open action.
type Close struct{ action }

// action implements the Action interface. Structures in this package may embed it to gets its functionality
// out of the box.
type action struct{}

func (action) __() {}

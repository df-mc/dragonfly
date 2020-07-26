package action

import "time"

// Action represents an action that may be performed by a block. Typically, these actions are sent to
// viewers in a world so that they can see these actions.
type Action interface {
	__()
}

// Open is an action to open a block at a position. It is sent for blocks such as chests.
type Open struct{ action }

// Close is an action to close a block at a position, complementary to the Open action.
type Close struct{ action }

// StartCrack is an action to make the cracks in a block start forming, following the break time set in the
// action.
type StartCrack struct {
	action
	BreakTime time.Duration
}

// ContinueCrack is an action sent every so often to continue the cracking process of the block. It is only
// ever sent after a StartCrack action, and may have an altered break time if the player is not on the ground,
// submerged or is using a different item than at first.
type ContinueCrack struct {
	action
	BreakTime time.Duration
}

// StopCrack is an action to make the cracks forming in a block stop and disappear.
type StopCrack struct{ action }

// action implements the Action interface. Structures in this package may embed it to gets its functionality
// out of the box.
type action struct{}

func (action) __() {}

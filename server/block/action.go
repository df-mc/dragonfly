package block

import "time"

// OpenAction is a world.BlockAction to open a block at a position. It is sent for blocks such as chests.
type OpenAction struct{ action }

// CloseAction is a world.BlockAction to close a block at a position, complementary to the OpenAction action.
type CloseAction struct{ action }

// StartCrackAction is a world.BlockAction to make the cracks in a block start forming, following the break time set in
// the action.
type StartCrackAction struct {
	action
	BreakTime time.Duration
}

// ContinueCrackAction is a world.BlockAction sent every so often to continue the cracking process of the block. It is
// only ever sent after a StartCrackAction action, and may have an altered break time if the player is not on the
// ground, submerged or is using a different item than at first.
type ContinueCrackAction struct {
	action
	BreakTime time.Duration
}

// StopCrackAction is a world.BlockAction to make the cracks forming in a block stop and disappear.
type StopCrackAction struct{ action }

// action implements the Action interface. Structures in this package may embed it to gets its functionality
// out of the box.
type action struct{}

// BlockAction serves to implement the world.BlockAction interface.
func (action) BlockAction() {}

package dialogue

import "github.com/df-mc/dragonfly/server/world"

// Submittable is a structure which may be submitted by sending it as a dialogue
// using dialogue.New(). The struct will have its Submit method called with the
// button pressed. A struct that implements the Submittable interface must only
// have exported fields with the type dialogue.Button.
type Submittable interface {
	// Submit is called when the Submitter submits the dialogue sent to it. The
	// method is called with the button that was pressed. It may be compared
	// with buttons in the Submittable struct to check which button was pressed.
	// Additionally, the world.Tx of the Submitter is passed.
	Submit(submitter Submitter, pressed Button, tx *world.Tx)
}

// Submitter is an entity that is able to submit a dialogue sent to it. It is
// able to interact with the buttons in the dialogue. The Submitter is also
// able to close the dialogue.
type Submitter interface {
	SendDialogue(d Dialogue, e world.Entity)
	CloseDialogue()
}

// Closer represents a dialogue which has special logic when being closed by a
// Submitter.
type Closer interface {
	// Close is called when the Submitter closes a dialogue.
	Close(submitter Submitter, tx *world.Tx)
}

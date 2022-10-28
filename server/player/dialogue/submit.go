package dialogue

// Submitter is an entity that is able to open a dialog and submit it by clicking a button. The entity can also have
// a dialogue force closed upon it by the server.
type Submitter interface {
	SendDialogue(d Dialogue)
	CloseDialogue(d Dialogue)
}

// Closer represents a scene that has special logic when the dialogue scene is closed.
type Closer interface {
	// Close gets called when a Submitter closes a dialogue.
	Close(submitter Submitter)
}
